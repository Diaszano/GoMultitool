package storageguard

import (
	"context"
	"errors"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type catalogFunc func(context.Context, string) (device.Device, error)

func (f catalogFunc) Resolve(ctx context.Context, path string) (device.Device, error) {
	return f(ctx, path)
}

func TestPlanAndRevalidate(t *testing.T) {
	current := device.Device{ID: "mmc0", Path: "/dev/mmc0", Kind: device.KindEMMC, SizeBytes: 1 << 20, LogicalSectorSize: 512}
	g := Guard{Catalog: catalogFunc(func(context.Context, string) (device.Device, error) { return current, nil })}
	plan, err := g.Plan(context.Background(), current.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Revalidate(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	current.ID = "replacement"
	if err := g.Revalidate(context.Background(), plan); !errors.Is(err, ErrDeviceChanged) {
		t.Fatalf("error = %v", err)
	}
}
