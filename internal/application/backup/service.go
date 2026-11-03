// Package backup creates streamed, checksummed storage backups.
package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Diaszano/GoMultitool/internal/domain/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// ErrInsufficientSpace indicates the destination cannot hold data and metadata.
var ErrInsufficientSpace = errors.New("insufficient destination space")

// Service creates a complete backup and its manifest.
type Service struct{}

// Create streams exactly the source size and finalizes data before its manifest.
func (Service) Create(
	ctx context.Context,
	source device.Device,
	src io.Reader,
	destination string,
	availableBytes int64,
) (backup.Manifest, error) {
	if err := source.Validate(); err != nil {
		return backup.Manifest{}, err
	}
	if availableBytes < source.SizeBytes || availableBytes-source.SizeBytes < 512 {
		return backup.Manifest{}, ErrInsufficientSpace
	}
	manifestPath := destination + ".json"
	for _, path := range []string{destination, manifestPath} {
		if _, err := os.Lstat(path); err == nil {
			return backup.Manifest{}, fmt.Errorf("backup output already exists: %s", path)
		} else if !os.IsNotExist(err) {
			return backup.Manifest{}, err
		}
	}
	dir := filepath.Dir(destination)
	dataTmp, err := os.CreateTemp(dir, ".backup-*")
	if err != nil {
		return backup.Manifest{}, err
	}
	tmpName := dataTmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	hash := sha256.New()
	buf := make([]byte, 32*1024)
	remaining := source.SizeBytes
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			_ = dataTmp.Close()
			return backup.Manifest{}, err
		}
		want := int64(len(buf))
		if remaining < want {
			want = remaining
		}
		n, readErr := io.ReadFull(src, buf[:want])
		if n > 0 {
			written, err := dataTmp.Write(buf[:n])
			if err != nil {
				_ = dataTmp.Close()
				return backup.Manifest{}, err
			}
			if written != n {
				_ = dataTmp.Close()
				return backup.Manifest{}, io.ErrShortWrite
			}
			_, _ = hash.Write(buf[:n])
			remaining -= int64(n)
		}
		if readErr != nil {
			_ = dataTmp.Close()
			return backup.Manifest{}, readErr
		}
	}
	if err := ctx.Err(); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, err
	}
	if err := dataTmp.Sync(); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, err
	}
	if err := dataTmp.Close(); err != nil {
		return backup.Manifest{}, err
	}
	if err := os.Rename(tmpName, destination); err != nil {
		return backup.Manifest{}, err
	}
	manifest := backup.Manifest{
		Version:           1,
		SourceID:          source.ID,
		SourcePath:        source.Path,
		SizeBytes:         source.SizeBytes,
		LogicalSectorSize: source.LogicalSectorSize,
		SHA256:            hex.EncodeToString(hash.Sum(nil)),
		Status:            backup.StatusComplete,
	}
	manifestTmp, err := os.CreateTemp(dir, ".manifest-*")
	if err != nil {
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	manifestTmpName := manifestTmp.Name()
	defer func() { _ = os.Remove(manifestTmpName) }()
	if err := json.NewEncoder(manifestTmp).Encode(manifest); err != nil {
		_ = manifestTmp.Close()
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	if err := manifestTmp.Sync(); err != nil {
		_ = manifestTmp.Close()
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	if err := manifestTmp.Close(); err != nil {
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	if err := os.Rename(manifestTmpName, manifestPath); err != nil {
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	return manifest, nil
}
