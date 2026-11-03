package compression

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOpenRawImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.img")
	if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if string(got) != "image" {
		t.Fatalf("content = %q", got)
	}
}

func TestInspectXZAndRejectsCorruption(t *testing.T) {
	if _, err := exec.LookPath("xz"); err != nil {
		t.Skip("xz unavailable")
	}
	dir := t.TempDir()
	raw := filepath.Join(dir, "image.img")
	if err := os.WriteFile(raw, []byte("image data"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("xz", "-c", raw).Output()
	if err != nil {
		t.Fatal(err)
	}
	xz := filepath.Join(dir, "image.img.xz")
	if err := os.WriteFile(xz, out, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Inspect(context.Background(), xz)
	if err != nil {
		t.Fatal(err)
	}
	if got.SizeBytes != int64(len("image data")) {
		t.Fatalf("size = %d", got.SizeBytes)
	}
	bad := filepath.Join(dir, "bad.img.xz")
	if err := os.WriteFile(bad, []byte("not xz"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(context.Background(), bad); err == nil {
		t.Fatal("corrupt XZ accepted")
	}
}

func TestInspectRejectsTruncatedXZ(t *testing.T) {
	if _, err := exec.LookPath("xz"); err != nil {
		t.Skip("xz unavailable")
	}
	path := filepath.Join(t.TempDir(), "truncated.img.xz")
	if err := os.WriteFile(path, []byte{0xfd, '7', 'z', 'X', 'Z', 0, 1}, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(context.Background(), path); err == nil {
		t.Fatal("truncated XZ accepted")
	}
}
