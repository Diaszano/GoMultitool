package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProgressUpdatesAndCancels(t *testing.T) {
	p := Progress{Total: 100}
	p.Update(ProgressUpdate{Done: 25})
	if p.Done != 25 || !strings.Contains(p.View(), "25%") {
		t.Fatalf("progress=%+v view=%q", p, p.View())
	}
	p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !p.Canceled {
		t.Fatal("cancel not recorded")
	}
}
