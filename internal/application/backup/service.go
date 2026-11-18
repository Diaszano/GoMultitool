// Package backup creates streamed, checksummed storage backups.
package backup

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/domain/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// ErrNoSpace indicates the backup partition filled up while writing.
var ErrNoSpace = errors.New("no space left on backup partition")

// ErrExists indicates the backup output already exists.
var ErrExists = errors.New("backup already exists")

// dataSink is the writer the gzip stream goes through; tests replace it to inject faults.
var dataSink = func(f *os.File) io.Writer { return f }

// noSpace wraps any ENOSPC in err as ErrNoSpace.
func noSpace(err error) error {
	if errors.Is(err, syscall.ENOSPC) {
		return fmt.Errorf("%w: %w", ErrNoSpace, err)
	}
	return err
}

// Service creates a complete backup and its manifest.
type Service struct{}

// Create streams exactly the source size and finalizes data before its manifest.
func (Service) Create(
	ctx context.Context,
	source device.Device,
	src io.Reader,
	destination string,
) (backup.Manifest, error) {
	if err := source.Validate(); err != nil {
		return backup.Manifest{}, err
	}
	manifestPath := destination + ".json"
	for _, path := range []string{destination, manifestPath} {
		if _, err := os.Lstat(path); err == nil {
			return backup.Manifest{}, fmt.Errorf("%w: %s", ErrExists, path)
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
	zw, err := gzip.NewWriterLevel(dataSink(dataTmp), gzip.BestSpeed)
	if err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, err
	}
	if err := application.CopyExact(
		ctx,
		io.MultiWriter(zw, hash),
		src,
		source.SizeBytes,
	); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, noSpace(err)
	}
	if err := ctx.Err(); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, err
	}
	if err := zw.Close(); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, noSpace(err)
	}
	if err := dataTmp.Sync(); err != nil {
		_ = dataTmp.Close()
		return backup.Manifest{}, noSpace(err)
	}
	if err := dataTmp.Close(); err != nil {
		return backup.Manifest{}, noSpace(err)
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
		return backup.Manifest{}, noSpace(err)
	}
	if err := manifestTmp.Sync(); err != nil {
		_ = manifestTmp.Close()
		_ = os.Remove(destination)
		return backup.Manifest{}, noSpace(err)
	}
	if err := manifestTmp.Close(); err != nil {
		_ = os.Remove(destination)
		return backup.Manifest{}, noSpace(err)
	}
	if err := os.Rename(manifestTmpName, manifestPath); err != nil {
		_ = os.Remove(destination)
		return backup.Manifest{}, err
	}
	// Persist both renames before callers overwrite the source device.
	d, err := os.Open(dir)
	if err != nil {
		_ = os.Remove(destination)
		_ = os.Remove(manifestPath)
		return backup.Manifest{}, err
	}
	// Some NTFS/FUSE drivers reject directory fsync; that is not a failed backup.
	if err := d.Sync(); err != nil && !errors.Is(err, syscall.EINVAL) &&
		!errors.Is(err, syscall.ENOTSUP) {
		_ = d.Close()
		_ = os.Remove(destination)
		_ = os.Remove(manifestPath)
		return backup.Manifest{}, err
	}
	if err := d.Close(); err != nil {
		_ = os.Remove(destination)
		_ = os.Remove(manifestPath)
		return backup.Manifest{}, err
	}
	return manifest, nil
}
