package main

import (
	"reflect"
	"regexp"
	"testing"
)

func TestSpitTagsNPath(t *testing.T) {
	jctx := &JCtx{
		influxCtx: InfluxCtx{
			reXpath: regexp.MustCompile(MatchExpressionXpath),
			reKey:   regexp.MustCompile(MatchExpressionKey),
		},
	}

	tests := []struct {
		name    string
		input   string
		xmlpath string
		tags    map[string]string
	}{
		{
			"path-without-tags",
			"/path/without/tags",
			"/path/without/tags",
			map[string]string{},
		},
		{
			"ifd-admin-status",
			"/interfaces/interface[name='ge-0/0/0']/state/admin-status/",
			"/interfaces/interface/state/admin-status/",
			map[string]string{
				"/interfaces/interface/@name": "ge-0/0/0",
			},
		},
		{
			"ifl-admin-status",
			"/interfaces/interface[name='ge-0/0/0']/subinterfaces/subinterface[index='9']/state/admin-status",
			"/interfaces/interface/subinterfaces/subinterface/state/admin-status",
			map[string]string{
				"/interfaces/interface/@name":                             "ge-0/0/0",
				"/interfaces/interface/subinterfaces/subinterface/@index": "9",
			},
		},
		{
			"cmerror",
			"/junos/chassis/cmerror/counters[name='/fpc/1/pfe/0/cm/0/CM0/0/CM_CMERROR_FABRIC_REMOTE_PFE_RATE']/error",
			"/junos/chassis/cmerror/counters/error",
			map[string]string{
				"/junos/chassis/cmerror/counters/@name": "/fpc/1/pfe/0/cm/0/CM0/0/CM_CMERROR_FABRIC_REMOTE_PFE_RATE",
			},
		},
		{
			"events",
			"/junos/events/event[id='SYSTEM' and type='3' and facility='5']/attributes[key='message']/",
			"/junos/events/event/attributes/",
			map[string]string{
				"/junos/events/event/@id":             "SYSTEM",
				"/junos/events/event/@type":           "3",
				"/junos/events/event/@facility":       "5",
				"/junos/events/event/attributes/@key": "message",
			},
		},
		{
			"events-2",
			"/junos/rpm/history-results/history-test-results/history-single-test-results[owner='orlando' and test-name='orlando']/",
			"/junos/rpm/history-results/history-test-results/history-single-test-results/",
			map[string]string{
				"/junos/rpm/history-results/history-test-results/history-single-test-results/@owner":     "orlando",
				"/junos/rpm/history-results/history-test-results/history-single-test-results/@test-name": "orlando",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotxmlpath, gottags := spitTagsNPath(jctx, test.input)
			if gotxmlpath != test.xmlpath {
				t.Errorf("splitTagsNPath xmlpath failed, got: %s, want: %s", gotxmlpath, test.xmlpath)
			}
			if !reflect.DeepEqual(gottags, test.tags) {
				t.Errorf("splitTagsNPath tags failed, got: %v, want: %v", gottags, test.tags)
			}
		})
	}
}

func TestSubscriptionPathFromPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{
			"empty",
			"",
			"",
		},
		{
			"lacpd",
			"sensor_1008:/lacp/:/lacp/:lacpd",
			"/lacp/",
		},
		{
			"l2cpd",
			"sensor_1009:/lldp/:/lldp/:l2cpd",
			"/lldp/",
		},
		{
			"xmlproxyd",
			"sensor_1000_5_1:/interfaces/:/interfaces/:xmlproxyd",
			"/interfaces/",
		},
		{
			"pfe-ifd",
			"sensor_1000_1_1:/junos/system/linecard/interface/:/interfaces/:PFE",
			"/interfaces/",
		},
		{
			"arp-mib2d",
			"sensor_1002:/arp-information/ipv4/:/arp-information/ipv4/:mib2d",
			"/arp-information/ipv4/",
		},
		{
			"pfe-ifl",
			"sensor_1000_1_2:/junos/system/linecard/interface/logical/usage/:/interfaces/:PFE",
			"/interfaces/",
		},
		{
			"pfe-firewall",
			"sensor_1004:/junos/system/linecard/firewall/:/junos/system/linecard/firewall/:PFE",
			"/junos/system/linecard/firewall/",
		},
		{
			"pfe-npu-memory",
			"sensor_1018:/junos/system/linecard/npu/memory/:/junos/system/linecard/npu/memory/:PFE",
			"/junos/system/linecard/npu/memory/",
		},
		{
			"mib2d-ifindex",
			"sensor_1006:/interfaces/interface/state/ifindex/:/interfaces/interface/state/ifindex/:mib2d",
			"/interfaces/interface/state/ifindex/",
		},
		{
			"rpd",
			"sensor_1010:/network-instances/network-instance/mpls/:/network-instances/network-instance/mpls/:rpd",
			"/network-instances/network-instance/mpls/",
		},
	}

	for _, test := range tests {
		got := SubscriptionPathFromPath(test.input)
		if got != test.exp {
			t.Errorf("SubscriptionPathFromPath failed, got: %s, want: %s", got, test.exp)
		}
	}

}
