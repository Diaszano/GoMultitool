package discovery

import (
	"context"
	"reflect"
	"testing"

	"github.com/Diaszano/GoMultitool/internal/domain/device"
)

type catalogStub []device.Device

func (s catalogStub) List(context.Context) ([]device.Device, error) { return s, nil }
func (s catalogStub) Resolve(context.Context, string) (device.Device, error) {
	return device.Device{}, nil
}

func TestServiceInventoryIsDeterministic(t *testing.T) {
	input := catalogStub{{ID: "sdb"}, {ID: "mmc0"}}
	got, err := (Service{Catalog: input}).Inventory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string{got[0].ID, got[1].ID}, []string{"mmc0", "sdb"}) {
		t.Fatalf("inventory order = %v", got)
	}
}
