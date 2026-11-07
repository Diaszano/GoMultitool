package restore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type testCatalog struct{ current device.Device }

func (c testCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.current}, nil
}

func (c testCatalog) Resolve(_ context.Context, path string) (device.Device, error) {
	if path != c.current.Path {
		return device.Device{}, os.ErrNotExist
	}
	return c.current, nil
}

func testDevice() device.Device {
	return device.Device{
		ID:                "dst",
		Path:              "/dev/dst",
		Kind:              device.KindEMMC,
		SizeBytes:         512,
		LogicalSectorSize: 512,
	}
}

func testGuard() storageguard.Guard { return storageguard.Guard{Catalog: testCatalog{testDevice()}} }

func testManifest(data []byte) backup.Manifest {
	sum := sha256.Sum256(data)
	return backup.Manifest{
		Version:           1,
		SourceID:          "src",
		SourcePath:        "/dev/src",
		SizeBytes:         int64(len(data)),
		LogicalSectorSize: 512,
		SHA256:            hex.EncodeToString(sum[:]),
		Status:            backup.StatusComplete,
	}
}

func TestRestoreRejectsInvalidManifestBeforeOpeningTarget(t *testing.T) {
	opened := false
	err := (Service{}).Execute(
		context.Background(),
		backup.Manifest{},
		nil,
		storageguard.Plan{},
		false,
		storageguard.Guard{},
		func(string) (application.Target, error) { opened = true; return nil, nil },
	)
	if err == nil || opened {
		t.Fatalf("error=%v, opened=%v", err, opened)
	}
}

func TestRestoreRequiresConfirmation(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 512)
	err := (Service{}).Execute(
		context.Background(),
		testManifest(data),
		bytes.NewReader(data),
		storageguard.Plan{Device: testDevice()},
		false,
		testGuard(),
		nil,
	)
	if !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("error=%v", err)
	}
}

func TestRestoreRejectsTruncatedBackup(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 512)
	opened := false
	err := (Service{}).Execute(
		context.Background(),
		testManifest(data),
		bytes.NewReader([]byte("short")),
		storageguard.Plan{Device: testDevice()},
		true,
		testGuard(),
		func(string) (application.Target, error) { opened = true; return nil, nil },
	)
	if err == nil || opened {
		t.Fatalf("error=%v, opened=%v", err, opened)
	}
}

func TestRestoreWritesOnlyAfterVerification(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 512)
	path := t.TempDir() + "/target"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(512); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	opened := false
	err = (Service{}).Execute(
		context.Background(),
		testManifest(data),
		bytes.NewReader(data),
		storageguard.Plan{Device: testDevice()},
		true,
		testGuard(),
		func(got string) (application.Target, error) {
			opened = true
			if got != testDevice().Path {
				t.Fatalf("path=%q", got)
			}
			return os.OpenFile(path, os.O_RDWR, 0)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !opened {
		t.Fatal("target was not opened")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("readback differs")
	}
}

func TestRestoreRejectsSmallTargetBeforeOpening(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 512)
	dev := testDevice()
	dev.SizeBytes = 256
	opened := false
	err := (Service{}).Execute(
		context.Background(),
		testManifest(data),
		bytes.NewReader(data),
		storageguard.Plan{Device: dev},
		true,
		testGuard(),
		func(string) (application.Target, error) { opened = true; return nil, nil },
	)
	if err == nil || opened {
		t.Fatalf("error=%v, opened=%v", err, opened)
	}
}

func TestRestoreRevalidatesIdentityAfterOpening(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 512)
	dev := testDevice()
	catalog := &mutableCatalog{current: dev}
	guard := storageguard.Guard{Catalog: catalog}
	plan, err := guard.Plan(context.Background(), dev.Path)
	if err != nil {
		t.Fatal(err)
	}
	target, err := os.CreateTemp(t.TempDir(), "target")
	if err != nil {
		t.Fatal(err)
	}
	if err := target.Truncate(512); err != nil {
		t.Fatal(err)
	}
	err = (Service{}).Execute(
		context.Background(),
		testManifest(data),
		bytes.NewReader(data),
		plan,
		true,
		guard,
		func(string) (application.Target, error) { catalog.current.ID = "replacement"; return target, nil },
	)
	if !errors.Is(err, storageguard.ErrDeviceChanged) {
		t.Fatalf("error=%v", err)
	}
}

type mutableCatalog struct{ current device.Device }

func (c *mutableCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.current}, nil
}

func (c *mutableCatalog) Resolve(_ context.Context, path string) (device.Device, error) {
	if path != c.current.Path {
		return device.Device{}, os.ErrNotExist
	}
	return c.current, nil
}
