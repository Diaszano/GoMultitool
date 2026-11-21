package tui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

const barWidth = 40

// burnScreen is the single-screen image burn: destination, image, warning, buttons.
type burnScreen struct {
	app        *tview.Application
	services   Services
	catalog    i18n.Catalog
	root       *tview.Pages
	dest       *tview.List
	images     *tview.List
	status     *tview.TextView
	form       *tview.Form
	devices    []device.Device
	files      []image.Image
	target     *device.Device
	selected   int
	imageWidth int
	deviceNote string
	imageNote  string
	burning    atomic.Bool
	cancel     context.CancelFunc
	meter      meter
}

func newBurnScreen(
	app *tview.Application, s Services, catalog i18n.Catalog, back func(),
) *burnScreen {
	b := &burnScreen{app: app, services: s, catalog: catalog, selected: -1}
	t := catalog.Burn
	b.dest = styleList(tview.NewList().ShowSecondaryText(false))
	stylePanel(b.dest.Box, t.Destination)
	b.images = styleList(tview.NewList().ShowSecondaryText(false))
	stylePanel(b.images.Box, t.Images)
	b.images.SetDrawFunc(func(tcell.Screen, int, int, int, int) (int, int, int, int) {
		x, y, width, height := b.images.GetInnerRect()
		if b.imageWidth != width {
			b.imageWidth = width
			b.updateImageRows()
		}
		return x, y, width, height
	})
	b.status = statusView(catalog.Loading)
	b.form = actionForm().
		AddButton(t.Start, b.review).
		AddButton(t.Cancel, func() {
			if !b.burning.Load() {
				back()
			}
		})
	b.updateButton()
	selection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(b.dest, 0, 1, true).
		AddItem(b.images, 0, 1, false).
		AddItem(b.status, 3, 0, false).
		AddItem(b.form, 3, 0, false)
	selection.SetInputCapture(focusNavigation(app, b.burning.Load, b.dest, b.images, b.form.GetButton(0), b.form.GetButton(1)))
	b.root = tview.NewPages().AddPage("selection", selection, true, true)
	b.dest.SetSelectedFunc(func(i int, _ string, _ string, _ rune) { b.selectDevice(i) })
	b.images.SetSelectedFunc(func(i int, _ string, _ string, _ rune) {
		if !b.burning.Load() && i >= 0 && i < len(b.files) {
			b.selected = i
			b.updateImageRows()
			b.updateSelection()
		}
	})
	return b
}

func availableDevice(d device.Device) bool {
	return !d.Mounted && d.Kind != device.KindNAND && d.Kind != device.KindUnknown
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
	for range b.files {
		b.images.AddItem("", "", 0, nil)
	}
	b.updateImageRows()
	switch {
	case derr != nil:
		b.deviceNote = errorText(b.catalog, derr)
	case !slices.ContainsFunc(b.devices, availableDevice):
		b.deviceNote = t.NoDestination
	}
	switch {
	case ferr != nil:
		b.imageNote = errorText(b.catalog, ferr)
	case len(b.files) == 0:
		b.imageNote = t.NoImages
	}
	b.updateSelection()
}

func imageName(img image.Image) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '�'
		}
		return r
	}, filepath.Base(img.Path))
}

func (b *burnScreen) updateImageRows() {
	for i, img := range b.files {
		prefix, size := "  ", ""
		if i == b.selected {
			prefix = "▶ "
		}
		if img.SizeBytes > 0 {
			size = "  " + b.catalog.Size(img.SizeBytes)
		}
		name := imageName(img)
		width := b.imageWidth - tview.TaggedStringWidth(prefix+size)
		if tview.TaggedStringWidth(tview.Escape(name)) > width {
			for name != "" && tview.TaggedStringWidth(tview.Escape(name)) >= width {
				_, n := utf8.DecodeRuneInString(name)
				name = name[n:]
			}
			name = "…" + name
		}
		b.images.SetItemText(i, prefix+tview.Escape(name)+size, "")
	}
}

func (b *burnScreen) selectDevice(i int) {
	if b.burning.Load() || i < 0 || i >= len(b.devices) || !availableDevice(b.devices[i]) {
		return
	}
	b.target = &b.devices[i]
	for j, d := range b.devices {
		b.dest.SetItemText(j, b.deviceText(d), "")
	}
	b.updateSelection()
}

