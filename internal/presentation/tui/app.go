// Package tui implements the keyboard-driven terminal interface.
package tui

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/application/installer"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
	"github.com/Diaszano/GoMultitool/internal/infrastructure/multitool"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui/screens"
)

// Progress reports streamed bytes; total 0 means unknown.
// It is called from the operation goroutine; wrap UI updates in QueueUpdateDraw.
type Progress func(done, total int64)

// Services are the operations the TUI runs off the UI goroutine.
type Services struct {
	Inventory func(context.Context) ([]device.Device, error)
	Images    func(context.Context) ([]image.Image, error)
	FreeBytes func(context.Context) (int64, error)
	Burn      func(
		ctx context.Context, img image.Image, target device.Device, p Progress,
	) (installer.BurnResult, error)
	Backup    func(ctx context.Context, source device.Device, name string, p Progress) error
	RunLegacy func() error
}

// New builds the full-screen terminal application.
func New(s Services, c i18n.Catalog) *tview.Application {
	app, _, _ := newApp(s, c)
	return app
}

func newApp(s Services, catalog i18n.Catalog) (*tview.Application, *tview.List, *tview.TextView) {
	app := tview.NewApplication()
	onHome := true
	var list *tview.List
	var burn *burnScreen
	var bak *backupScreen
	show := func(p tview.Primitive) {
		onHome = false
		app.SetRoot(withNavigation(p, catalog), true)
	}
	goHome := func() {
		onHome = true
		app.SetRoot(withNavigation(list, catalog), true)
	}
	hardwareView := tview.NewTextView().SetDynamicColors(false).SetScrollable(true)
	hardwareView.SetBorder(true).SetTitle(catalog.HardwareTitle)

	list = screens.Home(catalog, func(id i18n.MenuItemID) {
		switch id {
		case i18n.MenuBurn:
			burn = newBurnScreen(app, s, catalog, goHome)
			show(burn.root)
			burn.load()
		case i18n.MenuBackup:
			bak = newBackupScreen(app, s, catalog, goHome)
			show(bak.root)
			bak.load()
		case i18n.MenuHardware:
			hardwareView.SetText(catalog.Loading)
			show(hardwareView)
			go func() {
				devices, err := s.Inventory(context.Background())
				app.QueueUpdateDraw(func() {
					hardwareView.SetText(renderHardware(devices, err, catalog))
				})
			}()
		case i18n.MenuLegacy:
			var err error
			app.Suspend(func() { err = s.RunLegacy() })
			if err != nil {
				onHome = false
				app.SetRoot(tview.NewModal().
					SetText(errorText(catalog, err)).
					AddButtons([]string{catalog.OK}).
					SetDoneFunc(func(int, string) { goHome() }), true)
			}
		case i18n.MenuExit:
			app.Stop()
		}
	})
	list.SetTitle(catalog.ProductName).SetBorder(true)
	goHome()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var running *atomic.Bool
		var cancel func()
		switch {
		case burn != nil && burn.burning.Load():
			running, cancel = &burn.burning, burn.cancel
		case bak != nil && bak.running.Load():
			running, cancel = &bak.running, bak.cancel
		}
		if running != nil && !onHome {
			if event.Key() == tcell.KeyCtrlC {
				cancel() // stopping mid-operation would leave partial work silently
				return nil
			}
			// Never leave or quit mid-operation.
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyRune {
				return nil
			}
		}
		_, typing := app.GetFocus().(*tview.InputField)
		if !typing && event.Key() == tcell.KeyRune &&
			strings.EqualFold(string(event.Rune()), "q") {
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyEscape && !onHome {
			goHome()
			return nil
		}
		return event
	})
	return app, list, hardwareView
}

func withNavigation(p tview.Primitive, catalog i18n.Catalog) tview.Primitive {
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(tview.NewTextView().SetText(catalog.Navigation), 1, 0, false)
}

// errorText localizes known errors and prefixes the rest.
func errorText(catalog i18n.Catalog, err error) string {
	for sentinel, text := range map[error]string{
		multitool.ErrBusy:             catalog.Busy,
		installation.ErrImageTooLarge: catalog.ImageTooLarge,
		storageguard.ErrDeviceChanged: catalog.DeviceChanged,
		device.ErrUnsafeDevice:        catalog.UnsafeDevice,
		installer.ErrReadbackMismatch: catalog.ReadbackMismatch,
	} {
		if errors.Is(err, sentinel) {
			return text
		}
	}
	return catalog.ErrorPrefix + err.Error()
}

func renderHardware(devices []device.Device, err error, catalog i18n.Catalog) string {
	if err != nil {
		return errorText(catalog, err)
	}
	if len(devices) == 0 {
		return catalog.EmptyInventory
	}
	var output strings.Builder
	for _, d := range devices {
		output.WriteString(catalog.DeviceLine(d))
		output.WriteByte('\n')
	}
	return output.String()
}
