// Package main provides the multitool command-line interface.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/diagnostics"
	"github.com/Diaszano/GoMultitool/internal/application/discovery"
	"github.com/Diaszano/GoMultitool/internal/buildinfo"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	linuxsys "github.com/Diaszano/GoMultitool/internal/infrastructure/linux"
	"github.com/Diaszano/GoMultitool/internal/presentation/console"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui"
)

func main() {
	devices := linuxsys.SysfsCatalog{SysRoot: "/sys", ProcRoot: "/proc"}
	messages := i18n.For(i18n.DetectLocale(os.Getenv))
	startTUI := func() error {
		service := discovery.Service{Catalog: devices}
		return tui.New(
			func() ([]device.Device, error) { return service.Inventory(context.Background()) },
			messages,
		).Run()
	}
	os.Exit(run(os.Args[1:], "/", devices, messages, os.Stdin, os.Stdout, os.Stderr, startTUI))
}

func run(
	args []string,
	root string,
	devices application.DeviceCatalog,
	messages i18n.Catalog,
	stdin io.Reader,
	stdout, stderr io.Writer,
	startTUI func() error,
) int {
	if len(args) == 1 && args[0] == "tui" {
		if err := startTUI(); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) == 1 && args[0] == "--console" {
		service := discovery.Service{Catalog: devices}
		if err := console.Run(stdin, stdout, service.Inventory, messages); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) == 1 && args[0] == "version" {
		info := buildinfo.Current()
		if _, err := fmt.Fprintf(
			stdout,
			"%s %s\n%s: %s\n%s: %s\n",
			messages.ProductName,
			info.Version,
			messages.CommitLabel,
			info.Commit,
			messages.BuildLabel,
			info.BuildTime,
		); err != nil {
			return 1
		}
		return 0
	}
	if len(args) == 2 && args[0] == "hardware" && args[1] == "--json" {
		inventory, err := (discovery.Service{Catalog: devices}).Inventory(context.Background())
		if err != nil {
			_ = json.NewEncoder(stderr).Encode(map[string]string{"error": err.Error()})
			return 1
		}
		return encodeJSON(stdout, stderr, "encode hardware inventory", inventory)
	}
	if len(args) == 2 && args[0] == "diagnose" && args[1] == "--json" {
		cmdline, err := os.ReadFile(filepath.Join(root, "proc/cmdline"))
		if err != nil {
			_ = json.NewEncoder(stderr).Encode(map[string]string{"error": err.Error()})
			return 1
		}
		boot := diagnostics.CollectBootReport(root, string(cmdline))
		network := diagnostics.CollectNetworkReport()
		report := struct {
			Boot    diagnostics.BootReport    `json:"boot"`
			Network diagnostics.NetworkReport `json:"network"`
		}{boot, network}
		return encodeJSON(stdout, stderr, "encode diagnostics", report)
	}
	_, _ = fmt.Fprintln(stderr, messages.Usage)
	return 2
}

func encodeJSON(stdout, stderr io.Writer, context string, value any) int {
	if err := json.NewEncoder(stdout).Encode(value); err != nil {
		_, _ = fmt.Fprintf(stderr, "%s: %v\n", context, err)
		return 1
	}
	return 0
}
