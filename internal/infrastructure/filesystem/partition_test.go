package filesystem

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseMBR(t *testing.T) {
	disk := make([]byte, 4*1024*1024)
	disk[510], disk[511] = 0x55, 0xaa
	entry := disk[446:462]
	entry[4] = 0x83
	binary.LittleEndian.PutUint32(entry[8:12], 2048)
	binary.LittleEndian.PutUint32(entry[12:16], 4096)
	parts, err := ParseMBR(bytes.NewReader(disk), int64(len(disk)))
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 || parts[0].StartSector != 2048 || parts[0].SectorCount != 4096 {
		t.Fatalf("partitions=%+v", parts)
	}
}
func TestParseMBRRejectsMalformedTable(t *testing.T) {
	disk := make([]byte, 512)
	if _, err := ParseMBR(bytes.NewReader(disk), 512); err == nil {
		t.Fatal("missing MBR signature accepted")
	}
	disk[510], disk[511] = 0x55, 0xaa
	entry := disk[446:462]
	entry[4] = 0x83
	binary.LittleEndian.PutUint32(entry[8:12], 2048)
	binary.LittleEndian.PutUint32(entry[12:16], 1)
	if _, err := ParseMBR(bytes.NewReader(disk), 512); err == nil {
		t.Fatal("out-of-range partition accepted")
	}
}
func TestHasExt4Signature(t *testing.T) {
	data := make([]byte, 2048)
	data[1080], data[1081] = 0x53, 0xef
	if !HasExt4Signature(bytes.NewReader(data)) {
		t.Fatal("ext4 signature not found")
	}
	if HasExt4Signature(bytes.NewReader(make([]byte, 2048))) {
		t.Fatal("missing ext4 signature accepted")
	}
}
