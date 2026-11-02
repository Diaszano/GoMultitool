package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CommandRunner runs one executable with explicit arguments.
type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}
type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// ImageMounter attaches a regular image file and mounts its ext4 root partition.
type ImageMounter struct{ Runner CommandRunner }

var loopPath = regexp.MustCompile(`^/dev/loop[0-9]+$`)

// Mount mounts the image root and returns a cleanup function for the mount and loop device.
func (m ImageMounter) Mount(ctx context.Context, imagePath string) (string, func() error, error) {
	info, err := os.Lstat(imagePath)
	if err != nil {
		return "", nil, err
	}
	if !info.Mode().IsRegular() {
		return "", nil, errors.New("image must be a regular file")
	}
	f, err := os.Open(imagePath)
	if err != nil {
		return "", nil, err
	}
	partition, err := findExt4Partition(f, info.Size())
	_ = f.Close()
	if err != nil {
		return "", nil, err
	}
	runner := m.Runner
	if runner == nil {
		runner = execRunner{}
	}
	output, err := runner.Run(ctx, "losetup", "--find", "--show", "--partscan", imagePath)
	if err != nil {
		return "", nil, fmt.Errorf("attach image loop device: %w: %s", err, strings.TrimSpace(string(output)))
	}
	loop := strings.TrimSpace(string(output))
	if !loopPath.MatchString(loop) {
		return "", nil, errors.New("invalid loop-device path")
	}
	mountpoint, err := os.MkdirTemp("", "gomultitool-rootfs-")
	if err != nil {
		_, _ = runner.Run(context.Background(), "losetup", "-d", loop)
		return "", nil, err
	}
	mounted, attached := false, true
	cleanup := func() error {
		if mounted {
			if _, err := runner.Run(context.Background(), "umount", mountpoint); err != nil {
				return err
			}
			mounted = false
		}
		if attached {
			if _, err := runner.Run(context.Background(), "losetup", "-d", loop); err != nil {
				return err
			}
			attached = false
		}
		return os.Remove(mountpoint)
	}
	devicePath := loop
	if partition > 0 {
		devicePath += "p" + strconv.Itoa(partition)
	}
	output, err = runner.Run(ctx, "mount", devicePath, mountpoint)
	if err != nil {
		cleanupErr := cleanup()
		return "", nil, errors.Join(fmt.Errorf("mount image root: %w: %s", err, strings.TrimSpace(string(output))), cleanupErr)
	}
	mounted = true
	return mountpoint, cleanup, nil
}

func findExt4Partition(r io.ReaderAt, diskBytes int64) (int, error) {
	if HasExt4Signature(r) {
		return 0, nil
	}
	parts, err := ParseMBR(r, diskBytes)
	if err != nil {
		return 0, errors.New("unsupported image partition table or filesystem")
	}
	for _, p := range parts {
		section := io.NewSectionReader(r, int64(p.StartSector*512), int64(p.SectorCount*512))
		if HasExt4Signature(section) {
			return p.Number, nil
		}
	}
	return 0, errors.New("image has no ext4 root partition")
}
