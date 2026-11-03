package linux

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// HasZRAMManager reports whether a known ZRAM service is installed.
func HasZRAMManager(root string) bool {
	for _, name := range []string{"zramswap.service", "systemd-zram-setup@.service", "zram-config.service"} {
		for _, dir := range []string{"etc/systemd/system", "lib/systemd/system", "usr/lib/systemd/system"} {
			if info, err := os.Stat(
				root + "/" + dir + "/" + name,
			); err == nil &&
				info.Mode().IsRegular() {
				return true
			}
		}
	}
	return false
}

// HasSwapConfiguration reports existing fstab or systemd swap-file entries.
func HasSwapConfiguration(root string) bool {
	for _, path := range []string{"etc/fstab", "etc/systemd/system/swapfile.swap", "lib/systemd/system/swapfile.swap", "usr/lib/systemd/system/swapfile.swap"} {
		f, err := os.Open(root + "/" + path)
		if err != nil {
			continue
		}
		s := bufio.NewScanner(f)
		found := false
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if strings.HasPrefix(line, "/swapfile ") || strings.HasPrefix(line, "What=/swapfile") {
				found = true
				break
			}
		}
		_ = f.Close()
		if found {
			return true
		}
	}
	return false
}

// ReadSwapPriorities returns active swap device priorities in procfs order.
func ReadSwapPriorities(path string) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	if s.Scan() { // header
		var priorities []int
		for s.Scan() {
			fields := strings.Fields(s.Text())
			if len(fields) < 5 {
				continue
			}
			priority, err := strconv.Atoi(fields[4])
			if err != nil {
				continue
			}
			priorities = append(priorities, priority)
		}
		return priorities, s.Err()
	}
	return nil, s.Err()
}
