// Package storageguard validates destructive storage targets before use.
package storageguard

import (
	"context"
	"errors"
	"fmt"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// ErrDeviceChanged indicates the selected device no longer matches the reviewed plan.
var ErrDeviceChanged = errors.New("device identity changed")

// Guard creates safe write plans and checks them again immediately before use.
type Guard struct{ Catalog application.DeviceCatalog }

// Plan records the device identity approved for a write operation.
type Plan struct{ Device device.Device }

// Plan resolves and validates the requested write target.
func (g Guard) Plan(ctx context.Context, path string) (Plan, error) {
	d, err := g.Catalog.Resolve(ctx, path)
	if err != nil {
		return Plan{}, err
	}
	if err := d.ValidateWriteTarget(); err != nil {
		return Plan{}, fmt.Errorf("%w: %s", device.ErrUnsafeDevice, path)
	}
	return Plan{Device: d}, nil
}

// Revalidate ensures the planned device is still present and unchanged.
func (g Guard) Revalidate(ctx context.Context, plan Plan) error {
	current, err := g.Catalog.Resolve(ctx, plan.Device.Path)
	if err != nil {
		return err
	}
	if current != plan.Device {
		return ErrDeviceChanged
	}
	return current.ValidateWriteTarget()
}
