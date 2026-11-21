package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	textColor  = tcell.ColorWhite
	mutedColor = tcell.ColorSilver
)

var (
	panelColor   = tcell.NewHexColor(0x222625)
	accentColor  = tcell.NewHexColor(0x50bc68)
	focusStyle   = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(accentColor)
	buttonStyle  = tcell.StyleDefault.Foreground(textColor).Background(panelColor)
	statusStyle  = tcell.StyleDefault.Foreground(textColor)
	warningStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
)

func frame(content tview.Primitive, product, title, navigation string) *tview.Flex {
	header := tview.NewTextView().SetDynamicColors(true).SetWrap(false).
		SetText(fmt.Sprintf("[black:#%06x] %s [-:-]  %s", accentColor.Hex(), tview.Escape(product), tview.Escape(title)))
	header.SetTextColor(textColor).SetBackgroundColor(panelColor)
	footer := tview.NewTextView().SetTextColor(mutedColor).SetWordWrap(true).SetText(navigation)
	footer.SetBackgroundColor(panelColor).SetBorderPadding(0, 0, 1, 1)
	body := tview.NewFlex().AddItem(content, 0, 1, true)
	body.SetBorderPadding(0, 0, 2, 2)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(footer, 2, 0, false)
	layout.SetDrawFunc(func(_ tcell.Screen, x, y, width, height int) (int, int, int, int) {
		w, h := min(width, 112), min(height, 32)
		return x + (width-w)/2, y + (height-h)/2, w, h
	})
	return layout
}

func homeMenu(list *tview.List) *tview.Grid {
	_, firstHint := list.GetItemText(0)
	hint := statusView(firstHint)
	hint.SetBorderPadding(1, 0, 2, 2)
	styleList(list).ShowSecondaryText(false).SetChangedFunc(func(_ int, _, help string, _ rune) {
		hint.SetText(help)
	})
	grid := tview.NewGrid().SetRows(list.GetItemCount(), 3).SetColumns(-1, -1).SetGap(1, 2)
	grid.SetBorderPadding(1, 0, 0, 0)
	return grid.AddItem(list, 0, 0, 1, 2, 0, 0, true).
		AddItem(hint, 1, 0, 1, 2, 0, 0, false).
		AddItem(list, 0, 0, 1, 1, 0, 64, true).
		AddItem(hint, 0, 1, 1, 1, 0, 64, false)
}

func styleList(list *tview.List) *tview.List {
	return list.SetMainTextColor(textColor).SetSecondaryTextColor(mutedColor).
		SetSelectedStyle(focusStyle).
		SetSelectedFocusOnly(true).SetHighlightFullLine(true)
}

func stylePanel(box *tview.Box, title string) {
	box.SetBorder(true).SetBorderColor(mutedColor).SetTitle(" "+title+" ").
		SetTitleAlign(tview.AlignLeft).SetTitleColor(accentColor).SetBorderPadding(0, 0, 1, 1)
}

func statusView(text string) *tview.TextView {
	view := tview.NewTextView().SetTextStyle(statusStyle).SetWordWrap(true).SetText(text)
	view.SetBackgroundColor(panelColor).SetBorderPadding(0, 0, 1, 1)
	return view
}

func actionForm() *tview.Form {
	return tview.NewForm().SetButtonStyle(buttonStyle).SetButtonActivatedStyle(focusStyle).
		SetButtonDisabledStyle(tcell.StyleDefault.Foreground(mutedColor).Background(tcell.ColorBlack))
}

func setActionDisabled(button *tview.Button, label, unavailable string, disabled bool) {
	if disabled {
		label += " (" + unavailable + ")"
	}
	button.SetLabel(label).SetDisabled(disabled)
}

func actionModal() *tview.Modal {
	modal := tview.NewModal().SetBackgroundColor(panelColor).SetTextColor(textColor).
		SetButtonStyle(buttonStyle).SetButtonActivatedStyle(focusStyle)
	modal.Box.SetBackgroundColor(panelColor).SetBorderColor(mutedColor)
	return modal
}

func focusNavigation(app *tview.Application, busy func() bool, order ...tview.Primitive) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		key := event.Key()
		switch key {
		case tcell.KeyLeft, tcell.KeyRight:
			if _, editing := app.GetFocus().(*tview.InputField); editing {
				return event
			}
			if key == tcell.KeyLeft {
				key = tcell.KeyBacktab
			} else {
				key = tcell.KeyTab
			}
		case tcell.KeyTab, tcell.KeyBacktab:
		default:
			return event
		}
		if busy() {
			return nil
		}
		step := 1
		if key == tcell.KeyBacktab {
			step = -1
		}
		for i, p := range order {
			if !p.HasFocus() {
				continue
			}
			for offset := 1; offset < len(order); offset++ {
				next := order[(i+step*offset+len(order))%len(order)]
				if button, ok := next.(*tview.Button); ok && button.IsDisabled() {
					continue
				}
				app.SetFocus(next)
				return nil
			}
			break
		}
		return nil
	}
}
