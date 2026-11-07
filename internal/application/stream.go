package application

import (
	"context"
	"errors"
	"io"
)

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (c contextReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

// ContextReader stops reading once ctx is canceled.
func ContextReader(ctx context.Context, r io.Reader) io.Reader { return contextReader{ctx, r} }

// CopyExact copies exactly size bytes and reports a short source as io.ErrUnexpectedEOF.
func CopyExact(ctx context.Context, dst io.Writer, src io.Reader, size int64) error {
	_, err := io.CopyN(dst, ContextReader(ctx, src), size)
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}
