package imageconfig

import (
	"errors"
	"fmt"
)

// MemoryPreset requests an optional swap file in an offline image.
type MemoryPreset struct {
	SwapMiB      int
	AvailableMiB int
	Priority     int
	RootOnEMMC   bool
}

// MemoryArtifact contains the configuration file and an eMMC wear warning.
type MemoryArtifact struct {
	Data    []byte
	Warning string
}

// RenderMemory emits an idempotent systemd swap unit only when explicitly requested.
func RenderMemory(p MemoryPreset, zramManagerExists bool) (MemoryArtifact, error) {
	if p.SwapMiB <= 0 || p.Priority < 0 || p.Priority > 32767 {
		return MemoryArtifact{}, errors.New("invalid swap size or priority")
	}
	if zramManagerExists {
		return MemoryArtifact{}, errors.New(
			"existing ZRAM manager prevents duplicate memory configuration",
		)
	}
	if p.AvailableMiB < p.SwapMiB {
		return MemoryArtifact{}, errors.New("insufficient image space for requested swap file")
	}
	warning := ""
	if p.RootOnEMMC {
		warning = "swap writes increase eMMC wear"
	}
	return MemoryArtifact{
		Data:    []byte(fmt.Sprintf("[Swap]\nWhat=/swapfile\nPriority=%d\n", p.Priority)),
		Warning: warning,
	}, nil
}
