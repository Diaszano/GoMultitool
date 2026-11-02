package diagnostics

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// CollectBootReport records boot observations without changing the target.
func CollectBootReport(root, cmdline string) BootReport {
	r := BootReport{}
	cmdRoot := ""
	for _, field := range strings.Fields(cmdline) {
		if strings.HasPrefix(field, "root=") {
			cmdRoot = strings.TrimPrefix(field, "root=")
			break
		}
	}
	fstabPath := filepath.Join(root, "etc/fstab")
	f, err := os.Open(fstabPath)
	if err != nil {
		r.Missing = append(r.Missing, "/etc/fstab")
	} else {
		scanner := bufio.NewScanner(f)
		foundRoot := false
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 2 || strings.HasPrefix(fields[0], "#") || fields[1] != "/" {
				continue
			}
			foundRoot = true
			r.Observations = append(r.Observations, "fstab root source: "+fields[0])
			if cmdRoot != "" && fields[0] != cmdRoot {
				r.Hypotheses = append(r.Hypotheses, "kernel root= and fstab root source differ")
			}
		}
		if scanner.Err() != nil {
			r.Missing = append(r.Missing, "/etc/fstab unreadable")
		}
		if !foundRoot {
			r.Missing = append(r.Missing, "root filesystem entry in /etc/fstab")
		}
		_ = f.Close()
	}
	boot := InspectBootAssets(filepath.Join(root, "boot"))
	r.Missing = append(r.Missing, boot.Missing...)
	for _, item := range boot.Observations {
		r.Observations = append(r.Observations, "/boot: "+item)
	}
	if info, err := os.Stat(filepath.Join(root, "var/log/journal")); err == nil && info.IsDir() {
		r.JournalAvailable = true
	} else {
		r.Missing = append(r.Missing, "persistent journal directory")
	}
	return r
}
