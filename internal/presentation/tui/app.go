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
	Reboot    func() error
	PowerOff  func() error
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
	var home *tview.Grid
	var burn *burnScreen
	var bak *backupScreen
	show := func(p tview.Primitive, title, navigation string) {
		onHome = false
		app.SetRoot(frame(p, catalog.ProductName, title, navigation), true)
	}
	goHome := func() {
		onHome = true
		app.SetRoot(frame(home, catalog.ProductName, catalog.HomeTitle, catalog.HomeNav), true)
	}
	showMessage := func(message string) {
		onHome = false
		app.SetRoot(actionModal().SetText(tview.Escape(message)).
			AddButtons([]string{catalog.OK}).SetDoneFunc(func(int, string) { goHome() }), true)
	}
	confirmPower := func(prompt string, action func() error) {
		if action == nil {
			showMessage(catalog.PowerUnavailable)
			return
		}
		onHome = false
		app.SetRoot(actionModal().SetText(prompt).
			AddButtons([]string{catalog.Backup.Cancel, catalog.OK}).SetDoneFunc(func(button int, _ string) {
			if button == 1 {
				if err := action(); err != nil {
					showMessage(errorText(catalog, err))
					return
				}
			}
			goHome()
		}), true)
	}
	hardwareView := tview.NewTextView().SetDynamicColors(false).SetScrollable(true).SetWordWrap(true)
	hardwareView.SetBorderPadding(1, 0, 0, 0)

	list = screens.Home(catalog, func(id i18n.MenuItemID) {
		switch id {
		case i18n.MenuBurn:
			burn = newBurnScreen(app, s, catalog, goHome)
			show(burn.root, catalog.Burn.Title, catalog.FormNav)
			burn.load()
		case i18n.MenuBackup:
			bak = newBackupScreen(app, s, catalog, goHome)
			show(bak.root, catalog.Backup.Title, catalog.FormNav)
			bak.load()
		case i18n.MenuHardware:
			hardwareView.SetText(catalog.Loading)
			show(hardwareView, catalog.HardwareTitle, catalog.HardwareNav)
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
				showMessage(errorText(catalog, err))
			}
		case i18n.MenuReboot:
			confirmPower(catalog.ConfirmReboot, s.Reboot)
		case i18n.MenuPowerOff:
			confirmPower(catalog.ConfirmPowerOff, s.PowerOff)
		case i18n.MenuExit:
			app.Stop()
		}
	})
	home = homeMenu(list)
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
			if burn == nil || !burn.dismissReview() {
				goHome()
			}
			return nil
		}
		if onHome && list.HasFocus() {
			switch event.Key() {
			case tcell.KeyLeft:
				return tcell.NewEventKey(tcell.KeyUp, 0, event.Modifiers())
			case tcell.KeyRight:
				return tcell.NewEventKey(tcell.KeyDown, 0, event.Modifiers())
			}
		}
		return event
	})
	return app, list, hardwareView
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
