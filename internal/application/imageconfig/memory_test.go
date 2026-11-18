package imageconfig

import (
	"strings"
	"testing"
)

func TestMemoryPresetSpaceAndExistingManager(t *testing.T) {
	p := MemoryPreset{SwapMiB: 512, AvailableMiB: 600, RootOnEMMC: true}
	if _, err := RenderMemory(p, true); err == nil {
		t.Fatal("existing zram manager must prevent duplicate configuration")
	}
	if _, err := RenderMemory(p, false); err != nil {
		t.Fatal(err)
	}
	p.AvailableMiB = 511
	if _, err := RenderMemory(p, false); err == nil {
		t.Fatal("insufficient space accepted")
	}
}

func TestMemoryPresetPriorityAndIdempotentArtifact(t *testing.T) {
	p := MemoryPreset{SwapMiB: 512, AvailableMiB: 1024, RootOnEMMC: true, Priority: 10}
	a, err := RenderMemory(p, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderMemory(p, false)
	if err != nil || string(a.Data) != string(b.Data) || a.Data == nil {
		t.Fatalf("not idempotent: %q %q err=%v", a.Data, b.Data, err)
	}
	if !strings.Contains(string(a.Data), "Priority=10") || a.Warning == "" {
		t.Fatalf("preset=%+v", a)
	}
}
