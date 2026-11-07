package rk322x

import (
	"errors"
	"testing"
)

func TestFindCommandRate(t *testing.T) {
	sig := []byte{0x65, 0x65, 0x86, 0xa7}
	build := func(size, sigAt, rateAt int, rate byte) []byte {
		b := make([]byte, size)
		if sigAt >= 0 {
			copy(b[sigAt:], sig)
		}
		if rateAt >= 0 {
			b[rateAt] = rate
		}
		return b
	}
	tests := []struct {
		name    string
		head    []byte
		offset  int64
		rate    byte
		wantErr error
	}{
		{"2T", build(CommandRateScan, 4096, 4080, 1), 4080, 1, nil},
		{"1T", build(CommandRateScan, 4096, 4080, 0), 4080, 0, nil},
		{"unaligned", build(CommandRateScan, 4098, 4082, 1), 0, 0, ErrNoCommandRate},
		{"absent", build(CommandRateScan, -1, -1, 0), 0, 0, ErrNoCommandRate},
		{"bad rate byte", build(CommandRateScan, 4096, 4080, 7), 0, 0, ErrNoCommandRate},
		{
			"past window", build(CommandRateScan+4096, CommandRateScan, CommandRateScan-16, 1),
			0, 0, ErrNoCommandRate,
		},
		{"too early", build(64, 8, -1, 0), 0, 0, ErrNoCommandRate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			off, rate, err := FindCommandRate(tt.head)
			if !errors.Is(err, tt.wantErr) || off != tt.offset || rate != tt.rate {
				t.Fatalf("got (%d, %d, %v), want (%d, %d, %v)",
					off, rate, err, tt.offset, tt.rate, tt.wantErr)
			}
		})
	}
}
