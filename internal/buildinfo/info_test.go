package buildinfo

import "testing"

func TestCurrentDefaults(t *testing.T) {
	got := Current()
	if got.Version != "dev" {
		t.Errorf("Version = %q, want %q", got.Version, "dev")
	}
	if got.Commit != "unknown" {
		t.Errorf("Commit = %q, want %q", got.Commit, "unknown")
	}
	if got.BuildTime != "unknown" {
		t.Errorf("BuildTime = %q, want %q", got.BuildTime, "unknown")
	}
}
