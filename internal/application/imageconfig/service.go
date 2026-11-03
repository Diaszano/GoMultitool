package imageconfig

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/infrastructure/filesystem"
)

// ErrOutputExists prevents replacing an existing prepared image.
var ErrOutputExists = errors.New("prepared image output already exists")

// RootFSMounter attaches the copied image and exposes its root filesystem.
type RootFSMounter interface {
	Mount(context.Context, string) (string, func() error, error)
}

// Service applies a preset to a private copy and publishes it only after cleanup.
type Service struct {
	Profile Compatibility
	Mounter RootFSMounter
}

// Prepare modifies a temporary image copy and atomically renames it on success.
func (s Service) Prepare(
	ctx context.Context,
	sourcePath, outputPath, version string,
	p Preset,
) error {
	if s.Mounter == nil {
		return errors.New("image mounter is required")
	}
	sourceInfo, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceInfo.Mode().IsRegular() {
		return errors.New("source image must be a regular file")
	}
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	if sourceAbs == outputAbs {
		return errors.New("source and output paths must differ")
	}
	if _, err := os.Lstat(outputPath); err == nil {
		return ErrOutputExists
	} else if !os.IsNotExist(err) {
		return err
	}
	artifacts, err := Render(s.Profile, version, p)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".prepared-image-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	keepTemp := false
	defer func() {
		if !keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()
	src, err := os.Open(sourcePath)
	if err != nil {
		_ = tmp.Close()
		return err
	}
	if err = copyContext(ctx, tmp, src); err != nil {
		_ = src.Close()
		_ = tmp.Close()
		return err
	}
	if err = src.Close(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	rootPath, cleanup, err := s.Mounter.Mount(ctx, tmpPath)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		cleanupErr := cleanup()
		return errors.Join(err, cleanupErr)
	}
	keys := make([]string, 0, len(artifacts.Files))
	for name := range artifacts.Files {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		if err := ctx.Err(); err != nil {
			_ = root.Close()
			cleanupErr := cleanup()
			return errors.Join(err, cleanupErr)
		}
		if err := writeArtifact(root, name, artifacts.Files[name]); err != nil {
			_ = root.Close()
			cleanupErr := cleanup()
			return errors.Join(err, cleanupErr)
		}
	}
	if err := root.Close(); err != nil {
		cleanupErr := cleanup()
		return errors.Join(err, cleanupErr)
	}
	if err := cleanup(); err != nil {
		keepTemp = true
		return fmt.Errorf(
			"image mount cleanup failed; private temporary image retained at %s: %w",
			tmpPath,
			err,
		)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	prepared, err := os.OpenFile(tmpPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	if err := prepared.Sync(); err != nil {
		_ = prepared.Close()
		return err
	}
	if err := prepared.Close(); err != nil {
		return err
	}
	if _, err := os.Lstat(outputPath); err == nil {
		return ErrOutputExists
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmpPath, outputPath)
}

func writeArtifact(root *os.Root, name string, file File) error {
	clean := filepath.Clean(name)
	if filepath.IsAbs(name) || clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("invalid artifact path")
	}
	parent := filepath.Dir(clean)
	if parent != "." {
		if err := root.MkdirAll(parent, 0o700); err != nil {
			return err
		}
	}
	if info, err := root.Lstat(clean); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("refuse to replace non-regular image file %s", clean)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := root.OpenFile(clean, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode)
	if err != nil {
		return err
	}
	if _, err := f.Write(file.Data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Chmod(file.Mode); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func copyContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 64*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			written, err := dst.Write(buf[:n])
			if err != nil {
				return err
			}
			if written != n {
				return io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

var _ RootFSMounter = filesystem.ImageMounter{}
