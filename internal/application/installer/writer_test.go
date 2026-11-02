package installer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

type writerCatalog struct{ d device.Device }

func (c writerCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.d}, nil
}
func (c writerCatalog) Resolve(_ context.Context, p string) (device.Device, error) {
	if p != c.d.Path {
		return device.Device{}, os.ErrNotExist
	}
	return c.d, nil
}

type shortTarget struct{ bytes.Buffer }

func (shortTarget) Sync() error                    { return nil }
func (shortTarget) Close() error                   { return nil }
func (shortTarget) Read([]byte) (int, error)       { return 0, io.EOF }
func (shortTarget) Seek(int64, int) (int64, error) { return 0, nil }
func (shortTarget) Write(p []byte) (int, error)    { return len(p) - 1, nil }

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("decompression failed") }

type cancelReader struct {
	cancel context.CancelFunc
	sent   bool
}

func (r *cancelReader) Read(p []byte) (int, error) {
	if r.sent {
		return 0, io.EOF
	}
	r.sent = true
	n := copy(p, bytes.Repeat([]byte("x"), 512))
	r.cancel()
	return n, nil
}

type syncErrorTarget struct{ Target }

func (syncErrorTarget) Sync() error { return errors.New("sync failed") }

func TestInstallerWritesAndVerifies(t *testing.T) {
	data := bytes.Repeat([]byte("i"), 512)
	path := t.TempDir() + "/target"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(512); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	d := device.Device{ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 512, LogicalSectorSize: 512}
	g := storageguard.Guard{Catalog: writerCatalog{d}}
	gp, err := g.Plan(context.Background(), d.Path)
	if err != nil {
		t.Fatal(err)
	}
	plan := installation.Plan{Image: image.Image{Path: "image.img", Format: image.FormatRaw, SizeBytes: 512}, Target: d, Summary: "write"}
	err = (Writer{}).Execute(context.Background(), plan, bytes.NewReader(data), gp, true, g, func(string) (Target, error) { return os.OpenFile(path, os.O_RDWR, 0) })
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("readback mismatch")
	}
}

func TestInstallerRejectsUnconfirmedAndShortSource(t *testing.T) {
	d := device.Device{ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 512, LogicalSectorSize: 512}
	g := storageguard.Guard{Catalog: writerCatalog{d}}
	gp, _ := g.Plan(context.Background(), d.Path)
	plan := installation.Plan{Image: image.Image{Path: "image.img", Format: image.FormatRaw, SizeBytes: 512}, Target: d, Summary: "write"}
	if err := (Writer{}).Execute(context.Background(), plan, bytes.NewReader(nil), gp, false, g, nil); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("error=%v", err)
	}
	err := (Writer{}).Execute(context.Background(), plan, bytes.NewReader([]byte("short")), gp, true, g, func(string) (Target, error) { return newFileTarget(t), nil })
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error=%v", err)
	}
}

func TestInstallerRejectsDecompressionAndShortWriteErrors(t *testing.T) {
	d := device.Device{ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 512, LogicalSectorSize: 512}
	g := storageguard.Guard{Catalog: writerCatalog{d}}
	gp, _ := g.Plan(context.Background(), d.Path)
	plan := installation.Plan{Image: image.Image{Path: "image.img", Format: image.FormatRaw, SizeBytes: 512}, Target: d, Summary: "write"}
	if err := (Writer{}).Execute(context.Background(), plan, errorReader{}, gp, true, g, func(string) (Target, error) { return newFileTarget(t), nil }); err == nil {
		t.Fatal("decompression error ignored")
	}
	err := (Writer{}).Execute(context.Background(), plan, bytes.NewReader(bytes.Repeat([]byte("x"), 512)), gp, true, g, func(string) (Target, error) { return shortTarget{}, nil })
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("error=%v", err)
	}
}

func TestInstallerRejectsCancellationAndSyncFailure(t *testing.T) {
	d := device.Device{ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 512, LogicalSectorSize: 512}
	g := storageguard.Guard{Catalog: writerCatalog{d}}
	gp, _ := g.Plan(context.Background(), d.Path)
	plan := installation.Plan{Image: image.Image{Path: "image.img", Format: image.FormatRaw, SizeBytes: 512}, Target: d, Summary: "write"}
	ctx, cancel := context.WithCancel(context.Background())
	err := (Writer{}).Execute(ctx, plan, &cancelReader{cancel: cancel}, gp, true, g, func(string) (Target, error) { return newFileTarget(t), nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	err = (Writer{}).Execute(context.Background(), plan, bytes.NewReader(bytes.Repeat([]byte("x"), 512)), gp, true, g, func(string) (Target, error) { return syncErrorTarget{Target: newFileTarget(t)}, nil })
	if err == nil || err.Error() != "sync failed" {
		t.Fatalf("error=%v", err)
	}
}

func newFileTarget(t *testing.T) Target {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "target")
	if err != nil {
		t.Fatal(err)
	}
	return f
}
