// Package console provides a line-oriented interface for basic recovery information.
package console

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

// Run serves the numeric menu using only line-oriented input and output.
func Run(
	input io.Reader,
	output io.Writer,
	inventory func(context.Context) ([]device.Device, error),
	catalog i18n.Catalog,
) error {
	in := bufio.NewScanner(input)
	for {
		if _, err := fmt.Fprintf(
			output,
			"%s\n1. %s\n0. %s\n",
			catalog.ConsoleTitle,
			catalog.ConsoleHardwareOption,
			catalog.ConsoleExitOption,
		); err != nil {
			return err
		}
		if _, err := fmt.Fprint(output, catalog.ConsolePrompt); err != nil {
			return err
		}
		if !in.Scan() {
			if err := in.Err(); err != nil {
				return err
			}
			return nil
		}
		switch strings.TrimSpace(in.Text()) {
		case "0":
			return nil
		case "1":
			devices, err := inventory(context.Background())
			if err != nil {
				if _, writeErr := fmt.Fprintf(
					output,
					"%s%v\n",
					catalog.ErrorPrefix,
					err,
				); writeErr != nil {
					return writeErr
				}
				continue
			}
			if len(devices) == 0 {
				if _, err := fmt.Fprintln(output, catalog.EmptyInventory); err != nil {
					return err
				}
				continue
			}
			for _, d := range devices {
				if _, err := fmt.Fprintf(
					output,
					"%s (%s) — %d bytes\n",
					d.Path,
					d.Kind,
					d.SizeBytes,
				); err != nil {
					return err
				}
			}
		default:
			if _, err := fmt.Fprintln(output, catalog.InvalidOption); err != nil {
				return err
			}
		}
	}
}
