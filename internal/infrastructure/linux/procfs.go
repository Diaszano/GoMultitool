package linux

import (
	"os"
	"path/filepath"
	"strings"
)

func (c SysfsCatalog) isMounted(path string) bool {
	b, err := os.ReadFile(filepath.Join(c.ProcRoot, "mounts"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == path {
			return true
		}
	}
	return false
}
