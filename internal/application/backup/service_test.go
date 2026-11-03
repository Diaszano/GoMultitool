package backup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failure") }

type cancelReader struct {
	cancel context.CancelFunc
	sent   bool
}

func (r *cancelReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		n := copy(p, []byte("data"))
		r.cancel()
		return n, nil
	}
	return 0, io.EOF
}

func TestBackupWritesDataAndManifest(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "backup.img")
	source := device.Device{
		ID:                "mmc0",
		Path:              "/dev/mmc0",
		Kind:              device.KindEMMC,
		SizeBytes:         512,
		LogicalSectorSize: 512,
	}
	result, err := (Service{}).Create(
		context.Background(),
		source,
		bytes.NewReader(bytes.Repeat([]byte("d"), 512)),
		dest,
		1024,
	)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 512 {
		t.Fatalf("data=%q", data)
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBackupRejectsSpaceAndReadFailuresWithoutManifest(t *testing.T) {
	for _, tc := range []struct {
		name      string
		src       io.Reader
		available int64
	}{
		{"space", bytes.NewReader(bytes.Repeat([]byte("d"), 512)), 1}, {"read", failingReader{}, 1024},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "backup.img")
			_, err := (Service{}).Create(
				context.Background(),
				device.Device{
					ID:                "mmc0",
					Path:              "/dev/mmc0",
					Kind:              device.KindEMMC,
					SizeBytes:         512,
					LogicalSectorSize: 512,
				},
				tc.src,
				dest,
				tc.available,
			)
			if err == nil {
				t.Fatal("Create succeeded")
			}
			if _, err := os.Stat(dest + ".json"); !os.IsNotExist(err) {
				t.Fatalf("manifest exists: %v", err)
			}
		})
	}
}

func TestBackupCancellationAndUnexpectedEOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dest := filepath.Join(t.TempDir(), "backup.img")
	_, err := (Service{}).Create(
		ctx,
		device.Device{
			ID:                "mmc0",
			Path:              "/dev/mmc0",
			Kind:              device.KindEMMC,
			SizeBytes:         512,
			LogicalSectorSize: 512,
		},
		&cancelReader{cancel: cancel},
		dest,
		1024,
	)
	if err == nil {
		t.Fatal("canceled backup succeeded")
	}
	if _, err := os.Stat(dest + ".json"); !os.IsNotExist(err) {
		t.Fatalf("manifest exists: %v", err)
	}
	_, err = (Service{}).Create(
		context.Background(),
		device.Device{
			ID:                "mmc0",
			Path:              "/dev/mmc0",
			Kind:              device.KindEMMC,
			SizeBytes:         512,
			LogicalSectorSize: 512,
		},
		bytes.NewReader([]byte("short")),
		dest,
		1024,
	)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}
