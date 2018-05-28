package main

import (
	"reflect"
	"regexp"
	"testing"
)

func TestSpitTagsNPath(t *testing.T) {
	jctx := &JCtx{
		influxCtx: InfluxCtx{
			re: regexp.MustCompile(MatchExpression),
		},
	}

	data := []struct {
		input   string
		xmlpath string
		tags    map[string]string
	}{
		{
			"/path/without/tags",
			"/path/without/tags",
			map[string]string{},
		},
		{
			"/interfaces/interface[name='ge-0/0/0']/state/admin-status/",
			"/interfaces/interface/state/admin-status/",
			map[string]string{
				"/interfaces/interface/@name": "ge-0/0/0",
			},
		},
		{
			"/interfaces/interface[name='ge-0/0/0']/subinterfaces/subinterface[index='9']/state/admin-status",
			"/interfaces/interface/subinterfaces/subinterface/state/admin-status",
			map[string]string{
				"/interfaces/interface/@name":                             "ge-0/0/0",
				"/interfaces/interface/subinterfaces/subinterface/@index": "9",
			},
		},
		{
			"/junos/chassis/cmerror/counters[name='/fpc/1/pfe/0/cm/0/CM0/0/CM_CMERROR_FABRIC_REMOTE_PFE_RATE']/error",
			"/junos/chassis/cmerror/counters/error",
			map[string]string{
				"/junos/chassis/cmerror/counters/@name": "/fpc/1/pfe/0/cm/0/CM0/0/CM_CMERROR_FABRIC_REMOTE_PFE_RATE",
			},
		},
	}

	for _, d := range data {
		gotxmlpath, gottags := spitTagsNPath(jctx, d.input)
		if gotxmlpath != d.xmlpath {
			t.Errorf("splitTagsNPath xmlpath failed, got: %s, want: %s", gotxmlpath, d.xmlpath)
		}
		if !reflect.DeepEqual(gottags, d.tags) {
			t.Errorf("splitTagsNPath tags failed, got: %v, want: %v", gottags, d.tags)
		}
	}
}

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
