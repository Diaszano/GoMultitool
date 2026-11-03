package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func TestMenuShowsAllAreasAndOnlyHardwareRuns(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		t.Run(string(locale), func(t *testing.T) {
			calls := 0
			app, list, _ := newApp(func() ([]device.Device, error) {
				calls++
				return []device.Device{{ID: "mmc0", Path: "/dev/mmc0", SizeBytes: 512}}, nil
			}, i18n.For(locale))
			if list.GetItemCount() != 10 {
				t.Fatalf("menu entries = %d, want 10", list.GetItemCount())
			}
			for i, item := range i18n.For(locale).Menu {
				label, hint := list.GetItemText(i)
				if label != item.Label || hint != item.Hint {
					t.Errorf(
						"menu item %d = (%q, %q), want (%q, %q)",
						i,
						label,
						hint,
						item.Label,
						item.Hint,
					)
				}
			}
			for i := 1; i < 9; i++ {
				if callback := list.GetItemSelectedFunc(i); callback != nil {
					callback()
				}
			}
			if calls != 0 {
				t.Fatalf("future menu items invoked hardware %d times", calls)
			}
			list.GetItemSelectedFunc(0)()
			if calls != 1 {
				t.Fatalf("hardware action calls = %d, want 1", calls)
			}
			if app == nil {
				t.Fatal("tview application is nil")
			}
		})
	}
}

func TestTUIKeyboardNavigationBackAndQuit(t *testing.T) {
	calls := 0
	app, list, hardware := newApp(func() ([]device.Device, error) {
		calls++
		return []device.Device{{ID: "mmc0", Path: "/dev/mmc0"}}, nil
	}, i18n.For(i18n.LocaleEN))
	simulation := startSimulation(t, app)
	if !simulationContains(simulation.screen, "GoMultitool") {
		t.Fatal("home screen was not drawn")
	}

	pressKey(app, tcell.KeyDown, 0)
	waitForApp(t, app, func() bool { return list.GetCurrentItem() == 1 })
	pressKey(app, tcell.KeyEnter, 0)
	pressKey(app, tcell.KeyDown, 0)
	waitForApp(t, app, func() bool { return list.GetCurrentItem() == 2 })
	if !waitForApp(t, app, func() bool {
		return calls == 0 && simulationContains(simulation.screen, "GoMultitool")
	}) {
		t.Fatalf("future item changed screen or invoked hardware: calls=%d", calls)
	}

	pressKey(app, tcell.KeyHome, 0)
	waitForApp(t, app, func() bool { return list.GetCurrentItem() == 0 })
	pressKey(app, tcell.KeyEnter, 0)
	waitForApp(t, app, func() bool {
		return calls == 1 && strings.Contains(hardware.GetText(false), "/dev/mmc0")
	})
	if !waitForApp(t, app, func() bool {
		return simulationContains(simulation.screen, "Storage devices")
	}) {
		t.Fatal("hardware screen was not drawn")
	}

	pressKey(app, tcell.KeyEscape, 0)
	waitForApp(t, app, func() bool {
		return simulationContains(simulation.screen, "GoMultitool") &&
			!simulationContains(simulation.screen, "Storage devices")
	})
	pressKey(app, tcell.KeyRune, 'q')
	waitForDone(t, simulation)
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
			app, list, _ := newApp(nil, i18n.For(i18n.LocaleEN))
			simulation := startSimulation(t, app)
			if tt.exit {
				app.QueueUpdateDraw(func() { list.SetCurrentItem(9) })
			}
			pressKey(app, tt.key, tt.rune)
			waitForDone(t, simulation)
		})
	}
}

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
			contains: "/dev/mmc0",
		},
		{name: "empty", contains: "Nenhum dispositivo encontrado."},
		{name: "error", err: errors.New("sysfs unavailable"), contains: "sysfs unavailable"},
	}
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		catalog := i18n.For(locale)
		for _, tt := range tests {
			t.Run(string(locale)+"/"+tt.name, func(t *testing.T) {
				_, list, hardware := newApp(
					func() ([]device.Device, error) { return tt.devices, tt.err },
					catalog,
				)
				list.GetItemSelectedFunc(0)()
				got := hardware.GetText(false)
				want := tt.contains
				if tt.name == "empty" {
					want = catalog.EmptyInventory
				}
				if !strings.Contains(got, want) {
					t.Fatalf("hardware text = %q, want it to contain %q", got, want)
				}
				if tt.name == "empty" && got != catalog.EmptyInventory {
					t.Errorf("empty text = %q, want %q", got, catalog.EmptyInventory)
				}
				if tt.err != nil && got != catalog.ErrorPrefix+tt.err.Error() {
					t.Errorf("error text = %q, want localized context and raw error", got)
				}
				if tt.name == "empty" && got == catalog.ErrorPrefix+"sysfs unavailable" {
					t.Fatal("empty inventory and error states were not distinguished")
				}
			})
		}
	}
}
