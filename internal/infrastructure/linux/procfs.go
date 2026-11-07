package linux

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var partitionSuffix = regexp.MustCompile(`(?:p)?[0-9]+$`)

// isBusy reports mounted filesystems, active swap, and kernel holders such as
// device-mapper or MD on the device or any of its partitions.
func (c SysfsCatalog) isBusy(name, path string) bool {
	b, err := os.ReadFile(filepath.Join(c.ProcRoot, "mounts"))
	if err != nil {
		return true
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && sameDeviceOrPartition(fields[0], path) {
			return true
		}
	}
	swaps, err := os.ReadFile(filepath.Join(c.ProcRoot, "swaps"))
	if err != nil && !os.IsNotExist(err) {
		return true
	}
	for _, line := range strings.Split(string(swaps), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) > 0 && sameDeviceOrPartition(fields[0], path) {
			return true
		}
	}
	root := filepath.Join(c.SysRoot, "block", name)
	partitions, _ := filepath.Glob(filepath.Join(root, name+"*", "holders"))
	for _, dir := range append(partitions, filepath.Join(root, "holders")) {
		if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
			return true
		}
	}
	return false
}

func (c SysfsCatalog) isBootMedia(path string) bool {
	b, err := os.ReadFile(filepath.Join(c.ProcRoot, "cmdline"))
	if err != nil {
		return true
	}
	for _, field := range strings.Fields(string(b)) {
		if !strings.HasPrefix(field, "root=") {
			continue
		}
		root := strings.TrimPrefix(field, "root=")
		if strings.HasPrefix(root, "/dev/") {
			return sameDeviceOrPartition(root, path)
		}
		parts := strings.SplitN(root, "=", 2)
		if len(parts) != 2 ||
			!map[string]bool{"UUID": true, "PARTUUID": true, "LABEL": true, "PARTLABEL": true}[parts[0]] {
			return true
		}
		devRoot := c.DevRoot
		if devRoot == "" {
			devRoot = "/dev"
		}
		link, err := filepath.EvalSymlinks(
			filepath.Join(devRoot, "disk", strings.ToLower(parts[0]), parts[1]),
		)
		if err != nil {
			matches, err := exec.Command(
				"blkid",
				"-t",
				parts[0]+"="+parts[1],
				"-o",
				"device",
			).Output()
			if err != nil {
				return true
			}
			devices := strings.Fields(string(matches))
			if len(devices) != 1 {
				return true
			}
			link = devices[0]
		}
		return sameDeviceOrPartition(link, path)
	}
	return true
}

func sameDeviceOrPartition(source, device string) bool {
	if source == device {
		return true
	}
	name, parent := filepath.Base(source), filepath.Base(device)
	if !strings.HasPrefix(name, parent) {
		return false
	}
	suffix := strings.TrimPrefix(name, parent)
	if strings.HasPrefix(parent, "mmcblk") || strings.HasPrefix(parent, "nvme") {
		suffix = strings.TrimPrefix(suffix, "p")
	}
	return suffix != "" && partitionSuffix.MatchString(suffix) &&
		strings.Trim(suffix, "0123456789") == ""
}
