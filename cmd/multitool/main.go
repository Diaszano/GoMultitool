// Package main provides the multitool command-line interface.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/discovery"
	"github.com/Diaszano/GoMultitool/internal/buildinfo"
	linuxsys "github.com/Diaszano/GoMultitool/internal/infrastructure/linux"
)

func main() {
	catalog := linuxsys.SysfsCatalog{SysRoot: "/sys", ProcRoot: "/proc"}
	os.Exit(run(os.Args[1:], catalog, os.Stdout, os.Stderr))
}

func run(args []string, catalog application.DeviceCatalog, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "version" {
		info := buildinfo.Current()
		if _, err := fmt.Fprintf(stdout, "RK322x Multitool %s\nCommit: %s\nBuild: %s\n", info.Version, info.Commit, info.BuildTime); err != nil {
			return 1
		}
		return 0
	}
	if len(args) == 2 && args[0] == "hardware" && args[1] == "--json" {
		devices, err := (discovery.Service{Catalog: catalog}).Inventory(context.Background())
		if err != nil {
			_ = json.NewEncoder(stderr).Encode(map[string]string{"error": err.Error()})
			return 1
		}
		if err := json.NewEncoder(stdout).Encode(devices); err != nil {
			_, _ = fmt.Fprintf(stderr, "encode hardware inventory: %v\n", err)
			return 1
		}
		return 0
	}
	_, _ = fmt.Fprintln(stderr, "Usage: multitool version | multitool hardware --json")
	return 2
}
