package image

import (
	"errors"
	"testing"
)

func TestFormatFromPath(t *testing.T) {
	for path, want := range map[string]Format{"disk.img": FormatRaw, "disk.img.xz": FormatXZ} {
		got, err := FormatFromPath(path)
		if err != nil || got != want {
			t.Errorf("FormatFromPath(%q) = %q, %v", path, got, err)
		}
	}
	if _, err := FormatFromPath("disk.zip"); err == nil {
		t.Fatal("unsupported format accepted")
	}
}

func TestValidateSize(t *testing.T) {
	if err := (Image{SizeBytes: 10}).ValidateSize(10); err != nil {
		t.Fatal(err)
	}
	if err := (Image{SizeBytes: 10}).ValidateSize(11); !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("error = %v", err)
	}
}
