// Package bootloader reads bootloader identity without modifying firmware.
package bootloader

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"slices"
)

// Inspection records the digest and whether it matches a known image.
type Inspection struct {
	SHA256 string
	Known  bool
}

// Inspect hashes bootloader data without modifying it.
func Inspect(r io.Reader, knownHashes []string) (Inspection, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return Inspection{}, err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	return Inspection{SHA256: sum, Known: slices.Contains(knownHashes, sum)}, nil
}
