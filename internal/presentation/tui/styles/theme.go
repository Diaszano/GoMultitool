// Package styles defines shared presentation styles.
package styles

import "github.com/charmbracelet/lipgloss"

// Warning highlights destructive-operation details.
var Warning = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
