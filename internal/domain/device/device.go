// Package device defines storage-device metadata and validation.
package device

// Kind identifies a storage transport or device type.
type Kind string

const (
	// KindEMMC identifies embedded MMC storage.
	KindEMMC Kind = "emmc"
	// KindSD identifies removable SD storage.
	KindSD Kind = "sd"
	// KindUSB identifies USB storage.
	KindUSB Kind = "usb"
	// KindNAND identifies raw NAND storage.
	KindNAND Kind = "nand"
	// KindUnknown identifies storage that could not be classified.
	KindUnknown Kind = "unknown"
)

// Device describes a block storage device.
type Device struct {
	ID                string
	Path              string
	Kind              Kind
	SizeBytes         int64
	LogicalSectorSize int64
	Removable         bool
	Mounted           bool
	BootMedia         bool
}

// Validate rejects incomplete or unsupported device metadata.
func (d Device) Validate() error {
	if d.ID == "" || d.Path == "" || d.SizeBytes <= 0 ||
		(d.LogicalSectorSize != 512 && d.LogicalSectorSize != 4096) ||
		d.SizeBytes%d.LogicalSectorSize != 0 {
		return ErrInvalidDevice
	}
	switch d.Kind {
	case KindEMMC, KindSD, KindUSB, KindNAND:
		return nil
	default:
		return ErrInvalidDevice
	}
}
