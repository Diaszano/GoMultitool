// Package compression opens raw and XZ-compressed disk images as streams.
package compression

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Diaszano/GoMultitool/internal/domain/image"
)

type commandReader struct {
	stdout io.ReadCloser
	cmd    *exec.Cmd
	closed bool
}

// Open returns a bounded-memory reader for a raw or XZ image.
func Open(ctx context.Context, path string) (io.ReadCloser, error) {
	format, err := image.FormatFromPath(path)
	if err != nil {
		return nil, err
	}
	if format == image.FormatRaw {
		return os.Open(path)
	}
	cmd := exec.CommandContext(ctx, "xz", "-dc", "--", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &commandReader{stdout: stdout, cmd: cmd}, nil
}

// Inspect counts uncompressed bytes while streaming the image.
func Inspect(ctx context.Context, path string) (image.Image, error) {
	format, err := image.FormatFromPath(path)
	if err != nil {
		return image.Image{}, err
	}
	r, err := Open(ctx, path)
	if err != nil {
		return image.Image{}, err
	}
	var size int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := r.Read(buf)
		size += int64(n)
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = r.Close()
			return image.Image{}, readErr
		}
	}
	if err := r.Close(); err != nil {
		return image.Image{}, err
	}
	return image.Image{Path: path, Format: format, SizeBytes: size}, nil
}

func (r *commandReader) Read(p []byte) (int, error) { return r.stdout.Read(p) }

func (r *commandReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	_ = r.stdout.Close()
	if err := r.cmd.Wait(); err != nil {
		return fmt.Errorf("decompress image: %w", err)
	}
	return nil
}
