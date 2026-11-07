package compression

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSizeXZReadsIndex(t *testing.T) {
	if _, err := exec.LookPath("xz"); err != nil {
		t.Skip("xz unavailable")
	}
	got, err := Size(context.Background(), fixtures+"sample.img.xz")
	if err != nil || got != 20 {
		t.Fatalf("Size = %d, %v", got, err)
	}
}

func TestListImagesHandlesSpacesAndSkipsOthers(t *testing.T) {
	if _, err := exec.LookPath("xz"); err != nil {
		t.Skip("xz unavailable")
	}
	dir := t.TempDir()
	xz, err := os.ReadFile(fixtures + "sample.img.xz")
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"Armbian (1) d'x.img.xz": xz, "b.img": []byte("12345"), "notes.txt": []byte("x"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "b.img"), filepath.Join(dir, "c.img")); err != nil {
		t.Fatal(err)
	}
	got, err := ListImages(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || filepath.Base(got[0].Path) != "Armbian (1) d'x.img.xz" ||
		got[0].SizeBytes != 20 || filepath.Base(got[1].Path) != "b.img" ||
		got[1].SizeBytes != 5 {
		t.Fatalf("images = %+v", got)
	}
	missing, err := ListImages(context.Background(), filepath.Join(dir, "none"))
	if err != nil || len(missing) != 0 {
		t.Fatalf("missing dir = %v, %v", missing, err)
	}
}
