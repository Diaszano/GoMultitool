package installer

import (
	"context"
	"errors"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
)

type testCatalog struct{ current device.Device }

func (c testCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.current}, nil
}
func (c testCatalog) Resolve(_ context.Context, p string) (device.Device, error) {
	if p != c.current.Path {
		return device.Device{}, errors.New("missing")
	}
	return c.current, nil
}

func TestPlannerRejectsIdentityMismatch(t *testing.T) {
	d := device.Device{ID: "new", Path: "/dev/sdb", Kind: device.KindUSB, SizeBytes: 1024, LogicalSectorSize: 512}
	p := Planner{Catalog: testCatalog{d}}
	_, err := p.Plan(context.Background(), image.Image{Path: "os.img", Format: image.FormatRaw, SizeBytes: 512}, device.Device{ID: "old", Path: d.Path, Kind: d.Kind, SizeBytes: d.SizeBytes, LogicalSectorSize: 512})
	if !errors.Is(err, ErrIdentityMismatch) {
		t.Fatalf("error=%v", err)
	}
}
