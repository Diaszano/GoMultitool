// Package storage provides read-only access to block-device streams.
package storage

import (
	"context"
	"os"
)

// OpenReader opens a block device path for streamed reading.
func OpenReader(ctx context.Context, path string) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.Open(path)
}
