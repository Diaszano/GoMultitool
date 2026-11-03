// Package main provides the multitool command-line interface.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

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
	if len(os.Args) == 2 && os.Args[1] == "tui" {
		service := discovery.Service{Catalog: devices}
		app := tui.New(
			func() ([]device.Device, error) { return service.Inventory(context.Background()) },
			messages,
		)
		err := app.Run()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--console" {
		service := discovery.Service{Catalog: devices}
		if err := console.Run(os.Stdin, os.Stdout, service.Inventory, messages); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	os.Exit(run(os.Args[1:], devices, messages, os.Stdout, os.Stderr))
}

func run(
	args []string,
	devices application.DeviceCatalog,
	messages i18n.Catalog,
	stdout, stderr io.Writer,
) int {
	if len(args) == 1 && args[0] == "version" {
		info := buildinfo.Current()
		if _, err := fmt.Fprintf(
			stdout,
			"%s %s\n%s: %s\n%s: %s\n",
			messages.VersionLabel,
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
