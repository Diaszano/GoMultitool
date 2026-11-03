package rk322x

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

func TestMatchKnownUnknownAndAmbiguousProfiles(t *testing.T) {
	profiles := []Profile{
		{
			Version:      1,
			Name:         "reference",
			Compatible:   []string{"rockchip,rk3229"},
			StorageKinds: []device.Kind{device.KindEMMC},
		},
	}
	got, err := Match(profiles, []string{"rockchip,rk3229"})
	if err != nil || got.Name != "reference" {
		t.Fatalf("profile=%+v error=%v", got, err)
	}
	if _, err := Match(profiles, []string{"vendor,unknown"}); !errors.Is(err, ErrUnknownBoard) {
		t.Fatalf("error=%v", err)
	}
	profiles = append(
		profiles,
		Profile{Version: 1, Name: "ambiguous", Compatible: []string{"rockchip,rk3229"}},
	)
	if _, err := Match(profiles, []string{"rockchip,rk3229"}); !errors.Is(err, ErrAmbiguousBoard) {
		t.Fatalf("error=%v", err)
	}
}

func TestLoadProfilesAndRejectStorageNotInProfile(t *testing.T) {
	profiles, err := Load(
		strings.NewReader(
			`[{"version":1,"name":"reference","compatible":["rockchip,rk3229"],"storage_kinds":["emmc"],"allow_bootloader_modification":false}]`,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Match(profiles, []string{"rockchip,rk3229"})
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(p.ValidateStorage(device.KindUSB), ErrUnsupportedStorage) {
		t.Fatal("unsupported storage accepted")
	}
	if p.MayModifyBootloader() {
		t.Fatal("unverified bootloader modification allowed")
	}
}

func TestReferenceProfileLoadsFromRepositoryConfig(t *testing.T) {
	f, err := os.Open("../../../configs/boards/rk3229-reference.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	profiles, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Match(profiles, []string{"rockchip,rk3229"})
	if err != nil {
		t.Fatal(err)
	}
	if p.AllowBootloaderModification {
		t.Fatal("profile enables unverified bootloader modification")
	}
}
