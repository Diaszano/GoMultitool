package compression

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/image"
)

// Size returns the uncompressed size of an image without decompressing it.
// Gzip stores no reliable size, so it returns 0 (unknown); use Inspect.
func Size(ctx context.Context, path string) (int64, error) {
	format, err := image.FormatFromPath(path)
	if err != nil {
		return 0, err
	}
	switch format {
	case image.FormatRaw:
		info, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	case image.FormatXZ:
		out, err := exec.CommandContext(ctx, "xz", "--robot", "--list", "--", path).Output()
		if err != nil {
			return 0, fmt.Errorf("list xz image: %w", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if f := strings.Split(line, "\t"); f[0] == "totals" && len(f) > 4 {
				return strconv.ParseInt(f[4], 10, 64)
			}
		}
		return 0, fmt.Errorf("list xz image: no totals line")
	}
	return 0, nil
}

// ListImages returns the regular image files in dir, sorted by name.
// A missing dir yields an empty list; files whose size cannot be read are skipped.
func ListImages(ctx context.Context, dir string) ([]image.Image, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var images []image.Image
	for _, e := range entries { // ReadDir already sorts by name
		if !e.Type().IsRegular() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		format, err := image.FormatFromPath(path)
		if err != nil {
			continue
		}
		size, err := Size(ctx, path)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if err != nil {
			continue // one corrupt image must not hide the others
		}
		images = append(images, image.Image{Path: path, Format: format, SizeBytes: size})
	}
	sort.SliceStable(images, func(i, j int) bool { return images[i].Path < images[j].Path })
	return images, nil
}
