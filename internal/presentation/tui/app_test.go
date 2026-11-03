package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestKeyboardNavigationHardwareEscapeAndQuit(t *testing.T) {
	m := NewModel(
		func() ([]device.Device, error) { return []device.Device{{ID: "mmc0", Kind: device.KindEMMC}}, nil },
	)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("cursor=%d", m.cursor)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Fatalf("cursor=%d", m.cursor)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenHardware || len(m.devices) != 1 {
		t.Fatalf("screen=%v devices=%v", m.screen, m.devices)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != screenHome {
		t.Fatalf("screen=%v", m.screen)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("quit command missing")
	}
}
