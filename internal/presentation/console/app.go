// Package console provides a line-oriented interface for basic recovery information.
package console

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// Run serves the numeric menu using only line-oriented input and output.
func Run(
	input io.Reader,
	output io.Writer,
	inventory func(context.Context) ([]device.Device, error),
) error {
	in := bufio.NewScanner(input)
	for {
		if _, err := fmt.Fprintln(
			output,
			"RK322x Multitool\n1. Informações de hardware\n0. Sair",
		); err != nil {
			return err
		}
		if _, err := fmt.Fprint(output, "Escolha: "); err != nil {
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
				if _, writeErr := fmt.Fprintf(output, "Erro: %v\n", err); writeErr != nil {
					return writeErr
				}
				continue
			}
			if len(devices) == 0 {
				if _, err := fmt.Fprintln(output, "Nenhum dispositivo encontrado."); err != nil {
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
			if _, err := fmt.Fprintln(output, "Opção inválida."); err != nil {
				return err
			}
		}
	}
}

// Confirm returns true only when the user explicitly enters y.
func Confirm(input io.Reader) bool {
	line, err := bufio.NewReader(input).ReadString('\n')
	return (err == nil || err == io.EOF) && strings.EqualFold(strings.TrimSpace(line), "y")
}
