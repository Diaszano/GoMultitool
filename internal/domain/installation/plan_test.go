package installation

import (
	"errors"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
)

func TestPlanValidateCapacityAndSafety(t *testing.T) {
	target := device.Device{
		ID:                "usb1",
		Path:              "/dev/sdb",
		Kind:              device.KindUSB,
		SizeBytes:         1024,
		LogicalSectorSize: 512,
	}
	plan := Plan{
		Image:   image.Image{Path: "os.img", Format: image.FormatRaw, SizeBytes: 512},
		Target:  target,
		Summary: "write image",
	}
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
	plan.Image.SizeBytes = 2048
	if !errors.Is(plan.Validate(), ErrImageTooLarge) {
		t.Fatalf("error=%v", plan.Validate())
	}
	plan.Image.SizeBytes = 512
	plan.Target.BootMedia = true
	if !errors.Is(plan.Validate(), device.ErrUnsafeDevice) {
		t.Fatalf("error=%v", plan.Validate())
	}
	plan.Target.BootMedia = false
	plan.Target.Mounted = true
	if !errors.Is(plan.Validate(), device.ErrUnsafeDevice) {
		t.Fatalf("error=%v", plan.Validate())
	}
}
