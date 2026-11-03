package imageconfig

import (
	"strings"
	"testing"
)

func TestPresetValidationAndSecretRedaction(t *testing.T) {
	p := Preset{
		RootPassword:     "Root$;`pass",
		UserName:         "armbian",
		UserPassword:     "User$(x);pass",
		RealName:         "Armbian User",
		Hostname:         "rk3229",
		Locale:           "en_US.UTF-8",
		Timezone:         "Etc/UTC",
		SSHAuthorizedKey: "ssh-ed25519 AAAA test",
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	redacted := p.Redacted()
	for _, secret := range []string{p.RootPassword, p.UserPassword, p.SSHAuthorizedKey} {
		if strings.Contains(redacted, secret) {
			t.Fatalf("secret leaked in %q", redacted)
		}
	}
	p.UserName = "bad user"
	if err := p.Validate(); err == nil {
		t.Fatal("invalid username accepted")
	}
}
