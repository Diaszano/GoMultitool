// Package installer creates read-only plans before image writes.
package installer

import (
	"context"
	"errors"
	"fmt"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

// ErrIdentityMismatch indicates the selected device changed after selection.
var ErrIdentityMismatch = errors.New("selected device identity changed")

// Planner resolves and validates installation targets without opening them.
type Planner struct{ Catalog application.DeviceCatalog }

// Plan creates a read-only operation summary for an image and selected device.
func (p Planner) Plan(ctx context.Context, img image.Image, expected device.Device) (installation.Plan, error) {
	current, err := p.Catalog.Resolve(ctx, expected.Path)
	if err != nil {
		return installation.Plan{}, err
	}
	if current != expected {
		return installation.Plan{}, ErrIdentityMismatch
	}
	plan := installation.Plan{Image: img, Target: current, Summary: fmt.Sprintf("Write %s (%d bytes) to %s (%d bytes)", img.Path, img.SizeBytes, current.Path, current.SizeBytes)}
	if err := plan.Validate(); err != nil {
		return installation.Plan{}, err
	}
	return plan, nil
}
