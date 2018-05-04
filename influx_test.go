package main

import (
	"testing"
)

func TestSubscriptionPathFromPath(t *testing.T) {
	paths := []struct {
		input string
		exp   string
	}{
		{"", ""},
		{"sensor_1008:/lacp/:/lacp/:lacpd", "/lacp/"},
		{"sensor_1009:/lldp/:/lldp/:l2cpd", "/lldp/"},
		{"sensor_1000_5_1:/interfaces/:/interfaces/:xmlproxyd", "/interfaces/"},
		{"sensor_1000_1_1:/junos/system/linecard/interface/:/interfaces/:PFE", "/interfaces/"},
		{"sensor_1002:/arp-information/ipv4/:/arp-information/ipv4/:mib2d", "/arp-information/ipv4/"},
		{"sensor_1000_1_2:/junos/system/linecard/interface/logical/usage/:/interfaces/:PFE", "/interfaces/"},
		{"sensor_1004:/junos/system/linecard/firewall/:/junos/system/linecard/firewall/:PFE", "/junos/system/linecard/firewall/"},
		{"sensor_1018:/junos/system/linecard/npu/memory/:/junos/system/linecard/npu/memory/:PFE", "/junos/system/linecard/npu/memory/"},
		{"sensor_1006:/interfaces/interface/state/ifindex/:/interfaces/interface/state/ifindex/:mib2d", "/interfaces/interface/state/ifindex/"},
		{"sensor_1010:/network-instances/network-instance/mpls/:/network-instances/network-instance/mpls/:rpd", "/network-instances/network-instance/mpls/"},
	}

	for _, path := range paths {
		got := SubscriptionPathFromPath(path.input)
		if got != path.exp {
			t.Errorf("SubscriptionPathFromPath failed, got: %s, want: %s", got, path.exp)
		}
	}

}
