// Package diagnostics reports verified observations about boot assets.
package diagnostics

import (
	"os"
	"path/filepath"
)

// BootReport lists observed boot assets and missing expected files.
type BootReport struct {
	ExtlinuxConfigPresent bool
	KernelPresent         bool
	Observations          []string
	Missing               []string
	Hypotheses            []string `json:"hypotheses,omitempty"`
	JournalAvailable      bool     `json:"journal_available"`
}

// InspectBootAssets checks known kernel and extlinux filenames in a mounted tree.
func InspectBootAssets(root string) BootReport {
	r := BootReport{}
	if regularFile(filepath.Join(root, "extlinux", "extlinux.conf")) {
		r.ExtlinuxConfigPresent = true
		r.Observations = append(r.Observations, "extlinux/extlinux.conf is present")
	} else {
		r.Missing = append(r.Missing, "extlinux/extlinux.conf")
	}
	for _, name := range []string{"Image", "zImage", "kernel.img"} {
		if regularFile(filepath.Join(root, name)) {
			r.KernelPresent = true
			r.Observations = append(r.Observations, name+" is present")
			break
		}
	}
	if !r.KernelPresent {
		r.Missing = append(r.Missing, "kernel image (Image, zImage, or kernel.img)")
	}
	return r
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
