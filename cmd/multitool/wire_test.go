package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCountingReaderReportsProgress(t *testing.T) {
	var calls [][2]int64
	r := &countingReader{
		r:     strings.NewReader("abcdef"),
		total: 6,
		p:     func(done, total int64) { calls = append(calls, [2]int64{done, total}) },
	}
	buf := make([]byte, 4)
	for {
		if _, err := r.Read(buf); errors.Is(err, io.EOF) {
			break
		}
	}
	if len(calls) != 2 || calls[0] != [2]int64{4, 6} || calls[1] != [2]int64{6, 6} {
		t.Fatalf("progress calls = %v, want [[4 6] [6 6]]", calls)
	}
}

func TestBackupDestinationRejectsPathNames(t *testing.T) {
	for _, name := range []string{"", "a/b", ".x", ".."} {
		if _, err := backupDestination("/mt/backups", name); err == nil {
			t.Errorf("backupDestination(%q) accepted an unsafe name", name)
		}
	}
	got, err := backupDestination("/mt/backups", "tv box")
	if err != nil || got != "/mt/backups/tv box.img.gz" {
		t.Fatalf("backupDestination = %q, %v", got, err)
	}
}
