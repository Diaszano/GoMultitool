package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/infrastructure/multitool"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func TestMenuShowsFiveItems(t *testing.T) {
	want := map[i18n.Locale][5][2]string{
		i18n.LocalePtBR: {
			{"Gravar imagem no eMMC", "Instalar Armbian ou outra imagem"},
			{"Fazer backup", "Salvar o eMMC na partição MULTITOOL"},
			{"Informações de hardware", "Ver dispositivos de armazenamento conectados"},
			{"Menu antigo (Bash)", "Restaurar, apagar, steP-nand"},
			{"Sair", "Fechar o GoMultitool"},
		},
		i18n.LocaleEN: {
			{"Burn image to eMMC", "Install Armbian or another image"},
			{"Back up", "Save the eMMC to the MULTITOOL partition"},
			{"Hardware information", "Inspect connected storage devices"},
			{"Legacy menu (Bash)", "Restore, erase, steP-nand"},
			{"Exit", "Leave GoMultitool"},
		},
	}
	for locale, items := range want {
		t.Run(string(locale), func(t *testing.T) {
			_, list, _ := newApp(Services{}, i18n.For(locale))
			if list.GetItemCount() != 5 {
				t.Fatalf("menu entries = %d, want 5", list.GetItemCount())
			}
			for i, item := range items {
				label, hint := list.GetItemText(i)
				if label != item[0] || hint != item[1] {
					t.Errorf("item %d = (%q, %q), want (%q, %q)", i, label, hint, item[0], item[1])
				}
			}
		})
	}
}

func TestHardwareLoadsAsyncAndStaysResponsive(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	services := Services{Inventory: func(context.Context) ([]device.Device, error) {
		<-release
		return nil, nil
	}}
	app, list, _ := newApp(services, i18n.For(i18n.LocalePtBR))
	simulation := startSimulation(t, app)
	app.QueueUpdateDraw(func() { list.SetCurrentItem(hardwareIndex) })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool {
		return simulationContains(simulation.screen, "Carregando…")
	}) {
		t.Fatal("loading text was not shown while the inventory blocks")
	}
	pressKey(app, tcell.KeyRune, 'q')
	waitForDone(t, simulation)
}

func TestLegacyItemRunsBash(t *testing.T) {
	calls := 0
	app, list, _ := newApp(Services{RunLegacy: func() error {
		calls++
		return nil
	}}, i18n.For(i18n.LocaleEN))
	simulation := startSimulation(t, app)
	app.QueueUpdateDraw(func() { list.SetCurrentItem(legacyIndex) })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool {
		return calls == 1 && simulationContains(simulation.screen, "Legacy menu (Bash)")
	}) {
		t.Fatalf("legacy calls = %d, want 1 and the menu drawn again", calls)
	}
}

func TestLegacyErrorShowsModal(t *testing.T) {
	app, list, _ := newApp(Services{RunLegacy: func() error {
		return errors.New("multitool.sh missing")
	}}, i18n.For(i18n.LocaleEN))
	simulation := startSimulation(t, app)
	app.QueueUpdateDraw(func() { list.SetCurrentItem(legacyIndex) })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool {
		// The modal wraps the text, so match the start of it.
		return simulationContains(simulation.screen, "Error: multitool.sh")
	}) {
		t.Fatal("legacy error was not shown")
	}
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool {
		return !simulationContains(simulation.screen, "Error: multitool.sh") &&
			simulationContains(simulation.screen, "Hardware information")
	}) {
		t.Fatal("dismissing the modal did not return to the menu")
	}
}

func TestTUIKeyboardNavigationBackAndQuit(t *testing.T) {
	var calls atomic.Int32
	app, list, hardware := newApp(Services{
		Inventory: func(context.Context) ([]device.Device, error) {
			calls.Add(1)
			return []device.Device{{ID: "mmc0", Path: "/dev/mmc0"}}, nil
		},
	}, i18n.For(i18n.LocaleEN))
	simulation := startSimulation(t, app)
	if !simulationContains(simulation.screen, "GoMultitool") {
		t.Fatal("home screen was not drawn")
	}

	pressKey(app, tcell.KeyDown, 0)
	waitForApp(t, app, func() bool { return list.GetCurrentItem() == 1 })
	pressKey(app, tcell.KeyDown, 0)
	waitForApp(t, app, func() bool { return list.GetCurrentItem() == hardwareIndex })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool {
		return calls.Load() == 1 && strings.Contains(hardware.GetText(false), "/dev/mmc0") &&
			simulationContains(simulation.screen, "Storage devices")
	}) {
		t.Fatal("hardware screen was not drawn")
	}

	pressKey(app, tcell.KeyEscape, 0)
	if !waitForApp(t, app, func() bool {
		return simulationContains(simulation.screen, "GoMultitool") &&
			!simulationContains(simulation.screen, "Storage devices")
	}) {
		t.Fatal("Esc did not return to the menu")
	}
	pressKey(app, tcell.KeyRune, 'q')
	waitForDone(t, simulation)
}

