package console

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
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
		i18n.For(i18n.LocalePtBR),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Opção inválida") ||
		!strings.Contains(out.String(), "/dev/mmc0") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestNumericMenuUsesCatalog(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		t.Run(string(locale), func(t *testing.T) {
			catalog := i18n.For(locale)
			var out bytes.Buffer
			err := Run(
				strings.NewReader("x\n1\n0\n"),
				&out,
				func(context.Context) ([]device.Device, error) {
					return []device.Device{{
						Path:      "/dev/mmc0",
						Kind:      device.KindEMMC,
						SizeBytes: 512,
					}}, nil
				},
				catalog,
			)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{catalog.ConsoleTitle, catalog.ConsoleHardwareOption, catalog.ConsoleExitOption, catalog.ConsolePrompt, catalog.InvalidOption, "/dev/mmc0"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("output missing %q: %q", want, out.String())
				}
			}
		})
	}
}

func TestInventoryErrorUsesCatalogContext(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		var out bytes.Buffer
		err := Run(
			strings.NewReader("1\n0\n"),
			&out,
			func(context.Context) ([]device.Device, error) {
				return nil, errors.New("sysfs unavailable")
			},
			i18n.For(locale),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), i18n.For(locale).ErrorPrefix+"sysfs unavailable") {
			t.Errorf("locale %q output = %q", locale, out.String())
		}
	}
}
