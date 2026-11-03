// Package tui implements the keyboard-driven terminal interface.
package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui/screens"
)

// New builds the full-screen terminal application.
func New(hardware func() ([]device.Device, error), catalog i18n.Catalog) *tview.Application {
	app, _, _ := newApp(hardware, catalog)
	return app
}

func newApp(
	hardware func() ([]device.Device, error),
	catalog i18n.Catalog,
) (*tview.Application, *tview.List, *tview.TextView) {
	app := tview.NewApplication()
	var showingHardware bool
	var hardwareView *tview.TextView

	list := screens.Home(catalog, func(id i18n.MenuItemID) {
		switch id {
		case i18n.MenuHardware:
			var devices []device.Device
			var err error
			if hardware != nil {
				devices, err = hardware()
			}
			hardwareView.SetText(renderHardware(devices, err, catalog))
			showingHardware = true
			app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(hardwareView, 0, 1, true).
				AddItem(tview.NewTextView().SetText(catalog.Navigation), 1, 0, false), true)
		case i18n.MenuExit:
			app.Stop()
		}
	})
	list.SetTitle(catalog.ProductName).SetBorder(true)
	hardwareView = tview.NewTextView().SetDynamicColors(false).SetScrollable(true)
	hardwareView.SetBorder(true).SetTitle(catalog.HardwareTitle)
	app.SetRoot(homeRoot(list, catalog), true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && strings.EqualFold(string(event.Rune()), "q") {
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyEscape && showingHardware {
			showingHardware = false
			app.SetRoot(homeRoot(list, catalog), true)
			return nil
		}
		return event
	})
	return app, list, hardwareView
}

func homeRoot(list *tview.List, catalog i18n.Catalog) tview.Primitive {
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(tview.NewTextView().SetText(catalog.Navigation), 1, 0, false)
}

func renderHardware(devices []device.Device, err error, catalog i18n.Catalog) string {
	if err != nil {
		return catalog.ErrorPrefix + err.Error()
	}
	if len(devices) == 0 {
		return catalog.EmptyInventory
	}
	var output strings.Builder
	for _, d := range devices {
		fmt.Fprintf(&output, "%s  %s  %s  %d bytes", d.Path, d.Kind, d.ID, d.SizeBytes)
		if d.Mounted {
			fmt.Fprintf(&output, "  (%s)", catalog.MountedLabel)
		}
		output.WriteByte('\n')
	}
	return output.String()
}
