// Package screens builds terminal-interface widgets.
package screens

import (
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

// Home builds the localized menu; every entry invokes onSelect with its ID.
func Home(catalog i18n.Catalog, onSelect func(i18n.MenuItemID)) *tview.List {
	list := tview.NewList().ShowSecondaryText(true)
	for _, item := range catalog.Menu {
		id := item.ID
		list.AddItem(item.Label, item.Hint, 0, func() { onSelect(id) })
	}
	return list
}
