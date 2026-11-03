package diagnostics

import "testing"

func TestNetworkReportWithoutAddresses(t *testing.T) {
	r := BuildNetworkReport([]NetworkInterface{{Name: "eth0"}})
	if len(r.Interfaces) != 1 || r.Interfaces[0].HasAddress {
		t.Fatalf("report=%+v", r)
	}
	if len(r.Hypotheses) == 0 {
		t.Fatal("network hypothesis missing")
	}
}
