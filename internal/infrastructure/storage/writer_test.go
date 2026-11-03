package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := OpenWriter(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteAt([]byte("new"), 0); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("content=%q", got)
	}
}
