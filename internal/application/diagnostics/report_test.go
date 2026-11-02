package diagnostics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootReportMissingJournalFilesAndRootUUIDMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "etc/fstab"), []byte("UUID=wrong / ext4 defaults 0 1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := CollectBootReport(root, "root=UUID=expected")
	if r.JournalAvailable {
		t.Fatal("absent journal reported available")
	}
	if !strings.Contains(strings.Join(r.Observations, " "), "wrong") {
		t.Fatalf("observations=%v", r.Observations)
	}
	if len(r.Hypotheses) == 0 {
		t.Fatal("mismatch hypothesis missing")
	}
	if len(r.Missing) == 0 {
		t.Fatal("missing boot files not reported")
	}
}
func TestBootReportMatchingRootUUID(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "etc/fstab"), []byte("UUID=expected / ext4 defaults 0 1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := CollectBootReport(root, "root=UUID=expected")
	if len(r.Hypotheses) != 0 {
		t.Fatalf("hypotheses=%v", r.Hypotheses)
	}
}
