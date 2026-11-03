package bootloader

import (
	"strings"
	"testing"
)

func TestInspectOnlyReportsKnownHash(t *testing.T) {
	got, err := Inspect(strings.NewReader("firmware"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SHA256) != 64 || got.Known {
		t.Fatalf("inspection=%+v", got)
	}
	got, err = Inspect(strings.NewReader("firmware"), []string{got.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Known {
		t.Fatal("known hash not identified")
	}
}
