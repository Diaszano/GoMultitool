// Package main provides the multitool command-line interface.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/diagnostics"
	"github.com/Diaszano/GoMultitool/internal/application/discovery"
	"github.com/Diaszano/GoMultitool/internal/buildinfo"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	linuxsys "github.com/Diaszano/GoMultitool/internal/infrastructure/linux"
	"github.com/Diaszano/GoMultitool/internal/presentation/console"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui"
)

func main() {
	catalog := linuxsys.SysfsCatalog{SysRoot: "/sys", ProcRoot: "/proc"}
	if len(os.Args) == 2 && os.Args[1] == "tui" {
		service := discovery.Service{Catalog: catalog}
		model := tui.NewModel(func() ([]device.Device, error) { return service.Inventory(context.Background()) })
		_, err := tea.NewProgram(model).Run()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--console" {
		service := discovery.Service{Catalog: catalog}
		if err := console.Run(os.Stdin, os.Stdout, service.Inventory); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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
	if len(args) == 2 && args[0] == "diagnose" && args[1] == "--json" {
		cmdline, err := os.ReadFile("/proc/cmdline")
		if err != nil {
			_ = json.NewEncoder(stderr).Encode(map[string]string{"error": err.Error()})
			return 1
		}
		boot := diagnostics.CollectBootReport("/", string(cmdline))
		network := diagnostics.CollectNetworkReport()
		_, available, err := linuxsys.ReadJournal("/var/log/syslog", 64*1024)
		if err != nil {
			_ = json.NewEncoder(stderr).Encode(map[string]string{"error": err.Error()})
			return 1
		}
		report := struct {
			Boot             diagnostics.BootReport    `json:"boot"`
			Network          diagnostics.NetworkReport `json:"network"`
			JournalAvailable bool                      `json:"journal_available"`
		}{boot, network, available}
		if err := json.NewEncoder(stdout).Encode(report); err != nil {
			_, _ = fmt.Fprintf(stderr, "encode diagnostics: %v\n", err)
			return 1
		}
		return 0
	}
	_, _ = fmt.Fprintln(stderr, "Usage: multitool version | multitool hardware --json | multitool diagnose --json | multitool tui | multitool --console")
	return 2
}
