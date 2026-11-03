package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

func TestMenuShowsAllAreasAndOnlyHardwareRuns(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		t.Run(string(locale), func(t *testing.T) {
			calls := 0
			app, list, _ := newApp(func() ([]device.Device, error) {
				calls++
				return []device.Device{{ID: "mmc0", Path: "/dev/mmc0", SizeBytes: 512}}, nil
			}, i18n.For(locale))
			if list.GetItemCount() != 10 {
				t.Fatalf("menu entries = %d, want 10", list.GetItemCount())
			}
			for i, item := range i18n.For(locale).Menu {
				label, hint := list.GetItemText(i)
				if label != item.Label || hint != item.Hint {
					t.Errorf(
						"menu item %d = (%q, %q), want (%q, %q)",
						i,
						label,
						hint,
						item.Label,
						item.Hint,
					)
				}
			}
			selectItem := list.GetSelectedFunc()
			if selectItem == nil {
				t.Fatal("menu selection handler is nil")
			}
			for i := 1; i < 9; i++ {
				selectItem(i, "", "", 0)
			}
			if calls != 0 {
				t.Fatalf("future menu items invoked hardware %d times", calls)
			}
			selectItem(0, "", "", 0)
			if calls != 1 {
				t.Fatalf("hardware action calls = %d, want 1", calls)
			}
			if app == nil {
				t.Fatal("tview application is nil")
			}
		})
	}
}

func TestHardwareScreenShowsInventoryEmptyAndError(t *testing.T) {
	tests := []struct {
		name     string
		devices  []device.Device
		err      error
		contains string
	}{
		{
			name: "inventory",
			devices: []device.Device{{
				ID: "mmc0", Path: "/dev/mmc0", Kind: device.KindEMMC, SizeBytes: 512,
			}},
			contains: "/dev/mmc0",
		},
		{name: "empty", contains: "Nenhum dispositivo encontrado."},
		{name: "error", err: errors.New("sysfs unavailable"), contains: "sysfs unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, list, hardware := newApp(
				func() ([]device.Device, error) { return tt.devices, tt.err },
				i18n.For(i18n.LocalePtBR),
			)
			list.GetSelectedFunc()(0, "", "", 0)
			if got := hardware.GetText(false); !strings.Contains(got, tt.contains) {
				t.Fatalf("hardware text = %q, want it to contain %q", got, tt.contains)
			}
			if app == nil {
				t.Fatal("tview application is nil")
			}
		})
	}
}
