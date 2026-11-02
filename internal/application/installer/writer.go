package installer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

// ErrConfirmationRequired indicates an image write was not explicitly approved.
var ErrConfirmationRequired = errors.New("explicit installation confirmation required")

// Target supports synchronized writes and readback verification.
type Target interface {
	io.ReadWriteSeeker
	Sync() error
	Close() error
}

// Writer streams an image into a revalidated destination and verifies readback.
type Writer struct{}

// Execute streams the planned image to its revalidated target and verifies readback.
func (Writer) Execute(ctx context.Context, plan installation.Plan, source io.Reader, targetPlan storageguard.Plan, confirmed bool, guard storageguard.Guard, openTarget func(string) (Target, error)) (retErr error) {
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
	digest, err := copyAndHash(ctx, target, source, plan.Image.SizeBytes)
	if err != nil {
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
	readback, err := hashExact(ctx, target, plan.Image.SizeBytes)
	if err != nil {
		return err
	}
	if !bytes.Equal(digest, readback) {
		return ErrReadbackMismatch
	}
	return nil
}

// ErrReadbackMismatch indicates the target differs from the image stream.
var ErrReadbackMismatch = errors.New("image readback mismatch")

func copyAndHash(ctx context.Context, dst io.Writer, src io.Reader, size int64) ([]byte, error) {
	hash := sha256.New()
	buf := make([]byte, 32*1024)
	remaining := size
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		want := int64(len(buf))
		if remaining < want {
			want = remaining
		}
		n, err := io.ReadFull(src, buf[:want])
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return nil, writeErr
			}
			if written != n {
				return nil, io.ErrShortWrite
			}
			_, _ = hash.Write(buf[:n])
			remaining -= int64(n)
		}
		if err != nil {
			return nil, err
		}
	}
	return hash.Sum(nil), nil
}

func hashExact(ctx context.Context, src io.Reader, size int64) ([]byte, error) {
	h := sha256.New()
	buf := make([]byte, 32*1024)
	remaining := size
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		want := int64(len(buf))
		if remaining < want {
			want = remaining
		}
		n, err := io.ReadFull(src, buf[:want])
		if n > 0 {
			_, _ = h.Write(buf[:n])
			remaining -= int64(n)
		}
		if err != nil {
			return nil, err
		}
	}
	return h.Sum(nil), nil
}
