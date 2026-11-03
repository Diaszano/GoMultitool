package screens

import (
	"strings"
	"testing"
)

func TestHomeShowsBrazilianPortugueseLabelsAndSelection(t *testing.T) {
	view := Home(0)
	if !strings.Contains(view, "Informações de hardware") ||
		!strings.Contains(view, "› Informações de hardware") {
		t.Fatalf("view=%q", view)
	}
}
