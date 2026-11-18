package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/presentation/i18n"
)

type testCatalog struct {
	devices []device.Device
	err     error
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

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
		"/",
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
		"/",
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
		if code := run([]string{"version"}, "/", testCatalog{}, catalog, &out, &stderr); code != 0 {
			t.Fatalf("version exit = %d: %s", code, stderr.String())
		}
		if !bytes.Contains(out.Bytes(), []byte(catalog.VersionLabel)) ||
			!bytes.Contains(out.Bytes(), []byte(catalog.CommitLabel+":")) ||
			!bytes.Contains(out.Bytes(), []byte(catalog.BuildLabel+":")) {
			t.Errorf("locale %q version output = %q", locale, out.String())
		}
		out.Reset()
		if code := run(nil, "/", testCatalog{}, catalog, &out, &stderr); code != 2 ||
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
			"/",
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
			"/",
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

func TestJSONEncodeWriteErrorKeepsStableContext(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocalePtBR, i18n.LocaleEN} {
		var stderr bytes.Buffer
		code := run(
			[]string{"hardware", "--json"},
			"/",
			testCatalog{devices: []device.Device{{ID: "mmc0"}}},
			i18n.For(locale),
			failedWriter{},
			&stderr,
		)
		if code != 1 || stderr.String() != "encode hardware inventory: broken pipe\n" {
			t.Errorf("locale %q exit/stderr = %d/%q", locale, code, stderr.String())
		}
	}
}

func TestJSONEncodeErrorContextsRemainStable(t *testing.T) {
	for _, test := range []struct {
		context string
		want    string
	}{
		{"encode hardware inventory", "encode hardware inventory: broken pipe\n"},
		{"encode diagnostics", "encode diagnostics: broken pipe\n"},
	} {
		var stderr bytes.Buffer
		if code := encodeJSON(
			failedWriter{},
			&stderr,
			test.context,
			map[string]string{"ok": "yes"},
		); code != 1 || stderr.String() != test.want {
			t.Errorf("context %q exit/stderr = %d/%q", test.context, code, stderr.String())
		}
	}
}

func TestDiagnoseJSON(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"proc", "boot", "var/log"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]os.FileMode{
		"proc/cmdline":   0o600,
		"boot/zImage":    0o600,
		"var/log/syslog": 0o000,
	}
	for name, mode := range files {
		if err := os.WriteFile(
			filepath.Join(root, name),
			[]byte("root=/dev/mmcblk0p2\n"),
			mode,
		); err != nil {
			t.Fatal(err)
		}
	}
	var out, stderr bytes.Buffer
	code := run(
		[]string{"diagnose", "--json"},
		root,
		testCatalog{},
		i18n.For(i18n.LocalePtBR),
		&out,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["journal_available"]; ok || got["boot"] == nil || got["network"] == nil {
		t.Fatalf("unexpected report keys: %s", out.String())
	}
	var boot struct {
		KernelPresent bool `json:"kernel_present"`
	}
	if err := json.Unmarshal(got["boot"], &boot); err != nil {
		t.Fatal(err)
	}
	if !boot.KernelPresent {
		t.Fatalf("kernel not reported: %s", got["boot"])
	}
}