func (b *burnScreen) updateSelection() {
	t := b.catalog.Burn
	if b.target == nil {
		var text string
		switch {
		case b.deviceNote != "" && b.imageNote != "":
			text = b.deviceNote + "\n" + b.imageNote
		case b.deviceNote != "":
			text = b.deviceNote
		case b.imageNote != "":
			text = b.imageNote
		case b.selected < 0:
			text = t.PickBoth
		default:
			text = t.PickDevice
		}
		b.status.SetTextStyle(statusStyle).SetText(text)
	} else {
		warning := fmt.Sprintf(t.Warning, b.target.Path)
		if b.imageNote != "" {
			warning += "\n" + b.imageNote
		} else if b.selected < 0 {
			warning += "\n" + t.PickImage
		}
		b.status.SetTextStyle(warningStyle).SetText(warning)
	}
	b.updateButton()
}

// deviceText marks the chosen destination so the cursor position cannot mislead.
func (b *burnScreen) deviceText(d device.Device) string {
	return markedDevice(b.catalog, b.target, d)
}

func markedDevice(catalog i18n.Catalog, target *device.Device, d device.Device) string {
	line := catalog.DeviceLine(d)
	if !d.Mounted && !availableDevice(d) { // a mounted row already says "in use"
		line += "  · " + catalog.NotSelectable
	}
	if target != nil && target.ID == d.ID {
		return "▶ " + line
	}
	return line
}

func (b *burnScreen) updateButton() {
	setActionDisabled(b.form.GetButton(0), b.catalog.Burn.Start, b.catalog.Unavailable,
		b.burning.Load() || b.target == nil || b.selected < 0 || b.deviceNote != "" || b.imageNote != "")
}

func (b *burnScreen) review() {
	if b.form.GetButton(0).IsDisabled() {
		return
	}
	img, target := b.files[b.selected], *b.target
	t := b.catalog.Burn
	size := t.SizeAtBurn
	if img.SizeBytes > 0 {
		size = b.catalog.Size(img.SizeBytes)
	}
	summary := statusView(fmt.Sprintf("%s\n\n%s: %s\n%s: %s\n%s: %s",
		fmt.Sprintf(t.Warning, target.Path), t.Destination, b.catalog.DeviceLine(target), t.Size, size, t.Images, imageName(img)))
	summary.SetScrollable(true)
	stylePanel(summary.Box, t.Review)
	buttons := actionForm().AddButton(t.Cancel, func() { b.dismissReview() }).AddButton(t.Start, func() {
		b.dismissReview()
		b.start(img, target)
	})
	panel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(summary, 0, 1, false).AddItem(buttons, 3, 0, true)
	panel.SetInputCapture(focusNavigation(b.app, b.burning.Load, buttons.GetButton(0), buttons.GetButton(1), summary))
	b.root.AddAndSwitchToPage("review", panel, true)
}

func (b *burnScreen) dismissReview() bool {
	if !b.root.HasPage("review") {
		return false
	}
	b.root.RemovePage("review")
	b.app.SetFocus(b.form.GetButton(0))
	return true
}

// start runs on the UI goroutine after a reviewed selection is confirmed.
func (b *burnScreen) start(img image.Image, target device.Device) {
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.burning.Store(true)
	b.updateButton()
	b.form.GetButton(1).SetDisabled(true)
	b.meter = meter{}
	b.status.SetTextStyle(statusStyle).SetText(b.meter.line(b.catalog, 0, 0))
	go func() {
		result, err := b.services.Burn(ctx, img, target, func(done, total int64) {
			b.app.QueueUpdateDraw(func() { b.status.SetText(b.meter.line(b.catalog, done, total)) })
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

// meter times the transfer for the speed shown next to the progress.
// It is used only on the UI goroutine.
type meter struct{ started time.Time }

func (m *meter) line(catalog i18n.Catalog, done, total int64) string {
	switch {
	case done < 0:
		m.started = time.Time{} // a phase change restarts the clock for the next transfer
	case m.started.IsZero():
		m.started = time.Now()
	}
	return progressLine(catalog, done, total, time.Since(m.started))
}

// progressLine renders a progress update; done -1 and -2 are the sizing and readback phases.
func progressLine(catalog i18n.Catalog, done, total int64, elapsed time.Duration) string {
	switch done {
	case -1:
		return catalog.Burn.Sizing
	case -2:
		return catalog.Burn.Verifying
	}
	speed := ""
	if elapsed >= time.Second {
		mbps := fmt.Sprintf("%.1f", float64(done)/1e6/elapsed.Seconds())
		speed = " · " + strings.Replace(mbps, ".", catalog.DecimalSeparator, 1) + " MB/s"
	}
	if total <= 0 {
		return fmt.Sprintf("%s%s\n%s", catalog.Size(done), speed, strings.Repeat("░", barWidth))
	}
	done = max(0, done)
	filled := int(min(done, total) * barWidth / total)
	return fmt.Sprintf("%d%%  %s / %s%s\n%s%s", min(done*100/total, 100), catalog.Size(done),
		catalog.Size(total), speed, strings.Repeat("█", filled),
		strings.Repeat("░", barWidth-filled))
}
