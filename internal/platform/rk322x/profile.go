// Package rk322x matches exact Rockchip RK322x device-tree compatibility data.
package rk322x

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

var (
	// ErrUnknownBoard indicates no configured profile matches the device tree.
	ErrUnknownBoard = errors.New("unknown RK322x board")
	// ErrAmbiguousBoard indicates multiple profiles match the device tree.
	ErrAmbiguousBoard = errors.New("ambiguous RK322x board profile")
)

// Profile records exact compatibility and provenance metadata for a board.
type Profile struct {
	Version                     int           `json:"version"`
	Name                        string        `json:"name"`
	Compatible                  []string      `json:"compatible"`
	StorageKinds                []device.Kind `json:"storage_kinds"`
	BootloaderSHA256            string        `json:"bootloader_sha256"`
	KernelSHA256                string        `json:"kernel_sha256"`
	AllowBootloaderModification bool          `json:"allow_bootloader_modification"`
}

// Load reads profiles from JSON, which is also valid YAML 1.2 syntax.
func Load(r io.Reader) ([]Profile, error) {
	var profiles []Profile
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&profiles); err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if p.Version != 1 || p.Name == "" || len(p.Compatible) == 0 {
			return nil, fmt.Errorf("invalid RK322x profile %q", p.Name)
		}
	}
	return profiles, nil
}

// Match requires exactly one profile to share a compatible string with the board.
func Match(profiles []Profile, compatible []string) (Profile, error) {
	match := -1
	for i := range profiles {
		for _, want := range profiles[i].Compatible {
			for _, got := range compatible {
				if got == want {
					if match >= 0 && match != i {
						return Profile{}, ErrAmbiguousBoard
					}
					match = i
					break
				}
			}
		}
	}
	if match < 0 {
		return Profile{}, ErrUnknownBoard
	}
	return profiles[match], nil
}

// SupportsStorage reports whether the profile recognizes this storage kind.
func (p Profile) SupportsStorage(kind device.Kind) bool {
	for _, supported := range p.StorageKinds {
		if kind == supported {
			return true
		}
	}
	return false
}
