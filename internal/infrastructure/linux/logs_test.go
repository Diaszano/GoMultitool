package linux

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAbsentJournalAndSensitiveRedaction(t *testing.T) {
	if _, available, err := ReadJournal(
		filepath.Join(t.TempDir(), "missing"),
		1024,
	); err != nil ||
		available {
		t.Fatalf("available=%v error=%v", available, err)
	}
	input := "password=supersecret token: abc123 PASSWORD=\"two words\"\n-----BEGIN RSA PRIVATE KEY-----\nprivate-data\n-----END RSA PRIVATE KEY-----"
	out := RedactLog(input)
	for _, secret := range []string{"supersecret", "abc123", "two words", "private-data"} {
		if strings.Contains(out, secret) {
			t.Fatalf("secret leaked: %s", out)
		}
	}
}
