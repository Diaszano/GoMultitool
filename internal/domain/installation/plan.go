// Package installation defines validated image-write plans.
package installation

import (
	"errors"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
)

// ErrImageTooLarge indicates the target cannot contain the image.
var ErrImageTooLarge = errors.New("image exceeds target capacity")

// ErrInvalidPlan indicates the operation summary or image metadata is incomplete.
var ErrInvalidPlan = errors.New("invalid installation plan")

// Plan describes a validated image and destination before any write occurs.
type Plan struct {
	Image   image.Image
	Target  device.Device
	Summary string
}

// Validate rejects unsupported images and unsafe or undersized destinations.
func (p Plan) Validate() error {
	if p.Image.Path == "" || p.Image.SizeBytes <= 0 ||
		(p.Image.Format != image.FormatRaw && p.Image.Format != image.FormatXZ) ||
		p.Summary == "" {
		return ErrInvalidPlan
	}
	if err := p.Target.ValidateWriteTarget(); err != nil {
		return err
	}
	if p.Image.SizeBytes > p.Target.SizeBytes {
		return ErrImageTooLarge
	}
	return nil
}
