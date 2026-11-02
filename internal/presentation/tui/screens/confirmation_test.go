package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestConfirmationDefaultsNoAndNamesExactTarget(t *testing.T) {
	c := NewConfirmation(device.Device{ID: "mmcblk2", Path: "/dev/mmcblk2", SizeBytes: 8192})
	if c.Confirmed {
		t.Fatal("confirmation defaulted to yes")
	}
	view := c.View()
	if !strings.Contains(view, "/dev/mmcblk2") || !strings.Contains(view, "mmcblk2") {
		t.Fatalf("view=%q", view)
	}
	c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !c.Confirmed {
		t.Fatal("accept was not recorded")
	}
	c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if c.Confirmed {
		t.Fatal("reject did not clear confirmation")
	}
}
