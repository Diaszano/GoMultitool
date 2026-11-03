package console

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestNumericMenuAndInvalidSelection(t *testing.T) {
	var out bytes.Buffer
	err := Run(
		strings.NewReader("x\n1\n0\n"),
		&out,
		func(context.Context) ([]device.Device, error) {
			return []device.Device{
				{ID: "mmc0", Path: "/dev/mmc0", Kind: device.KindEMMC, SizeBytes: 512},
			}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Opção inválida") ||
		!strings.Contains(out.String(), "/dev/mmc0") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestConfirmationDefaultsNo(t *testing.T) {
	if Confirm(strings.NewReader("n\n")) {
		t.Fatal("negative response confirmed")
	}
	if Confirm(strings.NewReader("\n")) {
		t.Fatal("empty response confirmed")
	}
	if !Confirm(strings.NewReader("y\n")) {
		t.Fatal("explicit yes rejected")
	}
}
