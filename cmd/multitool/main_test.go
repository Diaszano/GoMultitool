package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
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
	catalog := testCatalog{
		devices: []device.Device{
			{ID: "sda", Kind: device.KindUSB},
			{ID: "mmc0", Kind: device.KindEMMC},
		},
	}
	if code := run(
		[]string{"hardware", "--json"},
		catalog,
		i18n.For(i18n.LocalePtBR),
		&out,
		&stderr,
	); code != 0 {
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
	if code := run(
		[]string{"hardware", "--json"},
		testCatalog{err: errors.New("sysfs unavailable")},
		i18n.For(i18n.LocalePtBR),
		&out,
		&stderr,
	); code == 0 {
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

func TestVersionAndUsageUseCatalog(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		catalog := i18n.For(locale)
		var out, stderr bytes.Buffer
		if code := run([]string{"version"}, testCatalog{}, catalog, &out, &stderr); code != 0 {
			t.Fatalf("version exit = %d: %s", code, stderr.String())
		}
		if !bytes.Contains(out.Bytes(), []byte(catalog.VersionLabel)) ||
			!bytes.Contains(out.Bytes(), []byte(catalog.CommitLabel+":")) ||
			!bytes.Contains(out.Bytes(), []byte(catalog.BuildLabel+":")) {
			t.Errorf("locale %q version output = %q", locale, out.String())
		}
		out.Reset()
		if code := run(nil, testCatalog{}, catalog, &out, &stderr); code != 2 ||
			!bytes.Contains(stderr.Bytes(), []byte(catalog.Usage)) {
			t.Errorf("locale %q usage exit/output = %d/%q", locale, code, stderr.String())
		}
	}
}

func TestJSONOutputDoesNotDependOnLocale(t *testing.T) {
	var outputs [2][]byte
	for i, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		var out, stderr bytes.Buffer
		if code := run(
			[]string{"hardware", "--json"},
			testCatalog{devices: []device.Device{{ID: "mmc0"}}},
			i18n.For(locale),
			&out,
			&stderr,
		); code != 0 {
			t.Fatalf("locale %q exit = %d: %s", locale, code, stderr.String())
		}
		outputs[i] = out.Bytes()
	}
	if string(outputs[0]) != string(outputs[1]) {
		t.Fatalf("JSON varies by locale: %q != %q", outputs[0], outputs[1])
	}
}

func TestJSONErrorDoesNotDependOnLocale(t *testing.T) {
	var outputs [2][]byte
	for i, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		var out, stderr bytes.Buffer
		code := run(
			[]string{"hardware", "--json"},
			testCatalog{err: errors.New("sysfs unavailable")},
			i18n.For(locale),
			&out,
			&stderr,
		)
		if code != 1 || out.Len() != 0 {
			t.Fatalf("locale %q exit/stdout = %d/%q", locale, code, out.String())
		}
		var got map[string]string
		if err := json.Unmarshal(stderr.Bytes(), &got); err != nil ||
			got["error"] != "sysfs unavailable" {
			t.Fatalf("locale %q stderr = %q, decode err = %v", locale, stderr.String(), err)
		}
		outputs[i] = stderr.Bytes()
	}
	if string(outputs[0]) != string(outputs[1]) {
		t.Fatalf("JSON error varies by locale: %q != %q", outputs[0], outputs[1])
	}
}
