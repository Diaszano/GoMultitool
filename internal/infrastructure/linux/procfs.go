package linux

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var partitionSuffix = regexp.MustCompile(`(?:p)?[0-9]+$`)

// hostState is one read of the host's mount, swap and boot-root state.
type hostState struct {
	mounts, swaps       []string
	swapsErr, mountsErr bool
	bootRoot            string
	bootKnown           bool
}

// snapshot reads procfs and resolves the boot root once per List.
func (c SysfsCatalog) snapshot(ctx context.Context) hostState {
	var s hostState
	if b, err := os.ReadFile(filepath.Join(c.ProcRoot, "mounts")); err != nil {
		s.mountsErr = true
	} else {
		s.mounts = firstFields(string(b), 0)
	}
	if b, err := os.ReadFile(filepath.Join(c.ProcRoot, "swaps")); err != nil {
		s.swapsErr = !os.IsNotExist(err)
	} else {
		s.swaps = firstFields(string(b), 1)
	}
	s.bootRoot, s.bootKnown = c.resolveBootRoot(ctx)
	return s
}

func firstFields(data string, skip int) []string {
	var out []string
	for i, line := range strings.Split(data, "\n") {
		if fields := strings.Fields(line); i >= skip && len(fields) > 0 {
			out = append(out, fields[0])
		}
	}
	return out
}

// isBusy reports mounted filesystems, active swap, and kernel holders such as
// device-mapper or MD on the device or any of its partitions.
func (c SysfsCatalog) isBusy(name, path string, s hostState) bool {
	if s.mountsErr || s.swapsErr {
		return true
	}
	for _, src := range append(append([]string{}, s.mounts...), s.swaps...) {
		if sameDeviceOrPartition(src, path) {
			return true
		}
	}
	root := filepath.Join(c.SysRoot, "block", name)
	partitions, _ := filepath.Glob(filepath.Join(root, name+"*", "holders"))
	for _, dir := range append(partitions, filepath.Join(root, "holders")) {
		entries, err := os.ReadDir(dir)
		if (err != nil && !os.IsNotExist(err)) || len(entries) > 0 {
			return true
		}
	}
	return false
}

func (c SysfsCatalog) isBootMedia(path string, s hostState) bool {
	return !s.bootKnown || sameDeviceOrPartition(s.bootRoot, path)
}

// resolveBootRoot returns the root device path; false means it is unknown.
func (c SysfsCatalog) resolveBootRoot(ctx context.Context) (string, bool) {
	b, err := os.ReadFile(filepath.Join(c.ProcRoot, "cmdline"))
	if err != nil {
		return "", false
	}
	for _, field := range strings.Fields(string(b)) {
		if !strings.HasPrefix(field, "root=") {
			continue
		}
		root := strings.TrimPrefix(field, "root=")
		if strings.HasPrefix(root, "/dev/") {
			return root, true
		}
		parts := strings.SplitN(root, "=", 2)
		if len(parts) != 2 ||
			!map[string]bool{"UUID": true, "PARTUUID": true, "LABEL": true, "PARTLABEL": true}[parts[0]] {
			return "", false
		}
		devRoot := c.DevRoot
		if devRoot == "" {
			devRoot = "/dev"
		}
		link, err := filepath.EvalSymlinks(
			filepath.Join(devRoot, "disk", strings.ToLower(parts[0]), parts[1]),
		)
		if err == nil {
			return link, true
		}
		timeout := c.BlkidTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		matches, err := exec.CommandContext(
			ctx, "blkid", "-t", parts[0]+"="+parts[1], "-o", "device",
		).Output()
		if err != nil {
			return "", false
		}
		devices := strings.Fields(string(matches))
		if len(devices) != 1 {
			return "", false
		}
		return devices[0], true
	}
	return "", false
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
