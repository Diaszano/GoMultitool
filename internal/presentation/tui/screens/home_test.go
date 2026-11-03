package screens

import (
	"testing"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func TestHomeRendersLocalizedMenu(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		catalog := i18n.For(locale)
		list := Home(catalog, func(int) {})
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
