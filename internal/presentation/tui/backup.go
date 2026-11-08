package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/application/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

const backupSuffix = ".img.gz"

// backupScreen is the single-screen backup: source, name, free space, buttons.
type backupScreen struct {
	app      *tview.Application
	services Services
	catalog  i18n.Catalog
	root     *tview.Flex
	list     *tview.List
	name     *tview.InputField
	free     *tview.TextView
	status   *tview.TextView
	form     *tview.Form
	devices  []device.Device
	target   *device.Device
	running  atomic.Bool
	cancel   context.CancelFunc
}

func newBackupScreen(
	app *tview.Application, s Services, catalog i18n.Catalog, back func(),
) *backupScreen {
	b := &backupScreen{app: app, services: s, catalog: catalog}
	t := catalog.Backup
	b.list = tview.NewList().ShowSecondaryText(false)
	b.list.SetBorder(true).SetTitle(t.Source)
	b.name = tview.NewInputField().SetLabel(t.Name + ": ").
		SetText("tvbox-backup-" + time.Now().Format("2006-01-02"))
	b.free = tview.NewTextView()
	b.status = tview.NewTextView()
	b.form = tview.NewForm().
		AddButton(t.Start, b.start).
		AddButton(t.Cancel, func() {
			if !b.running.Load() {
				back()
			}
		})
	b.form.GetButton(0).SetDisabled(true)
	nameRow := tview.NewFlex().
		AddItem(b.name, 0, 1, false).
		AddItem(tview.NewTextView().SetText(backupSuffix), len(backupSuffix)+1, 0, false)
	b.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(b.list, 0, 1, true).
		AddItem(nameRow, 1, 0, false).
		AddItem(b.free, 1, 0, false).
		AddItem(b.status, 3, 0, false).
		AddItem(b.form, 3, 0, false)
	order := []tview.Primitive{b.list, b.name, b.form}
	b.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyTab && event.Key() != tcell.KeyBacktab {
			return event
		}
		if b.running.Load() {
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
	b.list.SetSelectedFunc(func(i int, _ string, _ string, _ rune) { b.selectSource(i) })
	return b
}

// load fills the source list and free space off the UI goroutine.
func (b *backupScreen) load() {
	go func() {
		ctx := context.Background()
		devices, derr := b.services.Inventory(ctx)
		free, ferr := b.services.FreeBytes(ctx)
		b.app.QueueUpdateDraw(func() {
			b.devices = b.devices[:0]
			for _, d := range devices {
				if !d.BootMedia {
					b.devices = append(b.devices, d)
				}
			}
			b.list.Clear()
			for _, d := range b.devices {
				b.list.AddItem(markedDevice(b.catalog, b.target, d), "", 0, nil)
			}
			if ferr != nil {
				b.free.SetText(errorText(b.catalog, ferr))
			} else {
				b.free.SetText(fmt.Sprintf(b.catalog.Backup.Free, b.catalog.Size(free)))
			}
			switch {
			case derr != nil:
				b.status.SetText(errorText(b.catalog, derr))
			case len(b.devices) == 0:
				b.status.SetText(b.catalog.Backup.NoSource)
			}
			b.updateButton()
		})
	}()
}

func (b *backupScreen) selectSource(i int) {
	if b.running.Load() || i < 0 || i >= len(b.devices) || blocked(b.devices[i]) {
		return
	}
	b.target = &b.devices[i]
	for j, d := range b.devices {
		b.list.SetItemText(j, markedDevice(b.catalog, b.target, d), "")
	}
	b.updateButton()
}

func (b *backupScreen) updateButton() {
	b.form.GetButton(0).SetDisabled(b.running.Load() || b.target == nil)
}

// start runs on the UI goroutine (button handler).
func (b *backupScreen) start() {
	t := b.catalog.Backup
	name := b.name.GetText()
	if b.running.Load() || b.target == nil {
		return
	}
	if name == "" || strings.Contains(name, "/") || strings.HasPrefix(name, ".") {
		b.status.SetText(t.InvalidName)
		return
	}
	source := *b.target
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.running.Store(true)
	b.updateButton()
	b.form.GetButton(1).SetDisabled(true)
	b.status.SetText(progressLine(b.catalog, 0, 0))
	go func() {
		err := b.services.Backup(ctx, source, name, func(done, total int64) {
			b.app.QueueUpdateDraw(func() { b.status.SetText(progressLine(b.catalog, done, total)) })
		})
		b.app.QueueUpdateDraw(func() {
			cancel()
			b.running.Store(false)
			b.form.GetButton(1).SetDisabled(false)
			b.updateButton()
			switch {
			case errors.Is(err, context.Canceled):
				b.status.SetText(t.Cancelled)
			case errors.Is(err, backup.ErrNoSpace):
				b.status.SetText(t.NoSpace)
			case errors.Is(err, backup.ErrExists):
				b.status.SetText(t.Exists)
			case err != nil:
				b.status.SetText(errorText(b.catalog, err))
			default:
				b.status.SetText(fmt.Sprintf(t.Saved, name))
			}
		})
	}()
}
