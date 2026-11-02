package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type testCatalog struct {
	devices []device.Device
	err     error
}

func (c testCatalog) List(context.Context) ([]device.Device, error) { return c.devices, c.err }
func (c testCatalog) Resolve(context.Context, string) (device.Device, error) {
	return device.Device{}, nil
}

func TestHardwareJSON(t *testing.T) {
	var out, stderr bytes.Buffer
	catalog := testCatalog{devices: []device.Device{{ID: "sda", Kind: device.KindUSB}, {ID: "mmc0", Kind: device.KindEMMC}}}
	if code := run([]string{"hardware", "--json"}, catalog, &out, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var got []device.Device
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got[0].ID != "mmc0" || got[1].ID != "sda" {
		t.Fatalf("devices = %+v", got)
	}
}

func TestHardwareJSONErrorIsStructured(t *testing.T) {
	var out, stderr bytes.Buffer
	if code := run([]string{"hardware", "--json"}, testCatalog{err: errors.New("sysfs unavailable")}, &out, &stderr); code == 0 {
		t.Fatal("expected failure")
	}
	var got map[string]string
	if err := json.Unmarshal(stderr.Bytes(), &got); err != nil {
		t.Fatalf("error is not JSON: %v", err)
	}
	if got["error"] == "" {
		t.Fatalf("missing error: %v", got)
	}
}
