package linux

import (
	"context"
	"os"
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

func TestSysfsCatalogRejectsMountedPartitionsAndRootMedia(t *testing.T) {
	root := t.TempDir()
	sys := filepath.Join(root, "sys")
	proc := filepath.Join(root, "proc")
	if err := os.MkdirAll(proc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proc, "mounts"), []byte("/dev/mmcblk1p1 /boot vfat rw 0 0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	devLink := filepath.Join(root, "dev/disk/by-partuuid")
	if err := os.MkdirAll(devLink, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../../mmcblk1p3", filepath.Join(devLink, "recovery-root")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proc, "cmdline"), []byte("console=ttyS2 root=PARTUUID=recovery-root ro\n"), 0644); err != nil {
		t.Fatal(err)
	}
	copyFixtureBlockDevices(t, sys)
	got, err := (SysfsCatalog{SysRoot: sys, ProcRoot: proc, DevRoot: filepath.Join(root, "dev")}).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range got {
		if d.ID == "mmcblk1" && (!d.Mounted || !d.BootMedia) {
			t.Fatalf("recovery medium was not protected: %+v", d)
		}
	}
}

func TestSysfsCapacityUses512ByteSectors(t *testing.T) {
	root := t.TempDir()
	deviceRoot := filepath.Join(root, "sys/block/sda")
	for _, dir := range []string{filepath.Join(deviceRoot, "queue"), filepath.Join(deviceRoot, "device")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, value := range map[string]string{"size": "1000", "removable": "0", "queue/logical_block_size": "4096", "device/transport": "usb"} {
		if err := os.WriteFile(filepath.Join(deviceRoot, name), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	d, err := (SysfsCatalog{SysRoot: filepath.Join(root, "sys"), ProcRoot: filepath.Join(root, "proc")}).Resolve(context.Background(), "/dev/sda")
	if err != nil {
		t.Fatal(err)
	}
	if d.SizeBytes != 512000 {
		t.Fatalf("size = %d, want 512000", d.SizeBytes)
	}
}

func TestSysfsCatalogFailsClosedWhenProcStateIsUnavailable(t *testing.T) {
	root := t.TempDir()
	sys := filepath.Join(root, "sys")
	copyFixtureBlockDevices(t, sys)
	got, err := (SysfsCatalog{SysRoot: sys, ProcRoot: filepath.Join(root, "missing-proc")}).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range got {
		if !d.Mounted || !d.BootMedia {
			t.Fatalf("unsafe missing proc state accepted for %s: %+v", d.ID, d)
		}
	}
}

func copyFixtureBlockDevices(t *testing.T, sys string) {
	t.Helper()
	fixture := filepath.Join("..", "..", "..", "tests", "fixtures", "sysfs", "block")
	if err := os.CopyFS(filepath.Join(sys, "block"), os.DirFS(fixture)); err != nil {
		t.Fatal(err)
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
