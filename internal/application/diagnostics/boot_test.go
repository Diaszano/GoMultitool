package diagnostics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectBootAssets(t *testing.T) {
	root := t.TempDir()
	report := InspectBootAssets(root)
	if len(report.Missing) != 2 || report.KernelPresent || report.ExtlinuxConfigPresent {
		t.Fatalf("report=%+v", report)
	}
	if err := os.MkdirAll(filepath.Join(root, "extlinux"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "extlinux", "extlinux.conf"), []byte("LABEL test"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Image"), []byte("kernel"), 0600); err != nil {
		t.Fatal(err)
	}
	report = InspectBootAssets(root)
	if !report.KernelPresent || !report.ExtlinuxConfigPresent || len(report.Missing) != 0 {
		t.Fatalf("report=%+v", report)
	}
}
