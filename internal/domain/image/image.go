// Package image defines supported disk-image metadata.
package image

import (
	"errors"
	"path/filepath"
	"strings"
)

// Format identifies a supported disk-image encoding.
type Format string

const (
	// FormatRaw is an uncompressed disk image.
	FormatRaw Format = "raw"
	// FormatXZ is an XZ-compressed disk image.
	FormatXZ Format = "xz"
)

// ErrUnsupportedFormat indicates the path does not identify a supported image.
var ErrUnsupportedFormat = errors.New("unsupported image format")

// ErrSizeMismatch indicates image data differs from the expected byte count.
var ErrSizeMismatch = errors.New("image size mismatch")

// Image describes a disk image without loading its content into memory.
type Image struct {
	Path      string
	Format    Format
	SizeBytes int64
}

// FormatFromPath identifies raw .img and compressed .img.xz images.
func FormatFromPath(path string) (Format, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".img":
		return FormatRaw, nil
	case ".xz":
		if strings.HasSuffix(strings.ToLower(path), ".img.xz") {
			return FormatXZ, nil
		}
	}
	return "", ErrUnsupportedFormat
}

// ValidateSize compares the streamed image size with the expected size.
func (i Image) ValidateSize(expected int64) error {
	if expected <= 0 || i.SizeBytes != expected {
		return ErrSizeMismatch
	}
	return nil
}
