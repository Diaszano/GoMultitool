package device

import (
	"errors"
	"testing"
)

func TestDeviceValidateWriteTarget(t *testing.T) {
	base := Device{
		ID:                "mmc0",
		Path:              "/dev/mmc0",
		Kind:              KindEMMC,
		SizeBytes:         1 << 20,
		LogicalSectorSize: 512,
	}
	for _, tc := range []struct {
		name string
		edit func(*Device)
	}{
		{"mounted", func(d *Device) { d.Mounted = true }},
		{"boot media", func(d *Device) { d.BootMedia = true }},
		{"unknown", func(d *Device) { d.Kind = KindUnknown }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := base
			tc.edit(&d)
			if err := d.ValidateWriteTarget(); !errors.Is(err, ErrUnsafeDevice) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
