package tui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/Diaszano/GoMultitool/internal/application/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func freeBytes(context.Context) (int64, error) { return 2e9, nil }

func openBackup(t *testing.T, s Services) (*backupScreen, *appSimulation, i18n.Catalog) {
	t.Helper()
	if s.Inventory == nil {
		s.Inventory = burnInventory
	}
	if s.FreeBytes == nil {
		s.FreeBytes = freeBytes
	}
	catalog := i18n.For(i18n.LocaleEN)
	app, _, _ := newApp(s, catalog)
	b := newBackupScreen(app, s, catalog, func() {})
	sim := startSimulation(t, app)
	app.QueueUpdateDraw(func() { app.SetRoot(b.root, true) })
	b.load()
	if !waitForApp(t, app, func() bool { return simulationContains(sim.screen, "/dev/mmcblk1") }) {
		t.Fatal("sources not shown")
	}
	return b, sim, catalog
}

// backupRun picks the eMMC (second row), sets the name and starts the backup.
func backupRun(t *testing.T, b *backupScreen, sim *appSimulation, name string) {
	pressKey(sim.app, tcell.KeyDown, 0)
	pressKey(sim.app, tcell.KeyEnter, 0)
	waitForApp(t, sim.app, func() bool { return b.target != nil })
	sim.app.QueueUpdateDraw(func() {
		b.name.SetText(name)
		b.start()
	})
}

func TestBackupScreenDefaultsNameWithDate(t *testing.T) {
	b, sim, _ := openBackup(t, Services{})
	want := "tvbox-backup-" + time.Now().Format("2006-01-02")
	if got := b.name.GetText(); got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	if !simulationContains(sim.screen, ".img.gz") || !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "Free on MULTITOOL: 2.0 GB")
	}) {
		t.Fatal("suffix or free space not shown")
	}
}

func TestBackupScreenRejectsBadNames(t *testing.T) {
	var calls atomic.Int32
	s := Services{Backup: func(context.Context, device.Device, string, Progress) error {
		calls.Add(1)
		return nil
	}}
	for _, name := range []string{"", "a/b", ".x"} {
		b, sim, _ := openBackup(t, s)
		backupRun(t, b, sim, name)
		if !waitForApp(
			t,
			sim.app,
			func() bool { return simulationContains(sim.screen, "Invalid name.") },
		) {
			t.Fatalf("name %q not rejected", name)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("Backup called with invalid name")
	}
}

func TestBackupScreenNoSpaceMessage(t *testing.T) {
	b, sim, _ := openBackup(t, Services{Backup: func(
		context.Context, device.Device, string, Progress,
	) error {
		return fmt.Errorf("wrap: %w", backup.ErrNoSpace)
	}})
	backupRun(t, b, sim, "x")
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "No space left on the MULTITOOL partition.")
	}) {
		t.Fatal("no-space text not shown")
	}
	if simulationContains(sim.screen, "Backup saved") {
		t.Fatal("success shown")
	}
}

func TestBackupScreenExistsMessage(t *testing.T) {
	b, sim, _ := openBackup(t, Services{Backup: func(
		context.Context, device.Device, string, Progress,
	) error {
		return fmt.Errorf("%w: /m/x", backup.ErrExists)
	}})
	backupRun(t, b, sim, "x")
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "A backup with this name already exists.")
	}) {
		t.Fatal("exists text not shown")
	}
}

func TestBackupScreenNameWithSpaces(t *testing.T) {
	got := make(chan string, 1)
	b, sim, _ := openBackup(t, Services{Backup: func(
		_ context.Context, _ device.Device, name string, _ Progress,
	) error {
		got <- name
		return nil
	}})
	backupRun(t, b, sim, "tv box")
	select {
	case name := <-got:
		if name != "tv box" {
			t.Fatalf("name = %q", name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Backup not called")
	}
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "Backup saved: tv box.img.gz")
	}) {
		t.Fatal("success not shown")
	}
}

func TestBackupScreenCtrlCCancels(t *testing.T) {
	catalog := i18n.For(i18n.LocaleEN)
	s := Services{
		Inventory: burnInventory, FreeBytes: freeBytes,
		Backup: func(ctx context.Context, _ device.Device, _ string, p Progress) error {
			p(1e9, 4e9)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	app, list, _ := newApp(s, catalog)
	sim := startSimulation(t, app)
	app.QueueUpdateDraw(func() { list.SetCurrentItem(1) })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool { return simulationContains(sim.screen, "/dev/mmcblk1") }) {
		t.Fatal("backup screen not shown")
	}
	pressKey(app, tcell.KeyDown, 0)
	pressKey(app, tcell.KeyEnter, 0)
	pressKey(app, tcell.KeyTab, 0)
	pressKey(app, tcell.KeyTab, 0)
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool { return simulationContains(sim.screen, "25%") }) {
		t.Fatal("progress not shown")
	}
	pressKey(app, tcell.KeyCtrlC, 0)
	if !waitForApp(t, app, func() bool {
		return simulationContains(sim.screen, catalog.Backup.Cancelled)
	}) {
		t.Fatal("cancelled text not shown")
	}
	if simulationContains(sim.screen, "Backup saved") {
		t.Fatal("success shown after cancel")
	}
	select {
	case <-sim.done:
		t.Fatal("Ctrl-C stopped the app mid-backup")
	default:
	}
	pressKey(app, tcell.KeyCtrlC, 0)
	waitForDone(t, sim)
}

func TestBackupScreenMarksSelectedSource(t *testing.T) {
	b, sim, _ := openBackup(t, Services{})
	pressKey(sim.app, tcell.KeyDown, 0)
	pressKey(sim.app, tcell.KeyEnter, 0)
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "▶ ") && b.target != nil
	}) {
		t.Fatal("selected source not marked")
	}
	if got, _ := b.list.GetItemText(0); strings.Contains(got, "▶") {
		t.Fatalf("wrong row marked: %q", got)
	}
}
