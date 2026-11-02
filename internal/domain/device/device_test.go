package device

import (
	"errors"
	"testing"
)

func TestDeviceValidate(t *testing.T) {
	for _, kind := range []Kind{KindEMMC, KindSD, KindUSB} {
		t.Run(string(kind), func(t *testing.T) {
			d := Device{ID: "device-1", Path: "/dev/test", Kind: kind, SizeBytes: 1 << 30, LogicalSectorSize: 512}
			if err := d.Validate(); err != nil {
				t.Fatalf("valid device rejected: %v", err)
			}
		})
	}
}

func TestDeviceValidateRejectsInvalidMetadata(t *testing.T) {
	valid := Device{ID: "device-1", Path: "/dev/test", Kind: KindEMMC, SizeBytes: 1 << 30, LogicalSectorSize: 512}
	for _, tc := range []struct {
		name string
		edit func(*Device)
	}{
		{"missing identity", func(d *Device) { d.ID = "" }},
		{"missing path", func(d *Device) { d.Path = "" }},
		{"nonpositive size", func(d *Device) { d.SizeBytes = 0 }},
		{"unaligned size", func(d *Device) { d.SizeBytes = 513 }},
		{"unsupported sector size", func(d *Device) { d.LogicalSectorSize = 1024 }},
		{"unknown kind", func(d *Device) { d.Kind = KindUnknown }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := valid
			tc.edit(&d)
			if err := d.Validate(); !errors.Is(err, ErrInvalidDevice) {
				t.Fatalf("Validate() error = %v, want ErrInvalidDevice", err)
			}
		})
	}
}
