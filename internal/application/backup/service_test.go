package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
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
	)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatalf("data=%q", data)
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBackupRejectsSpaceAndReadFailuresWithoutManifest(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  io.Reader
	}{
		{"read", failingReader{}},
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
	)
	if err == nil {
		t.Fatal("canceled backup succeeded")
	}
	if _, err := os.Stat(dest + ".json"); !os.IsNotExist(err) {
		t.Fatalf("manifest exists: %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("partial data exists: %v", err)
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
	)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}

var testDevice = device.Device{
	ID: "mmc0", Path: "/dev/mmc0", Kind: device.KindEMMC, SizeBytes: 4096, LogicalSectorSize: 512,
}

func TestCreateWritesGzipReadableByGunzip(t *testing.T) {
	raw := bytes.Repeat([]byte("abcd"), 1024)
	dest := filepath.Join(t.TempDir(), "backup.img.gz")
	m, err := (Service{}).Create(context.Background(), testDevice, bytes.NewReader(raw), dest)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(zr)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatalf("gunzip mismatch: %v", err)
	}
	sum := sha256.Sum256(raw)
	if m.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("sha=%s", m.SHA256)
	}
}

type failAfter struct {
	w io.Writer
	n int
}

func (f *failAfter) Write(p []byte) (int, error) {
	if f.n <= 0 {
		return 0, &os.PathError{Op: "write", Path: "x", Err: syscall.ENOSPC}
	}
	if len(p) > f.n {
		p = p[:f.n]
	}
	n, _ := f.w.Write(p)
	f.n -= n
	return n, nil
}

func TestCreateNoSpaceRemovesPartialFiles(t *testing.T) {
	old := dataSink
	dataSink = func(f *os.File) io.Writer { return &failAfter{w: f, n: 1024} }
	defer func() { dataSink = old }()
	// Incompressible data so gzip flushes more than 1024 bytes.
	raw := make([]byte, 4096)
	for i := range raw {
		raw[i] = byte(i*131 + i*i>>3)
	}
	dir := t.TempDir()
	dest := filepath.Join(dir, "backup.img.gz")
	_, err := (Service{}).Create(context.Background(), testDevice, bytes.NewReader(raw), dest)
	if !errors.Is(err, ErrNoSpace) {
		t.Fatalf("error = %v", err)
	}
	left, _ := os.ReadDir(dir)
	if len(left) != 0 {
		t.Fatalf("leftovers: %v", left)
	}
}

func TestBackupExistingOutputIsErrExists(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "backup.img")
	if err := os.WriteFile(dest, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	src := device.Device{
		ID:                "x",
		Path:              "/dev/x",
		Kind:              device.KindEMMC,
		SizeBytes:         512,
		LogicalSectorSize: 512,
	}
	_, err := Service{}.Create(context.Background(), src, bytes.NewReader([]byte("data")), dest)
	if !errors.Is(err, ErrExists) {
		t.Fatalf("err = %v, want ErrExists", err)
	}
}
