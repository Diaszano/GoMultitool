// Package backup defines verifiable backup metadata.
package backup

import (
	"encoding/hex"
	"errors"
	"strings"
)

// ErrInvalidManifest indicates missing or inconsistent backup metadata.
var ErrInvalidManifest = errors.New("invalid backup manifest")

// Manifest records the source and integrity of a completed backup.
type Manifest struct {
	Version           int    `json:"version"`
	SourceID          string `json:"source_id"`
	SourcePath        string `json:"source_path"`
	SizeBytes         int64  `json:"size_bytes"`
	LogicalSectorSize int64  `json:"logical_sector_size"`
	SHA256            string `json:"sha256"`
	Status            Status `json:"status"`
}

// Validate permits restoration only for complete, well-formed backups.
func (m Manifest) Validate() error {
	digest, err := hex.DecodeString(m.SHA256)
	if m.Version != 1 || m.SourceID == "" || m.SourcePath == "" || m.SizeBytes <= 0 ||
		(m.LogicalSectorSize != 512 && m.LogicalSectorSize != 4096) ||
		m.SizeBytes%m.LogicalSectorSize != 0 ||
		err != nil ||
		len(digest) != 32 ||
		strings.ToLower(m.SHA256) != m.SHA256 ||
		m.Status != StatusComplete {
		return ErrInvalidManifest
	}
	return nil
}
