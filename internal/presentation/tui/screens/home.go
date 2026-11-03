// Package screens builds terminal-interface widgets.
package screens

import (
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

// Home builds the localized menu; only Hardware and Exit invoke onSelect.
func Home(catalog i18n.Catalog, onSelect func(i18n.MenuItemID)) *tview.List {
	list := tview.NewList().ShowSecondaryText(true)
	for _, item := range catalog.Menu {
		var selected func()
		if item.ID == i18n.MenuHardware || item.ID == i18n.MenuExit {
			id := item.ID
			selected = func() { onSelect(id) }
		}
		list.AddItem(item.Label, item.Hint, 0, selected)
	}
	return list
}
