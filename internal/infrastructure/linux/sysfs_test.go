package linux

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestSysfsCatalogListsKnownDevices(t *testing.T) {
	catalog := SysfsCatalog{SysRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "sysfs"), ProcRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "procfs")}
	got, err := catalog.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d devices, want 4", len(got))
	}
	want := map[string]device.Kind{"mmcblk0": device.KindEMMC, "mmcblk1": device.KindSD, "sda": device.KindUSB, "sdb": device.KindUnknown}
	for _, d := range got {
		if d.Kind != want[d.ID] {
			t.Errorf("%s kind = %q, want %q", d.ID, d.Kind, want[d.ID])
		}
		if d.Mounted != (d.ID == "mmcblk1") {
			t.Errorf("%s mounted = %v", d.ID, d.Mounted)
		}
	}
}

func TestSysfsCatalogReturnsUnknownForInsufficientEvidence(t *testing.T) {
	catalog := SysfsCatalog{SysRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "sysfs"), ProcRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "procfs")}
	d, err := catalog.Resolve(context.Background(), "/dev/sdb")
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != device.KindUnknown {
		t.Fatalf("kind = %q, want unknown", d.Kind)
	}
}

func TestSysfsCatalogResolveRejectsMissingDevice(t *testing.T) {
	catalog := SysfsCatalog{SysRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "sysfs"), ProcRoot: filepath.Join("..", "..", "..", "tests", "fixtures", "procfs")}
	if _, err := catalog.Resolve(context.Background(), "/dev/not-there"); err == nil {
		t.Fatal("Resolve() succeeded for missing device")
	}
}
