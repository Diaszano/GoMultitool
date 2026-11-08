package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

const barWidth = 40

// burnScreen is the single-screen image burn: destination, image, warning, buttons.
type burnScreen struct {
	app      *tview.Application
	services Services
	catalog  i18n.Catalog
	root     *tview.Flex
	dest     *tview.List
	images   *tview.List
	status   *tview.TextView
	form     *tview.Form
	devices  []device.Device
	files    []image.Image
	target   *device.Device
	burning  atomic.Bool
	cancel   context.CancelFunc
}

func newBurnScreen(
	app *tview.Application, s Services, catalog i18n.Catalog, back func(),
) *burnScreen {
	b := &burnScreen{app: app, services: s, catalog: catalog}
	t := catalog.Burn
	b.dest = tview.NewList().ShowSecondaryText(false)
	b.dest.SetBorder(true).SetTitle(t.Destination)
	b.images = tview.NewList().ShowSecondaryText(false)
	b.images.SetBorder(true).SetTitle(t.Images)
	b.status = tview.NewTextView()
	b.form = tview.NewForm().
		AddButton(t.Start, b.start).
		AddButton(t.Cancel, func() {
			if !b.burning.Load() {
				back()
			}
		})
	b.form.GetButton(0).SetDisabled(true)
	b.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(b.dest, 0, 1, true).
		AddItem(b.images, 0, 1, false).
		AddItem(b.status, 3, 0, false).
		AddItem(b.form, 3, 0, false)
	order := []tview.Primitive{b.dest, b.images, b.form}
	b.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyTab && event.Key() != tcell.KeyBacktab {
			return event
		}
		if b.burning.Load() {
			return nil
		}
		step := 1
		if event.Key() == tcell.KeyBacktab {
			step = len(order) - 1
		}
		for i, p := range order {
			if p.HasFocus() {
				b.app.SetFocus(order[(i+step)%len(order)])
				break
			}
		}
		return nil
	})
	b.dest.SetSelectedFunc(func(i int, _ string, _ string, _ rune) { b.selectDevice(i) })
	return b
}

// blocked devices are listed but cannot be chosen.
func blocked(d device.Device) bool {
	return d.Mounted || d.Kind == device.KindNAND || d.Kind == device.KindUnknown
}

// load fills both lists off the UI goroutine.
func (b *burnScreen) load() {
	go func() {
		ctx := context.Background()
		devices, derr := b.services.Inventory(ctx)
		files, ferr := b.services.Images(ctx)
		b.app.QueueUpdateDraw(func() {
			b.devices = b.devices[:0]
			for _, d := range devices {
				if !d.BootMedia {
					b.devices = append(b.devices, d)
				}
			}
			b.files = files
			b.render(derr, ferr)
		})
	}()
}

func (b *burnScreen) render(derr, ferr error) {
	t := b.catalog.Burn
	b.dest.Clear()
	for _, d := range b.devices {
		b.dest.AddItem(b.deviceText(d), "", 0, nil)
	}
	b.images.Clear()
	for _, f := range b.files {
		size := t.SizeAtBurn
		if f.SizeBytes > 0 {
			size = b.catalog.Size(f.SizeBytes)
		}
		b.images.AddItem(fmt.Sprintf("%s  %s", f.Path, size), "", 0, nil)
	}
	var notes []string
	switch {
	case derr != nil:
		notes = append(notes, errorText(b.catalog, derr))
	case len(b.devices) == 0:
		notes = append(notes, t.NoDestination)
	}
	switch {
	case ferr != nil:
		notes = append(notes, errorText(b.catalog, ferr))
	case len(b.files) == 0:
		notes = append(notes, t.NoImages)
	}
	b.status.SetText(strings.Join(notes, "\n"))
	b.updateButton()
}

func (b *burnScreen) selectDevice(i int) {
	if b.burning.Load() || i < 0 || i >= len(b.devices) || blocked(b.devices[i]) {
		return
	}
	b.target = &b.devices[i]
	for j, d := range b.devices {
		b.dest.SetItemText(j, b.deviceText(d), "")
	}
	b.status.SetText(fmt.Sprintf(b.catalog.Burn.Warning, b.target.Path))
	b.updateButton()
}

// deviceText marks the chosen destination so the cursor position cannot mislead.
func (b *burnScreen) deviceText(d device.Device) string {
	if b.target != nil && b.target.ID == d.ID {
		return "▶ " + b.catalog.DeviceLine(d)
	}
	return b.catalog.DeviceLine(d)
}

func (b *burnScreen) updateButton() {
	b.form.GetButton(0).SetDisabled(b.burning.Load() || b.target == nil || len(b.files) == 0)
}

// start runs on the UI goroutine (button handler).
func (b *burnScreen) start() {
	i := b.images.GetCurrentItem()
	if b.burning.Load() || b.target == nil || i < 0 || i >= len(b.files) {
		return
	}
	img, target := b.files[i], *b.target
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.burning.Store(true)
	b.updateButton()
	b.form.GetButton(1).SetDisabled(true)
	b.status.SetText(b.progressText(0, 0))
	go func() {
		result, err := b.services.Burn(ctx, img, target, func(done, total int64) {
			b.app.QueueUpdateDraw(func() { b.status.SetText(b.progressText(done, total)) })
		})
		b.app.QueueUpdateDraw(func() {
			cancel()
			b.burning.Store(false)
			b.form.GetButton(1).SetDisabled(false)
			b.updateButton()
			t := b.catalog.Burn
			switch {
			case errors.Is(err, context.Canceled):
				b.status.SetText(t.Cancelled)
			case err != nil:
				b.status.SetText(errorText(b.catalog, err))
			case result.CommandRateNotApplied:
				b.status.SetText(t.Success + "\n" + t.CommandRate)
			default:
				b.status.SetText(t.Success)
			}
		})
	}()
}

func (b *burnScreen) progressText(done, total int64) string {
	if total <= 0 {
		return fmt.Sprintf("%s\n%s", b.catalog.Size(done), strings.Repeat("░", barWidth))
	}
	done = max(0, done)
	filled := int(min(done, total) * barWidth / total)
	return fmt.Sprintf("%d%%  %s / %s\n%s%s", min(done*100/total, 100), b.catalog.Size(done),
		b.catalog.Size(total), strings.Repeat("█", filled), strings.Repeat("░", barWidth-filled))
}
