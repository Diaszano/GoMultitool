// Package filesystem performs read-only partition and filesystem signature checks.
package filesystem

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrInvalidPartitionTable indicates a missing signature or out-of-range entry.
var ErrInvalidPartitionTable = errors.New("invalid MBR partition table")

// Partition records one bounded MBR entry.
type Partition struct {
	Number      int
	Type        byte
	StartSector uint64
	SectorCount uint64
}

// ParseMBR validates the MBR signature and partition bounds without writing.
func ParseMBR(r io.ReaderAt, diskBytes int64) ([]Partition, error) {
	if diskBytes < 512 {
		return nil, ErrInvalidPartitionTable
	}
	sector := make([]byte, 512)
	if _, err := r.ReadAt(sector, 0); err != nil {
		return nil, err
	}
	if sector[510] != 0x55 || sector[511] != 0xaa {
		return nil, ErrInvalidPartitionTable
	}
	var partitions []Partition
	for i := 0; i < 4; i++ {
		entry := sector[446+i*16 : 462+i*16]
		typ := entry[4]
		start := uint64(binary.LittleEndian.Uint32(entry[8:12]))
		count := uint64(binary.LittleEndian.Uint32(entry[12:16]))
		if typ == 0 {
			continue
		}
		if count == 0 || start+count < start || (start+count)*512 > uint64(diskBytes) {
			return nil, ErrInvalidPartitionTable
		}
		partitions = append(
			partitions,
			Partition{Number: i + 1, Type: typ, StartSector: start, SectorCount: count},
		)
	}
	return partitions, nil
}

// HasExt4Signature checks for the ext4 superblock magic at its fixed offset.
func HasExt4Signature(r io.ReaderAt) bool {
	var magic [2]byte
	_, err := r.ReadAt(magic[:], 1080)
	return err == nil && magic[0] == 0x53 && magic[1] == 0xef
}
