// Package discovery provides storage inventory use cases.
package discovery

import (
	"context"
	"sort"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// Service exposes a deterministic storage inventory.
type Service struct{ Catalog application.DeviceCatalog }

// Inventory returns devices ordered by stable identity.
func (s Service) Inventory(ctx context.Context) ([]device.Device, error) {
	devices, err := s.Catalog.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	return devices, nil
}
