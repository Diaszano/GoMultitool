package linux

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestHasZRAMManagerAndSwapPriorities(t *testing.T) {
	root := t.TempDir()
	service := filepath.Join(root, "usr/lib/systemd/system/zramswap.service")
	if err := os.MkdirAll(filepath.Dir(service), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(service, []byte("[Service]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !HasZRAMManager(root) {
		t.Fatal("ZRAM manager not detected")
	}
	fstab := filepath.Join(root, "etc/fstab")
	if err := os.MkdirAll(filepath.Dir(fstab), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fstab, []byte("/swapfile none swap sw 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !HasSwapConfiguration(root) {
		t.Fatal("existing swap config not detected")
	}
	proc := filepath.Join(root, "swaps")
	data := "Filename Type Size Used Priority\n/dev/zram0 partition 512 0 100\n/swapfile file 1024 0 -2\n"
	if err := os.WriteFile(proc, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	priorities, err := ReadSwapPriorities(proc)
	if err != nil || !reflect.DeepEqual(priorities, []int{100, -2}) {
		t.Fatalf("priorities=%v error=%v", priorities, err)
	}
}
