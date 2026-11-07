package multitool

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type call struct{ name, args string }

type fakeRunner struct {
	blkid string
	calls []call
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, call{name, strings.Join(args, " ")})
	if name == "blkid" {
		return []byte(f.blkid), nil
	}
	return nil, nil
}

func procWith(t *testing.T, mounts string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mounts"), []byte(mounts), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestOpenUsesExistingMount(t *testing.T) {
	r := &fakeRunner{blkid: "/dev/mmcblk0p1\n"}
	root := t.TempDir()
	proc := procWith(t, "proc /proc proc rw 0 0\n/dev/mmcblk0p1 "+root+" ntfs rw 0 0\n")
	ws, closeFn, err := Locator{
		Runner:     r,
		ProcRoot:   proc,
		MountPoint: "/unused",
	}.Open(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Root != root {
		t.Fatalf("root = %q", ws.Root)
	}
	if err := closeFn(); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 1 {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestOpenMountsAndUnmounts(t *testing.T) {
	r := &fakeRunner{blkid: "/dev/mmcblk0p1\n"}
	mp := t.TempDir()
	if err := os.WriteFile(filepath.Join(mp, "TARGET"), []byte("rk322x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	proc := procWith(t, "proc /proc proc rw 0 0\n")
	ws, closeFn, err := Locator{
		Runner:     r,
		ProcRoot:   proc,
		MountPoint: mp,
	}.Open(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Platform != "rk322x" || ws.Root != mp {
		t.Fatalf("ws = %+v", ws)
	}
	if st, err := os.Stat(ws.BackupsDir()); err != nil || !st.IsDir() {
		t.Fatalf("backups dir missing: %v", err)
	}
	if _, err := ws.FreeBytes(); err != nil {
		t.Fatal(err)
	}
	if err := closeFn(); err != nil {
		t.Fatal(err)
	}
	want := []call{
		{"blkid", "-l -o device --label MULTITOOL"},
		{"mount", "/dev/mmcblk0p1 " + mp},
		{"umount", mp},
	}
	if len(r.calls) != 3 || r.calls[0] != want[0] || r.calls[1] != want[1] ||
		r.calls[2] != want[2] {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestOpenMissingPartition(t *testing.T) {
	r := &fakeRunner{}
	_, _, err := Locator{Runner: r, ProcRoot: procWith(t, ""), MountPoint: t.TempDir()}.Open(
		context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	for _, c := range r.calls {
		if c.name == "mount" {
			t.Fatal("mount called")
		}
	}
}

func TestLockSecondCallerGetsBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "l.lock")
	unlock, err := Lock(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lock(path); !errors.Is(err, ErrBusy) {
		t.Fatalf("err = %v", err)
	}
	if err := unlock(); err != nil {
		t.Fatal(err)
	}
	u3, err := Lock(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = u3()
}

func TestOpenMatchesMountThroughSymlink(t *testing.T) {
	base := t.TempDir()
	dev0 := filepath.Join(base, "dev", "mmcblk0p1")
	link := filepath.Join(base, "dev", "disk", "by-label", "MULTITOOL")
	for _, d := range []string{filepath.Dir(dev0), filepath.Dir(link)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(dev0, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(dev0, link); err != nil {
		t.Fatal(err)
	}
	r := &fakeRunner{blkid: dev0 + "\n"}
	root := t.TempDir()
	proc := procWith(t, link+" "+root+" ntfs rw 0 0\n")
	ws, _, err := Locator{
		Runner:     r,
		ProcRoot:   proc,
		MountPoint: "/unused",
	}.Open(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Root != root || len(r.calls) != 1 {
		t.Fatalf("root=%q calls=%v", ws.Root, r.calls)
	}
}

func TestOpenDecodesMountEscapes(t *testing.T) {
	r := &fakeRunner{blkid: "/dev/mmcblk0p1\n"}
	root := filepath.Join(t.TempDir(), "My Disk")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	proc := procWith(t, "/dev/mmcblk0p1 "+strings.ReplaceAll(root, " ", `\040`)+" ntfs rw 0 0\n")
	ws, _, err := Locator{
		Runner:     r,
		ProcRoot:   proc,
		MountPoint: "/unused",
	}.Open(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Root != root {
		t.Fatalf("root = %q", ws.Root)
	}
}
