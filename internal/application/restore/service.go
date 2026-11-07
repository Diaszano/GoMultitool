// Package restore verifies backup data before writing it to a guarded target.
package restore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

var (
	// ErrConfirmationRequired indicates a destructive operation was not approved.
	ErrConfirmationRequired = errors.New("explicit restore confirmation required")
	// ErrChecksumMismatch indicates the backup or readback differs from its manifest.
	ErrChecksumMismatch = errors.New("backup checksum mismatch")
)

// Service restores verified backup data to a guarded device.
type Service struct{}

// Execute verifies the source, revalidates the target, writes, syncs, and reads back.
func (Service) Execute(
	ctx context.Context,
	manifest backup.Manifest,
	source io.ReadSeeker,
	plan storageguard.Plan,
	confirmed bool,
	guard storageguard.Guard,
	openTarget func(string) (application.Target, error),
) (retErr error) {
	if err := manifest.Validate(); err != nil {
		return err
	}
	if !confirmed {
		return ErrConfirmationRequired
	}
	if err := plan.Device.ValidateWriteTarget(); err != nil {
		return err
	}
	if plan.Device.SizeBytes < manifest.SizeBytes {
		return device.ErrUnsafeDevice
	}
	if guard.Catalog == nil || openTarget == nil {
		return device.ErrUnsafeDevice
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return err
	}
	digest, err := hashExact(ctx, source, manifest.SizeBytes)
	if err != nil {
		return err
	}
	if hex.EncodeToString(digest) != manifest.SHA256 {
		return ErrChecksumMismatch
	}
	if err := guard.Revalidate(ctx, plan); err != nil {
		return err
	}
	target, err := openTarget(plan.Device.Path)
	if err != nil {
		return err
	}
	defer func() {
		if err := target.Close(); retErr == nil {
			retErr = err
		}
	}()
	if err := guard.Revalidate(ctx, plan); err != nil {
		return err
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := target.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := application.CopyExact(ctx, target, source, manifest.SizeBytes); err != nil {
		return err
	}
	if err := target.Sync(); err != nil {
		return err
	}
	if _, err := target.Seek(0, io.SeekStart); err != nil {
		return err
	}
	readback, err := hashExact(ctx, target, manifest.SizeBytes)
	if err != nil {
		return err
	}
	if hex.EncodeToString(readback) != manifest.SHA256 {
		return ErrChecksumMismatch
	}
	return nil
}

func hashExact(ctx context.Context, src io.Reader, size int64) ([]byte, error) {
	h := sha256.New()
	if err := application.CopyExact(ctx, h, src, size); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
