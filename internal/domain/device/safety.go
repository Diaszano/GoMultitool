package device

import "errors"

// ErrUnsafeDevice indicates a device cannot be safely written.
var ErrUnsafeDevice = errors.New("unsafe write target")

// ValidateWriteTarget rejects devices whose identity or current use is unsafe.
func (d Device) ValidateWriteTarget() error {
	if d.Validate() != nil || d.Kind == KindUnknown || d.Mounted || d.BootMedia {
		return ErrUnsafeDevice
	}
	return nil
}
