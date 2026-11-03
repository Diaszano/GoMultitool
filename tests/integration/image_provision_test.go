package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/application/imageconfig"
)

type provisioningMount struct {
	root    string
	cleaned bool
}

func (m *provisioningMount) Mount(context.Context, string) (string, func() error, error) {
	return m.root, func() error { m.cleaned = true; return nil }, nil
}

func TestOfflineProvisioningPublishesOnlyCustomizedCopy(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "original.img")
	if err := os.WriteFile(source, []byte("original image"), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "prepared.img")
	mounter := &provisioningMount{root: root}
	profile := imageconfig.Compatibility{
		OSID:           "armbian",
		MinimumVersion: "26.08",
		SchemaVersion:  1,
		SupportedVariables: []string{
			"PRESET_ROOT_PASSWORD",
			"PRESET_USER_NAME",
			"PRESET_USER_PASSWORD",
			"PRESET_DEFAULT_REALNAME",
			"PRESET_LOCALE",
			"PRESET_TIMEZONE",
			"PRESET_NET_ETHERNET_ENABLED",
			"PRESET_NET_CHANGE_DEFAULTS",
		},
	}
	preset := imageconfig.Preset{
		RootPassword: "RootSecret123",
		UserName:     "armbian",
		UserPassword: "UserSecret123",
		RealName:     "Armbian User",
		Hostname:     "rk3229",
		Locale:       "en_US.UTF-8",
		Timezone:     "Etc/UTC",
	}
	if err := (imageconfig.Service{Profile: profile, Mounter: mounter}).Prepare(
		context.Background(),
		source,
		output,
		"26.08",
		preset,
	); err != nil {
		t.Fatal(err)
	}
	if !mounter.cleaned {
		t.Fatal("image remained mounted")
	}
	src, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(src) != "original image" || string(out) != string(src) {
		t.Fatal("source was modified or prepared copy differs")
	}
	secret, err := os.ReadFile(filepath.Join(root, "root/.not_logged_in_yet"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "root/.not_logged_in_yet"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 || len(secret) == 0 {
		t.Fatal("first-boot secret file is missing or exposed")
	}
}
