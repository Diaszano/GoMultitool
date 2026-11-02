package integration

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application/installer"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

type installCatalog struct{ d device.Device }

func (c installCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.d}, nil
}
func (c installCatalog) Resolve(_ context.Context, p string) (device.Device, error) {
	if p != c.d.Path {
		return device.Device{}, os.ErrNotExist
	}
	return c.d, nil
}

func TestInstallTemporaryFileReadback(t *testing.T) {
	data := bytes.Repeat([]byte("i"), 512)
	path := t.TempDir() + "/target"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(512); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	d := device.Device{ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 512, LogicalSectorSize: 512}
	g := storageguard.Guard{Catalog: installCatalog{d}}
	gp, err := g.Plan(context.Background(), d.Path)
	if err != nil {
		t.Fatal(err)
	}
	plan := installation.Plan{Image: image.Image{Path: "fixture.img", Format: image.FormatRaw, SizeBytes: 512}, Target: d, Summary: "Write fixture.img to /dev/sdb"}
	err = (installer.Writer{}).Execute(context.Background(), plan, bytes.NewReader(data), gp, true, g, func(string) (installer.Target, error) { return os.OpenFile(path, os.O_RDWR, 0) })
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("installed data differs")
	}
}
