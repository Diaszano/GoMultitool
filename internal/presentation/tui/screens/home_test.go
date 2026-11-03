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

func TestHomeOnlyActivatesHardwareAndExitOnEnter(t *testing.T) {
	var selected []i18n.MenuItemID
	list := Home(i18n.For(i18n.LocaleEN), func(id i18n.MenuItemID) {
		selected = append(selected, id)
	})
	handle := list.InputHandler()
	if handle == nil {
		t.Fatal("list input handler is nil")
	}
	for i, item := range i18n.For(i18n.LocaleEN).Menu {
		list.SetCurrentItem(i)
		callback := list.GetItemSelectedFunc(i)
		if item.ID == i18n.MenuHardware || item.ID == i18n.MenuExit {
			if callback == nil {
				t.Errorf("actionable item %q has no callback", item.ID)
			}
		} else if callback != nil {
			t.Errorf("future item %q has a callback", item.ID)
		}
		handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(tview.Primitive) {})
	}
	if len(selected) != 2 || selected[0] != i18n.MenuHardware || selected[1] != i18n.MenuExit {
		t.Fatalf("Enter selected %v, want hardware and exit only", selected)
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
