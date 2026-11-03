// Package screens builds terminal-interface widgets.
package screens

import (
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

// Home builds the localized menu and passes selected item indexes to onSelect.
func Home(catalog i18n.Catalog, onSelect func(int)) *tview.List {
	list := tview.NewList().ShowSecondaryText(true)
	for _, item := range catalog.Menu {
		list.AddItem(item.Label, item.Hint, 0, nil)
	}
	list.SetSelectedFunc(func(index int, _, _ string, _ rune) { onSelect(index) })
	return list
}
