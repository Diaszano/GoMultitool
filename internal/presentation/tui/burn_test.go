package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/Diaszano/GoMultitool/internal/application/installer"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

var burnImages = func(context.Context) ([]image.Image, error) {
	return []image.Image{{Path: "/m/images/armbian.img.xz", SizeBytes: 0}}, nil
}

func burnInventory(context.Context) ([]device.Device, error) {
	return []device.Device{
		{ID: "a", Path: "/dev/mmcblk0", Kind: device.KindSD, SizeBytes: 8e9, BootMedia: true},
		{ID: "b", Path: "/dev/sda", Kind: device.KindUSB, SizeBytes: 16e9, Mounted: true},
		{ID: "c", Path: "/dev/mmcblk1", Kind: device.KindEMMC, SizeBytes: 4e9},
	}, nil
}

func openBurn(t *testing.T, s Services) (*burnScreen, *appSimulation, i18n.Catalog) {
	t.Helper()
	catalog := i18n.For(i18n.LocaleEN)
	app, _, _ := newApp(s, catalog)
	b := newBurnScreen(app, s, catalog, func() {})
	simulation := startSimulation(t, app)
	app.QueueUpdateDraw(func() { app.SetRoot(b.root, true) })
	b.load()
	return b, simulation, catalog
}

func burnSelectEMMC(t *testing.T, b *burnScreen, sim *appSimulation) {
	t.Helper()
	if !waitForApp(
		t,
		sim.app,
		func() bool { return simulationContains(sim.screen, "/dev/mmcblk1") },
	) {
		t.Fatal("destinations not shown")
	}
	pressKey(sim.app, tcell.KeyDown, 0)
	pressKey(sim.app, tcell.KeyEnter, 0)
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, "Everything on /dev/mmcblk1 will be erased.")
	}) {
		t.Fatal("warning not shown before burning")
	}
	sim.app.QueueUpdateDraw(func() { sim.app.SetFocus(b.form) })
	pressKey(sim.app, tcell.KeyEnter, 0)
}

func TestBurnScreenHidesBootMediaAndBlocksBusy(t *testing.T) {
	b, sim, _ := openBurn(t, Services{Inventory: burnInventory, Images: burnImages})
	if !waitForApp(t, sim.app, func() bool { return simulationContains(sim.screen, "/dev/sda") }) {
		t.Fatal("destinations not shown")
	}
	if simulationContains(sim.screen, "/dev/mmcblk0") {
		t.Fatal("boot media is listed")
	}
	pressKey(sim.app, tcell.KeyEnter, 0) // current item is the busy USB
	if waitForApp(t, sim.app, func() bool { return b.target != nil }) {
		t.Fatalf("busy device selected: %v", b.target)
	}
	if simulationContains(sim.screen, "Everything on") {
		t.Fatal("warning shown for a blocked device")
	}
}

func TestBurnScreenShowsSuccessOnlyOnSuccess(t *testing.T) {
	for _, burnErr := range []error{nil, errors.New("verify failed")} {
		services := Services{
			Inventory: burnInventory, Images: burnImages,
			Burn: func(
				context.Context, image.Image, device.Device, Progress,
			) (installer.BurnResult, error) {
				return installer.BurnResult{}, burnErr
			},
		}
		b, sim, c := openBurn(t, services)
		burnSelectEMMC(t, b, sim)
		want := c.Burn.Success
		if burnErr != nil {
			want = c.ErrorPrefix + burnErr.Error()
		}
		if !waitForApp(t, sim.app, func() bool { return simulationContains(sim.screen, want) }) {
			t.Fatalf("%q not shown", want)
		}
		if burnErr != nil && simulationContains(sim.screen, c.Burn.Success) {
			t.Fatal("success shown after an error")
		}
	}
}

func TestBurnScreenWarnsOnCommandRate(t *testing.T) {
	b, sim, c := openBurn(t, Services{
		Inventory: burnInventory,
		Images:    burnImages,
		Burn: func(context.Context, image.Image, device.Device, Progress) (installer.BurnResult, error) {
			return installer.BurnResult{CommandRateNotApplied: true}, nil
		},
	})
	burnSelectEMMC(t, b, sim)
	if !waitForApp(t, sim.app, func() bool {
		return simulationContains(sim.screen, c.Burn.Success) &&
			simulationContains(sim.screen, "The image has no DDR command rate")
	}) {
		t.Fatal("success and command rate lines not both shown")
	}
}

func TestBurnScreenEmptyImages(t *testing.T) {
	b, sim, c := openBurn(t, Services{
		Inventory: burnInventory,
		Images:    func(context.Context) ([]image.Image, error) { return nil, nil },
	})
	if !waitForApp(
		t,
		sim.app,
		func() bool { return simulationContains(sim.screen, c.Burn.NoImages) },
	) {
		t.Fatal("empty-state text not shown")
	}
	var disabled bool
	waitForApp(t, sim.app, func() bool { disabled = b.form.GetButton(0).IsDisabled(); return true })
	if !disabled {
		t.Fatal("burn button enabled without images")
	}
}

func TestBurnScreenIgnoresEscWhileBurning(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	catalog := i18n.For(i18n.LocaleEN)
	s := Services{
		Inventory: burnInventory, Images: burnImages,
		Burn: func(
			_ context.Context, _ image.Image, _ device.Device, p Progress,
		) (installer.BurnResult, error) {
			p(1e9, 4e9)
			<-release
			return installer.BurnResult{}, nil
		},
	}
	app, list, _ := newApp(s, catalog)
	sim := startSimulation(t, app)
	app.QueueUpdateDraw(func() { list.SetCurrentItem(0) })
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool { return simulationContains(sim.screen, "/dev/mmcblk1") }) {
		t.Fatal("burn screen not shown")
	}
	pressKey(app, tcell.KeyDown, 0)
	pressKey(app, tcell.KeyEnter, 0)
	waitForApp(t, app, func() bool { return simulationContains(sim.screen, "Everything on") })
	pressKey(app, tcell.KeyTab, 0)
	pressKey(app, tcell.KeyTab, 0)
	pressKey(app, tcell.KeyEnter, 0)
	if !waitForApp(t, app, func() bool { return simulationContains(sim.screen, "25%") }) {
		t.Fatal("progress not shown")
	}
	pressKey(app, tcell.KeyEscape, 0)
	if waitForApp(
		t,
		app,
		func() bool { return simulationContains(sim.screen, "Hardware information") },
	) {
		t.Fatal("Esc left the screen while burning")
	}
}
