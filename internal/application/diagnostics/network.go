package diagnostics

import "net"

// NetworkInterface records an interface name and assigned IP addresses.
type NetworkInterface struct {
	Name       string   `json:"name"`
	Addresses  []string `json:"addresses,omitempty"`
	HasAddress bool     `json:"has_address"`
}

// NetworkReport separates interface observations from likely causes.
type NetworkReport struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	Hypotheses []string           `json:"hypotheses,omitempty"`
}

// BuildNetworkReport adds hypotheses for interfaces without addresses.
func BuildNetworkReport(interfaces []NetworkInterface) NetworkReport {
	r := NetworkReport{Interfaces: interfaces}
	for i := range r.Interfaces {
		r.Interfaces[i].HasAddress = len(r.Interfaces[i].Addresses) > 0
	}
	if len(r.Interfaces) == 0 {
		r.Hypotheses = append(r.Hypotheses, "no network interfaces were reported")
	} else {
		for _, iface := range r.Interfaces {
			if iface.Name != "lo" && !iface.HasAddress {
				r.Hypotheses = append(r.Hypotheses, "interface "+iface.Name+" has no assigned IP address")
			}
		}
	}
	return r
}

// CollectNetworkReport reads the host's interface state without changing it.
func CollectNetworkReport() NetworkReport {
	interfaces, err := net.Interfaces()
	if err != nil {
		return NetworkReport{Hypotheses: []string{"network interface list could not be read"}}
	}
	var entries []NetworkInterface
	for _, iface := range interfaces {
		entry := NetworkInterface{Name: iface.Name}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			entry.Addresses = append(entry.Addresses, addr.String())
		}
		entries = append(entries, entry)
	}
	return BuildNetworkReport(entries)
}
