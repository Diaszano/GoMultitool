// Package application defines interfaces shared by use cases and adapters.
package application

import (
	"context"
	"io"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// DeviceCatalog lists storage devices and resolves their current identity.
type DeviceCatalog interface {
	List(context.Context) ([]device.Device, error)
	Resolve(context.Context, string) (device.Device, error)
}

// Target is a block device opened for synchronized writes and readback.
type Target interface {
	io.ReadWriteSeeker
	Sync() error
	Close() error
}