func TestBackupOpensPlaceholderAndEscReturns(t *testing.T) {
	for _, index := range []int{1} {
		catalog := i18n.For(i18n.LocaleEN)
		app, list, _ := newApp(Services{}, catalog)
		simulation := startSimulation(t, app)
		app.QueueUpdateDraw(func() { list.SetCurrentItem(index) })
		pressKey(app, tcell.KeyEnter, 0)
		if !waitForApp(t, app, func() bool {
			return !simulationContains(simulation.screen, "Hardware information")
		}) {
			t.Fatalf("item %d did not leave the menu", index)
		}
		pressKey(app, tcell.KeyEscape, 0)
		if !waitForApp(t, app, func() bool {
			return simulationContains(simulation.screen, "Hardware information")
		}) {
			t.Fatalf("Esc on item %d did not return to the menu", index)
		}
	}
}

func TestTUIExitAndCtrlCStopApplication(t *testing.T) {
	tests := []struct {
		name string
		key  tcell.Key
		rune rune
		exit bool
	}{
		{name: "exit item", key: tcell.KeyEnter, exit: true},
		{name: "ctrl-c", key: tcell.KeyCtrlC},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, list, _ := newApp(Services{}, i18n.For(i18n.LocaleEN))
			simulation := startSimulation(t, app)
			if tt.exit {
				app.QueueUpdateDraw(func() { list.SetCurrentItem(exitIndex) })
			}
			pressKey(app, tt.key, tt.rune)
			waitForDone(t, simulation)
		})
	}
}

const (
	hardwareIndex = 2
	legacyIndex   = 3
	exitIndex     = 4
)

type appSimulation struct {
	app    *tview.Application
	screen tcell.SimulationScreen
	done   chan struct{}
	err    error
}

func startSimulation(t *testing.T, app *tview.Application) *appSimulation {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(80, 24)
	app.SetScreen(screen)
	simulation := &appSimulation{app: app, screen: screen, done: make(chan struct{})}
	go func() {
		simulation.err = app.Run()
		close(simulation.done)
	}()
	app.QueueUpdateDraw(func() {})
	t.Cleanup(func() {
		select {
		case <-simulation.done:
		default:
			app.Stop()
			waitForDone(t, simulation)
		}
	})
	return simulation
}

func pressKey(app *tview.Application, key tcell.Key, r rune) {
	app.QueueEvent(tcell.NewEventKey(key, r, tcell.ModNone))
}

func simulationContains(screen tcell.SimulationScreen, text string) bool {
	cells, _, _ := screen.GetContents()
	var output strings.Builder
	for _, cell := range cells {
		output.WriteString(string(cell.Runes))
	}
	return strings.Contains(output.String(), text)
}

func waitForApp(t *testing.T, app *tview.Application, condition func() bool) bool {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var met bool
		app.QueueUpdateDraw(func() { met = condition() })
		if met {
			return true
		}
		select {
		case <-timer.C:
			return false
		case <-ticker.C:
		}
	}
}

func waitForDone(t *testing.T, simulation *appSimulation) {
	t.Helper()
	select {
	case <-simulation.done:
		if simulation.err != nil {
			t.Fatal(simulation.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tview application did not stop")
	}
}

func TestHardwareScreenShowsInventoryEmptyAndError(t *testing.T) {
	tests := []struct {
		name     string
		devices  []device.Device
		err      error
		contains string
	}{
		{
			name: "inventory",
			devices: []device.Device{{
				ID: "mmc0", Path: "/dev/mmc0", Kind: device.KindEMMC, SizeBytes: 512,
			}},
			contains: "/dev/mmc0  eMMC  0 MB",
		},
		{name: "empty"},
		{name: "error", err: errors.New("sysfs unavailable")},
	}
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		catalog := i18n.For(locale)
		for _, tt := range tests {
			t.Run(string(locale)+"/"+tt.name, func(t *testing.T) {
				app, list, hardware := newApp(Services{
					Inventory: func(context.Context) ([]device.Device, error) {
						return tt.devices, tt.err
					},
				}, catalog)
				startSimulation(t, app)
				app.QueueUpdateDraw(func() { list.SetCurrentItem(hardwareIndex) })
				pressKey(app, tcell.KeyEnter, 0)
				want := tt.contains
				switch tt.name {
				case "empty":
					want = catalog.EmptyInventory
				case "error":
					want = catalog.ErrorPrefix + tt.err.Error()
				}
				var got string
				if !waitForApp(t, app, func() bool {
					got = hardware.GetText(false)
					return strings.Contains(got, want)
				}) {
					t.Fatalf("hardware text = %q, want it to contain %q", got, want)
				}
				if tt.name != "inventory" && got != want {
					t.Errorf("text = %q, want exactly %q", got, want)
				}
			})
		}
	}
}

func TestErrorTextLocalizesBusy(t *testing.T) {
	catalog := i18n.For(i18n.LocalePtBR)
	if got := errorText(catalog, fmt.Errorf("lock: %w", multitool.ErrBusy)); got != catalog.Busy {
		t.Fatalf("busy text = %q, want %q", got, catalog.Busy)
	}
}
