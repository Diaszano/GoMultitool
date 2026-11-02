package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device")
	if err := os.WriteFile(path, []byte("device"), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := OpenReader(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	buf := make([]byte, 6)
	if _, err := r.Read(buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "device" {
		t.Fatalf("read %q", buf)
	}
}
