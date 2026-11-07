package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCopyExactCopiesExactlySize(t *testing.T) {
	var dst bytes.Buffer
	if err := CopyExact(context.Background(), &dst, strings.NewReader("abcdef"), 4); err != nil {
		t.Fatal(err)
	}
	if dst.String() != "abcd" {
		t.Fatalf("dst=%q", dst.String())
	}
}

func TestCopyExactShortSource(t *testing.T) {
	err := CopyExact(context.Background(), io.Discard, strings.NewReader("ab"), 4)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("err=%v", err)
	}
}

func TestCopyExactStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var dst bytes.Buffer
	err := CopyExact(ctx, &dst, strings.NewReader("abcd"), 4)
	if !errors.Is(err, context.Canceled) || dst.Len() != 0 {
		t.Fatalf("err=%v dst=%q", err, dst.String())
	}
}

type eofWriter struct{}

func (eofWriter) Write([]byte) (int, error) { return 0, io.EOF }

func TestCopyExactKeepsWriterEOF(t *testing.T) {
	err := CopyExact(context.Background(), eofWriter{}, strings.NewReader("data"), 4)
	if !errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v", err)
	}
}
