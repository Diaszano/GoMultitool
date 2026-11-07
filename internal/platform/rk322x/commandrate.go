package rk322x

import (
	"bytes"
	"errors"
)

// CommandRateScan is how many leading bytes of the loader are searched.
const CommandRateScan = 128 * 1024

// ErrNoCommandRate indicates the DDR command rate byte was not found.
var ErrNoCommandRate = errors.New("command rate not found")

var commandRateSignature = []byte{0x65, 0x65, 0x86, 0xa7}

// FindCommandRate locates the DDR command rate byte (0 = 1T, 1 = 2T): it sits
// 16 bytes before the first 4-aligned signature in the scan window.
func FindCommandRate(head []byte) (int64, byte, error) {
	head = head[:min(len(head), CommandRateScan)]
	for s := 0; s+len(commandRateSignature) <= len(head); s += 4 {
		if !bytes.Equal(head[s:s+4], commandRateSignature) {
			continue
		}
		if s < 16 || head[s-16] > 1 {
			break
		}
		return int64(s - 16), head[s-16], nil
	}
	return 0, 0, ErrNoCommandRate
}
