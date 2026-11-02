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

// BlockOpener opens a block device for explicitly selected access.
type BlockOpener interface {
	OpenRead(context.Context, string) (io.ReadCloser, error)
	OpenWrite(context.Context, string) (BlockWriter, error)
}

// BlockWriter synchronizes writes before closing the device.
type BlockWriter interface {
	io.Writer
	Sync() error
	Close() error
}
