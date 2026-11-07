package screens

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func TestHomeRendersLocalizedMenu(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		catalog := i18n.For(locale)
		list := Home(catalog, func(i18n.MenuItemID) {})
		if list.GetItemCount() != len(catalog.Menu) {
			t.Fatalf("locale %q entries = %d", locale, list.GetItemCount())
		}
		for i, expected := range catalog.Menu {
			label, hint := list.GetItemText(i)
			if label != expected.Label || hint != expected.Hint {
				t.Errorf("locale %q item %d = (%q, %q)", locale, i, label, hint)
			}
		}
	}
}

func TestHomeActivatesEveryItemOnEnter(t *testing.T) {
	var selected []i18n.MenuItemID
	list := Home(i18n.For(i18n.LocaleEN), func(id i18n.MenuItemID) {
		selected = append(selected, id)
	})
	handle := list.InputHandler()
	if handle == nil {
		t.Fatal("list input handler is nil")
	}
	menu := i18n.For(i18n.LocaleEN).Menu
	for i := range menu {
		list.SetCurrentItem(i)
		handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(tview.Primitive) {})
	}
	if len(selected) != len(menu) {
		t.Fatalf("Enter selected %v, want every menu item", selected)
	}
	for i, item := range menu {
		if selected[i] != item.ID {
			t.Errorf("selection %d = %q, want %q", i, selected[i], item.ID)
		}
	}
}

func TestHomeArrowKeysMoveSelection(t *testing.T) {
	list := Home(i18n.For(i18n.LocaleEN), func(i18n.MenuItemID) {})
	handle := list.InputHandler()
	handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), func(tview.Primitive) {})
	if got := list.GetCurrentItem(); got != 1 {
		t.Fatalf("after Down, selection = %d, want 1", got)
	}
	handle(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), func(tview.Primitive) {})
	if got := list.GetCurrentItem(); got != 0 {
		t.Fatalf("after Up, selection = %d, want 0", got)
	}
}
