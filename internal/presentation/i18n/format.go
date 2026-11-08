package i18n

import (
	"fmt"
	"strings"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

// DeviceLine formats one device as "<path>  <kind>  <size>" plus status labels.
func (c Catalog) DeviceLine(d device.Device) string {
	kind, ok := c.Kinds[d.Kind]
	if !ok {
		kind = c.Kinds[device.KindUnknown]
	}
	line := fmt.Sprintf("%s  %s  %s", d.Path, kind, c.Size(d.SizeBytes))
	if d.Mounted {
		line += "  · " + c.InUseLabel
	}
	if d.BootMedia {
		line += "  · " + c.BootMediaLabel
	}
	return line
}

// Size formats bytes as decimal GB with one decimal, or whole MB below 1 GB.
func (c Catalog) Size(bytes int64) string {
	if bytes < 1_000_000_000 {
		return fmt.Sprintf("%d MB", bytes/1_000_000)
	}
	gb := fmt.Sprintf("%.1f", float64(bytes)/1e9)
	return strings.Replace(gb, ".", c.DecimalSeparator, 1) + " GB"
}
