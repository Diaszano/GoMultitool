package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ProgressUpdate reports the number of completed bytes.
type ProgressUpdate struct{ Done int64 }

// Progress displays operation progress and tracks cancellation.
type Progress struct {
	Done     int64
	Total    int64
	Canceled bool
}

// Update applies byte counts or keyboard cancellation.
func (p *Progress) Update(msg tea.Msg) {
	switch v := msg.(type) {
	case ProgressUpdate:
		p.Done = v.Done
		if p.Done < 0 {
			p.Done = 0
		}
		if p.Total > 0 && p.Done > p.Total {
			p.Done = p.Total
		}
	case tea.KeyMsg:
		if v.Type == tea.KeyEsc || v.Type == tea.KeyCtrlC {
			p.Cancel()
		}
	}
}

// Cancel marks the operation canceled.
func (p *Progress) Cancel() { p.Canceled = true }

// View renders a compact byte-count progress bar.
func (p Progress) View() string {
	if p.Total <= 0 {
		return "Sem dados"
	}
	percent := int(float64(p.Done) * 100 / float64(p.Total))
	return fmt.Sprintf("%s %d%% (%d/%d)", strings.Repeat("█", percent/10)+strings.Repeat("░", 10-percent/10), percent, p.Done, p.Total)
}
