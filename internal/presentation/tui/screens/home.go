// Package screens renders individual terminal-interface screens.
package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

// Home renders the main menu with its selected item highlighted.
func Home(selected int) string {
	var rows []string
	for i, label := range i18n.MainMenu {
		if i == selected {
			rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render("› "+label))
		} else {
			rows = append(rows, "  "+label)
		}
	}
	return lipgloss.NewStyle().Bold(true).Render(i18n.Title) + "\n\n" + strings.Join(rows, "\n") + "\n\n" + i18n.Navigation
}
