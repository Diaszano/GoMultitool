// Package tui implements the keyboard-driven terminal interface.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui/screens"
)

type screen uint8

const (
	screenHome screen = iota
	screenHardware
)

// Model contains the TUI navigation state and hardware inventory use case.
type Model struct {
	cursor    int
	screen    screen
	devices   []device.Device
	errorText string
	Hardware  func() ([]device.Device, error)
}

// NewModel creates the terminal interface with its inventory callback.
func NewModel(hardware func() ([]device.Device, error)) *Model { return &Model{Hardware: hardware} }

// Init starts the interface without background work.
func (m *Model) Init() tea.Cmd { return nil }

// Update applies keyboard navigation events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		if m.screen == screenHardware {
			m.screen = screenHome
		}
		return m, nil
	case tea.KeyUp:
		if m.screen == screenHome && m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.screen == screenHome && m.cursor < len(i18n.MainMenu)-1 {
			m.cursor++
		}
		return m, nil
	case tea.KeyEnter:
		if m.screen == screenHome {
			if m.cursor == 1 {
				return m, tea.Quit
			}
			if m.Hardware != nil {
				var err error
				m.devices, err = m.Hardware()
				if err != nil {
					m.errorText = err.Error()
				}
			}
			m.screen = screenHardware
		}
		return m, nil
	}
	if key.Type == tea.KeyRunes && strings.EqualFold(string(key.Runes), "q") {
		return m, tea.Quit
	}
	return m, nil
}

// View renders the current screen.
func (m *Model) View() string {
	if m.screen == screenHome {
		return screens.Home(m.cursor)
	}
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(i18n.HardwareTitle))
	b.WriteString("\n\n")
	switch {
	case m.errorText != "":
		b.WriteString("Erro: " + m.errorText)
	case len(m.devices) == 0:
		b.WriteString(i18n.EmptyInventory)
	default:
		for _, d := range m.devices {
			fmt.Fprintf(&b, "%s  %s  %s  %d bytes", d.Path, d.Kind, d.ID, d.SizeBytes)
			if d.Mounted {
				b.WriteString("  (montado)")
			}
			b.WriteByte('\n')
		}
	}
	b.WriteString("\n\n")
	b.WriteString(i18n.Navigation)
	return b.String()
}
