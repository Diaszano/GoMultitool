package screens

import (
	"fmt"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Confirmation requires an explicit key press before a destructive operation.
type Confirmation struct {
	Device    device.Device
	Confirmed bool
}

// NewConfirmation starts with approval disabled.
func NewConfirmation(d device.Device) Confirmation { return Confirmation{Device: d} }

// Update accepts only an explicit Y key and rejects Escape or N.
func (c *Confirmation) Update(msg tea.Msg) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return
	}
	if key.Type == tea.KeyEsc || key.Type == tea.KeyRunes && strings.EqualFold(string(key.Runes), "n") {
		c.Reject()
	} else if key.Type == tea.KeyRunes && strings.EqualFold(string(key.Runes), "y") {
		c.Accept()
	}
}

// Accept records explicit approval.
func (c *Confirmation) Accept() { c.Confirmed = true }

// Reject clears approval.
func (c *Confirmation) Reject() { c.Confirmed = false }

// View renders the exact device path and identity requiring approval.
func (c Confirmation) View() string {
	return lipgloss.NewStyle().Bold(true).Render("Confirmar operação destrutiva?\n") + styles.Warning.Render(fmt.Sprintf("Dispositivo: %s (%s)\nCapacidade: %d bytes", c.Device.Path, c.Device.ID, c.Device.SizeBytes)) + "\n\nDigite Y para confirmar; qualquer outra resposta mantém a operação cancelada.\n" + i18n.Navigation
}
