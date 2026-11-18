package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/restore"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/backup"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type restoreCatalog struct{ d device.Device }

func (c restoreCatalog) List(context.Context) ([]device.Device, error) {
	return []device.Device{c.d}, nil
}

func (c restoreCatalog) Resolve(_ context.Context, p string) (device.Device, error) {
	if p != c.d.Path {
		return device.Device{}, os.ErrNotExist
	}
	return c.d, nil
}

func TestRestoreTemporaryFileReadback(t *testing.T) {
	data := bytes.Repeat([]byte("r"), 512)
	source := t.TempDir() + "/source"
	if err := os.WriteFile(source, data, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	manifest := backup.Manifest{
		Version:           1,
		SourceID:          "source",
		SourcePath:        "/dev/source",
		SizeBytes:         512,
		LogicalSectorSize: 512,
		SHA256:            hex.EncodeToString(sum[:]),
		Status:            backup.StatusComplete,
	}
	path := t.TempDir() + "/target"
	target, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = target.Truncate(512); err != nil {
		t.Fatal(err)
	}
	if err = target.Close(); err != nil {
		t.Fatal(err)
	}
	dev := device.Device{
		ID:                "target",
		Path:              "/dev/target",
		Kind:              device.KindUSB,
		SizeBytes:         512,
		LogicalSectorSize: 512,
	}
	guard := storageguard.Guard{Catalog: restoreCatalog{dev}}
	plan, err := guard.Plan(context.Background(), dev.Path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = src.Close() }()
	err = (restore.Service{}).Execute(
		context.Background(),
		manifest,
		src,
		plan,
		true,
		guard,
		func(string) (application.Target, error) { return os.OpenFile(path, os.O_RDWR, 0) },
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("restored data does not match")
	}
}
