package filesystem

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls []string
	fail  string
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	call := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, call)
	if name == r.fail {
		return nil, errors.New("command failed")
	}
	if name == "losetup" && len(args) > 0 && args[0] == "--find" {
		return []byte("/dev/loop7\n"), nil
	}
	return nil, nil
}

func imageFixture(t *testing.T, ext4 bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "disk.img")
	data := make([]byte, 4*1024*1024)
	data[510], data[511] = 0x55, 0xaa
	entry := data[446:462]
	entry[4] = 0x83
	binary.LittleEndian.PutUint32(entry[8:12], 2048)
	binary.LittleEndian.PutUint32(entry[12:16], 4096)
	if ext4 {
		data[2048*512+1080], data[2048*512+1081] = 0x53, 0xef
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImageMounterMountsAndCleansUp(t *testing.T) {
	runner := &fakeRunner{}
	mount, cleanup, err := (ImageMounter{Runner: runner}).Mount(
		context.Background(),
		imageFixture(t, true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(mount); err != nil {
		t.Fatal(err)
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 4 || !strings.HasPrefix(runner.calls[1], "mount /dev/loop7p1 ") {
		t.Fatalf("calls=%v", runner.calls)
	}
}

func TestImageMounterRejectsMalformedAndCleansLoopOnMountFailure(t *testing.T) {
	runner := &fakeRunner{fail: "mount"}
	_, _, err := (ImageMounter{Runner: runner}).Mount(context.Background(), imageFixture(t, true))
	if err == nil {
		t.Fatal("mount failure ignored")
	}
	if !strings.HasPrefix(runner.calls[len(runner.calls)-1], "losetup -d /dev/loop7") {
		t.Fatalf("cleanup calls=%v", runner.calls)
	}
	runner = &fakeRunner{}
	if _, _, err := (ImageMounter{Runner: runner}).Mount(
		context.Background(),
		imageFixture(t, false),
	); err == nil {
		t.Fatal("image without ext4 root accepted")
	}
}
