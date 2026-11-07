package installer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

// ErrConfirmationRequired indicates an image write was not explicitly approved.
var ErrConfirmationRequired = errors.New("explicit installation confirmation required")

// Writer streams an image into a revalidated destination and verifies readback.
type Writer struct{}

// Execute streams the planned image to its revalidated target and verifies readback.
func (Writer) Execute(
	ctx context.Context,
	plan installation.Plan,
	source io.Reader,
	targetPlan storageguard.Plan,
	confirmed bool,
	guard storageguard.Guard,
	openTarget func(string) (application.Target, error),
) (retErr error) {
	if err := plan.Validate(); err != nil {
		return err
	}
	if !confirmed {
		return ErrConfirmationRequired
	}
	if plan.Target != targetPlan.Device || guard.Catalog == nil || openTarget == nil {
		return ErrIdentityMismatch
	}
	if err := guard.Revalidate(ctx, targetPlan); err != nil {
		return err
	}
	target, err := openTarget(targetPlan.Device.Path)
	if err != nil {
		return err
	}
	defer func() {
		if err := target.Close(); retErr == nil {
			retErr = err
		}
	}()
	if err := guard.Revalidate(ctx, targetPlan); err != nil {
		return err
	}
	if _, err := target.Seek(0, io.SeekStart); err != nil {
		return err
	}
	written := sha256.New()
	if err := application.CopyExact(
		ctx,
		io.MultiWriter(target, written),
		source,
		plan.Image.SizeBytes,
	); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := target.Sync(); err != nil {
		return err
	}
	if _, err := target.Seek(0, io.SeekStart); err != nil {
		return err
	}
	readback := sha256.New()
	if err := application.CopyExact(ctx, readback, target, plan.Image.SizeBytes); err != nil {
		return err
	}
	if !bytes.Equal(written.Sum(nil), readback.Sum(nil)) {
		return ErrReadbackMismatch
	}
	return nil
}

// ErrReadbackMismatch indicates the target differs from the image stream.
var ErrReadbackMismatch = errors.New("image readback mismatch")
