package backup

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestManifestJSONRoundTrip(t *testing.T) {
	m := Manifest{Version: 1, SourceID: "mmc0", SourcePath: "/dev/mmc0", SizeBytes: 1024, LogicalSectorSize: 512, SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Status: StatusComplete}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != m {
		t.Fatalf("round trip = %+v", got)
	}
}

func TestManifestRejectsIncompleteAndInvalidChecksums(t *testing.T) {
	base := Manifest{Version: 1, SourceID: "mmc0", SourcePath: "/dev/mmc0", SizeBytes: 1024, LogicalSectorSize: 512, SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Status: StatusComplete}
	for _, tc := range []struct {
		name string
		edit func(*Manifest)
	}{
		{"missing identity", func(m *Manifest) { m.SourceID = "" }},
		{"bad checksum", func(m *Manifest) { m.SHA256 = "bad" }},
		{"incomplete", func(m *Manifest) { m.Status = StatusPending }},
		{"zero size", func(m *Manifest) { m.SizeBytes = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			tc.edit(&m)
			if !errors.Is(m.Validate(), ErrInvalidManifest) {
				t.Fatalf("Validate() = %v", m.Validate())
			}
		})
	}
}
