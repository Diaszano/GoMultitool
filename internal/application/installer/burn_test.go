package installer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
)

const burnSize = 16384

var burnSignature = []byte{0x65, 0x65, 0x86, 0xa7}

// burnImage returns burnSize bytes; rate < 0 means no signature.
func burnImage(sigAt int, rate int) []byte {
	b := bytes.Repeat([]byte{0xff}, burnSize)
	if rate >= 0 {
		copy(b[sigAt:], burnSignature)
		b[sigAt-16] = byte(rate)
	}
	return b
}

func burnSetup(
	t *testing.T,
	targetSize int,
	targetData, imageData []byte,
) (Burn, installation.Plan, string, *int) {
	t.Helper()
	path := t.TempDir() + "/target"
	if err := os.WriteFile(path, targetData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, int64(targetSize)); err != nil {
		t.Fatal(err)
	}
	d := device.Device{
		ID: "usb", Path: "/dev/sdb", Kind: device.KindUSB,
		SizeBytes: int64(targetSize), LogicalSectorSize: 512,
	}
	opens := 0
	b := Burn{
		Guard: storageguard.Guard{Catalog: writerCatalog{d}},
		OpenTarget: func(string) (application.Target, error) {
			opens++
			return os.OpenFile(path, os.O_RDWR, 0)
		},
	}
	plan := installation.Plan{
		Image: image.Image{
			Path:      "i.img",
			Format:    image.FormatRaw,
			SizeBytes: int64(len(imageData)),
		},
		Target:  d,
		Summary: "write",
	}
	return b, plan, path, &opens
}

func TestBurnRestoresCommandRate(t *testing.T) {
	img := burnImage(8192, 0)
	b, plan, path, _ := burnSetup(t, burnSize, burnImage(4096, 1), img)
	b.Platform = "rk322x"
	res, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if err != nil || res.CommandRateNotApplied {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	got, _ := os.ReadFile(path)
	if got[8176] != 1 {
		t.Fatalf("rate=%d", got[8176])
	}
	got[8176] = 0
	if !bytes.Equal(got, img) {
		t.Fatal("rest of image differs")
	}
}

func TestBurnWarnsWhenImageHasNoCommandRate(t *testing.T) {
	img := burnImage(0, -1)
	b, plan, _, _ := burnSetup(t, burnSize, burnImage(4096, 1), img)
	b.Platform = "rk322x"
	res, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if err != nil || !res.CommandRateNotApplied {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestBurnSkipsCommandRateOffRK322x(t *testing.T) {
	img := burnImage(8192, 0)
	b, plan, path, _ := burnSetup(t, burnSize, burnImage(4096, 1), img)
	res, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if err != nil || res.CommandRateNotApplied {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, img) {
		t.Fatal("target differs from image")
	}
}

func TestBurnRejectsImageLargerThanTarget(t *testing.T) {
	img := bytes.Repeat([]byte{1}, 1024)
	b, plan, _, opens := burnSetup(t, 512, nil, img)
	b.Platform = "rk322x"
	_, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if !errors.Is(err, installation.ErrImageTooLarge) || *opens != 0 {
		t.Fatalf("err=%v opens=%d", err, *opens)
	}
}

type swapCatalog struct {
	d       device.Device
	swapped *bool
}

func (c swapCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.d}, nil
}

func (c swapCatalog) Resolve(context.Context, string) (device.Device, error) {
	d := c.d
	if *c.swapped {
		d.ID = "other"
	}
	return d, nil
}

type closeHook struct {
	application.Target
	onClose func()
}

func (h closeHook) Close() error { h.onClose(); return h.Target.Close() }

func TestBurnRevalidatesBeforeRestoringRate(t *testing.T) {
	img := burnImage(8192, 0)
	b, plan, path, _ := burnSetup(t, burnSize, burnImage(4096, 1), img)
	b.Platform = "rk322x"
	swapped := false
	b.Guard = storageguard.Guard{Catalog: swapCatalog{d: plan.Target, swapped: &swapped}}
	open, opens := b.OpenTarget, 0
	b.OpenTarget = func(p string) (application.Target, error) {
		opens++
		tg, err := open(p)
		if opens == 2 { // Execute's open: the device is swapped once it closes
			return closeHook{tg, func() { swapped = true }}, err
		}
		return tg, err
	}
	_, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if !errors.Is(err, storageguard.ErrDeviceChanged) {
		t.Fatalf("err=%v", err)
	}
	got, _ := os.ReadFile(path)
	if got[8176] != 0 {
		t.Fatalf("byte written to swapped device: %d", got[8176])
	}
}

func TestBurnSkipsRateWhenPreReadFails(t *testing.T) {
	img := burnImage(8192, 0)
	b, plan, path, _ := burnSetup(t, burnSize, burnImage(4096, 1), img)
	b.Platform = "rk322x"
	open, opens := b.OpenTarget, 0
	b.OpenTarget = func(p string) (application.Target, error) {
		opens++
		if opens == 1 {
			return nil, errors.New("open failed")
		}
		return open(p)
	}
	res, err := b.Run(context.Background(), plan, bytes.NewReader(img))
	if err != nil || res.CommandRateNotApplied {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, img) {
		t.Fatal("target differs from image")
	}
}
