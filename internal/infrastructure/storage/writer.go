package storage

import (
	"context"
	"os"
)

// OpenWriter opens an existing block-device path for synchronized writes.
func OpenWriter(ctx context.Context, path string) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR, 0)
}
