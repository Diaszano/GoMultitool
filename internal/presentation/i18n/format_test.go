package i18n

import (
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestDeviceLineLocalizesAndLabels(t *testing.T) {
	d := device.Device{
		Path: "/dev/mmcblk0", Kind: device.KindSD, SizeBytes: 14_800_000_000,
		Mounted: true, BootMedia: true,
	}
	want := "/dev/mmcblk0  SD  14,8 GB  · em uso  · mídia de boot"
	if got := For(LocalePtBR).DeviceLine(d); got != want {
		t.Errorf("pt-BR = %q, want %q", got, want)
	}
	d.Mounted, d.BootMedia = false, false
	if got := For(LocaleEN).DeviceLine(d); got != "/dev/mmcblk0  SD  14.8 GB" {
		t.Errorf("en = %q", got)
	}
}

func TestSizeUnderOneGB(t *testing.T) {
	if got := For(LocalePtBR).Size(4_194_304); got != "4 MB" {
		t.Errorf("Size = %q", got)
	}
}
