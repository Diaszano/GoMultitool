package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/infrastructure/multitool"
)

func TestCountingReaderReportsProgress(t *testing.T) {
	var calls [][2]int64
	r := &countingReader{
		r:     strings.NewReader("abcdef"),
		total: 6,
		p:     func(done, total int64) { calls = append(calls, [2]int64{done, total}) },
	}
	buf := make([]byte, 4)
	for {
		if _, err := r.Read(buf); errors.Is(err, io.EOF) {
			break
		}
	}
	if len(calls) != 2 || calls[0] != [2]int64{4, 6} || calls[1] != [2]int64{6, 6} {
		t.Fatalf("progress calls = %v, want [[4 6] [6 6]]", calls)
	}
}

func TestCountingReaderThrottlesProgress(t *testing.T) {
	const size = 10 << 20
	var calls int
	var last [2]int64
	r := &countingReader{
		r:     bytes.NewReader(make([]byte, size)),
		total: size,
		p: func(done, total int64) {
			calls++
			last = [2]int64{done, total}
		},
	}
	if _, err := io.CopyBuffer(struct{ io.Writer }{io.Discard}, struct{ io.Reader }{r},
		make([]byte, 4096)); err != nil {
		t.Fatal(err)
	}
	if calls > 110 || last != [2]int64{size, size} {
		t.Fatalf("calls = %d (want <= 110), last = %v (want done == total)", calls, last)
	}
}

func TestBackupDestinationRejectsPathNames(t *testing.T) {
	for _, name := range []string{"", "a/b", ".x", ".."} {
		if _, err := backupDestination("/mt/backups", name); err == nil {
			t.Errorf("backupDestination(%q) accepted an unsafe name", name)
		}
	}
	got, err := backupDestination("/mt/backups", "tv box")
	if err != nil || got != "/mt/backups/tv box.img.gz" {
		t.Fatalf("backupDestination = %q, %v", got, err)
	}
}

type fixedCatalog struct{ d device.Device }

func (c fixedCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.d}, nil
}

func (c fixedCatalog) Resolve(context.Context, string) (device.Device, error) { return c.d, nil }

type failingClose struct{ io.Reader }

var errDecompress = errors.New("xz: compressed data is corrupt")

func (failingClose) Close() error { return errDecompress }

func TestBurnReportsSourceCloseError(t *testing.T) {
	data := bytes.Repeat([]byte{7}, 4096)
	path := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(path, make([]byte, 8192), 0o600); err != nil {
		t.Fatal(err)
	}
	original := openImage
	t.Cleanup(func() { openImage = original })
	openImage = func(context.Context, string) (io.ReadCloser, error) {
		return failingClose{bytes.NewReader(data)}, nil
	}
	target := device.Device{
		ID: "usb", Path: path, Kind: device.KindUSB, SizeBytes: 8192, LogicalSectorSize: 512,
	}
	img := image.Image{Path: "a.img.xz", Format: image.FormatXZ, SizeBytes: int64(len(data))}
	_, err := burn(
		context.Background(), fixedCatalog{target}, multitool.Workspace{}, img, target, nil)
	if !errors.Is(err, errDecompress) {
		t.Fatalf("burn error = %v, want the source close error", err)
	}
}

func TestBurnSignalsSizingAndVerifyingPhases(t *testing.T) {
	var calls [][2]int64
	record := func(done, total int64) { calls = append(calls, [2]int64{done, total}) }
	missing := image.Image{
		Path:   filepath.Join(t.TempDir(), "none.img.gz"),
		Format: image.FormatGzip,
	}
	_, err := burn(context.Background(), fixedCatalog{}, multitool.Workspace{}, missing,
		device.Device{}, record)
	if err == nil || len(calls) != 1 || calls[0] != [2]int64{-1, 0} {
		t.Fatalf("sizing calls = %v, err = %v; want [[-1 0]] then the inspect error", calls, err)
	}

	calls = nil
	data := bytes.Repeat([]byte{7}, 4096)
	path := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(path, make([]byte, 8192), 0o600); err != nil {
		t.Fatal(err)
	}
	original := openImage
	t.Cleanup(func() { openImage = original })
	openImage = func(context.Context, string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	target := device.Device{
		ID: "usb", Path: path, Kind: device.KindUSB, SizeBytes: 8192, LogicalSectorSize: 512,
	}
	img := image.Image{Path: "a.img", Format: image.FormatRaw, SizeBytes: int64(len(data))}
	if _, err := burn(
		context.Background(), fixedCatalog{target}, multitool.Workspace{}, img, target, record,
	); err != nil {
		t.Fatal(err)
	}
	n := len(calls)
	if n < 2 || calls[n-2] != [2]int64{4096, 4096} || calls[n-1] != [2]int64{-2, 0} {
		t.Fatalf("burn calls = %v, want [4096 4096] then [-2 0] at the end", calls)
	}
}
