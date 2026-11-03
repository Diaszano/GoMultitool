// Package linux reads hardware metadata from Linux sysfs and procfs.
package linux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// SysfsCatalog reads block-device metadata from sysfs and mount state from procfs.
type SysfsCatalog struct {
	SysRoot  string
	ProcRoot string
	DevRoot  string
}

// List returns the block devices visible in the configured sysfs tree.
func (c SysfsCatalog) List(ctx context.Context) ([]device.Device, error) {
	entries, err := os.ReadDir(filepath.Join(c.SysRoot, "block"))
	if err != nil {
		return nil, err
	}
	var devices []device.Device
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if strings.HasPrefix(entry.Name(), "loop") || strings.HasPrefix(entry.Name(), "ram") {
			continue
		}
		d, err := c.readDevice(entry.Name())
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	return devices, nil
}

// Resolve returns a discovered device for its /dev path.
func (c SysfsCatalog) Resolve(ctx context.Context, path string) (device.Device, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || filepath.Dir(path) != "/dev" {
		return device.Device{}, fmt.Errorf("invalid device path %q", path)
	}
	devices, err := c.List(ctx)
	if err != nil {
		return device.Device{}, err
	}
	for _, d := range devices {
		if d.Path == path {
			return d, nil
		}
	}
	return device.Device{}, fmt.Errorf("device %q not found", path)
}

func (c SysfsCatalog) readDevice(name string) (device.Device, error) {
	root := filepath.Join(c.SysRoot, "block", name)
	sectors, err := readInt(filepath.Join(root, "size"))
	if err != nil {
		return device.Device{}, fmt.Errorf("read %s size: %w", name, err)
	}
	sectorSize, err := readInt(filepath.Join(root, "queue", "logical_block_size"))
	if err != nil {
		return device.Device{}, fmt.Errorf("read %s logical sector size: %w", name, err)
	}
	removable, _ := os.ReadFile(filepath.Join(root, "removable"))
	d := device.Device{
		ID:                name,
		Path:              filepath.Join("/dev", name),
		SizeBytes:         sectors * 512,
		LogicalSectorSize: sectorSize,
		Removable:         strings.TrimSpace(string(removable)) == "1",
	}
	d.Kind = c.classify(root, d.Removable)
	d.Mounted = c.isMounted(d.Path)
	d.BootMedia = c.isBootMedia(d.Path)
	return d, nil
}

func (c SysfsCatalog) classify(root string, removable bool) device.Kind {
	if typ, err := os.ReadFile(filepath.Join(root, "device", "type")); err == nil {
		switch strings.TrimSpace(string(typ)) {
		case "MMC":
			if !removable {
				return device.KindEMMC
			}
		case "SD":
			return device.KindSD
		}
	}
	transport, _ := os.ReadFile(filepath.Join(root, "device", "transport"))
	if strings.EqualFold(strings.TrimSpace(string(transport)), "usb") {
		return device.KindUSB
	}
	return device.KindUnknown
}

func readInt(path string) (int64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
}
