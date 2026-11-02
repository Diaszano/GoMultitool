package imageconfig

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRenderArmbianPresetsAndRejectsUnsupportedVersionsOrVariables(t *testing.T) {
	profile := Compatibility{OSID: "armbian", MinimumVersion: "26.08", SchemaVersion: 1, SupportedVariables: []string{"PRESET_ROOT_PASSWORD", "PRESET_USER_NAME", "PRESET_USER_PASSWORD", "PRESET_DEFAULT_REALNAME", "PRESET_LOCALE", "PRESET_TIMEZONE", "PRESET_NET_ETHERNET_ENABLED", "PRESET_NET_CHANGE_DEFAULTS"}}
	p := Preset{RootPassword: "Root$;pass", UserName: "armbian", UserPassword: "User$(x)", RealName: "Armbian User", Hostname: "rk3229", Locale: "en_US.UTF-8", Timezone: "Etc/UTC", SSHAuthorizedKey: "ssh-ed25519 AAAA test"}
	result, err := Render(profile, "26.08.3", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result.FirstBoot), `PRESET_ROOT_PASSWORD="Root\$;pass"`) || !strings.Contains(string(result.FirstBoot), `PRESET_NET_ETHERNET_ENABLED="1"`) {
		t.Fatalf("config=%s", result.FirstBoot)
	}
	if result.Files["root/.not_logged_in_yet"].Mode != 0600 || result.Files["home/armbian/.ssh/authorized_keys"].Mode != 0600 || string(result.Files["etc/hostname"].Data) != "rk3229\n" {
		t.Fatalf("artifacts=%+v", result.Files)
	}
	if _, err := Render(profile, "26.05", p); !errors.Is(err, ErrUnsupportedImage) {
		t.Fatalf("error=%v", err)
	}
	profile.SupportedVariables = nil
	if _, err := Render(profile, "26.08", p); !errors.Is(err, ErrUnsupportedVariable) {
		t.Fatalf("error=%v", err)
	}
}

func TestCompatibilityFileParsesAndUnsupportedOSRejected(t *testing.T) {
	f, err := os.Open("../../../configs/installers/armbian.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	profile, err := LoadCompatibility(f)
	if err != nil {
		t.Fatal(err)
	}
	profile.OSID = "debian"
	if _, err := Render(profile, "26.08", Preset{}); !errors.Is(err, ErrUnsupportedImage) {
		t.Fatalf("error=%v", err)
	}
}
