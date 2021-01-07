package main

import (
	"fmt"
	"math"
	"math/big"
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

func TestCheckAndCeilFloatValues(t *testing.T) {

	inf64PracticalMaxLowerBound := math.MaxFloat64 - 9.979e291
	inf64PracticalMaxUpperBound := math.MaxFloat64 + 9.979e291
	val64PracticalNonMax := math.MaxFloat64 - 9.98e291
	val32Max := float32(math.MaxFloat32)
	negInf64PracticalMaxLowerBound := -(math.MaxFloat64 - 9.979e291)
	negInf64PracticalMaxUpperBound := -(math.MaxFloat64 + 9.979e291)
	val64PracticalNonMin := -(math.MaxFloat64 - 9.98e291)
	val32Min := float32(-math.MaxFloat32)
	zero32 := float32(0)
	zero64 := float64(0)
	tests := []struct {
		name     string
		val32    *float32
		val64    *float64
		expected float64
	}{
		{
			name:     "PracticalMaxLowerBoun",
			val32:    nil,
			val64:    &inf64PracticalMaxLowerBound,
			expected: math.MaxFloat64,
		},
		{
			name:     "PracticalMaxUpperBound",
			val32:    nil,
			val64:    &inf64PracticalMaxUpperBound,
			expected: math.MaxFloat64,
		},
		{
			name:     "PracticalNonMax",
			val32:    nil,
			val64:    &val64PracticalNonMax,
			expected: val64PracticalNonMax,
		},
		{
			name:     "Float32",
			val32:    &val32Max,
			val64:    nil,
			expected: float64(val32Max),
		},
		{
			name:     "negInf64PracticalMaxLowerBound",
			val32:    nil,
			val64:    &negInf64PracticalMaxLowerBound,
			expected: -math.MaxFloat64,
		},
		{
			name:     "negInf64PracticalMaxUpperBound",
			val32:    nil,
			val64:    &negInf64PracticalMaxUpperBound,
			expected: -math.MaxFloat64,
		},
		{
			name:     "PracticalNonMin",
			val32:    nil,
			val64:    &val64PracticalNonMin,
			expected: val64PracticalNonMin,
		},
		{
			name:     "Float32Min",
			val32:    &val32Min,
			val64:    nil,
			expected: float64(val32Min),
		},
		{
			name:     "Zero32",
			val32:    &zero32,
			val64:    nil,
			expected: float64(zero32),
		},
		{
			name:     "Zero64",
			val32:    nil,
			val64:    &zero64,
			expected: float64(zero64),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output float64
			checkAndCeilFloatValues(test.val32, test.val64, &output)
			diff := uint64(big.NewFloat(output).Cmp(big.NewFloat(test.expected)))
			if diff != 0 {
				var errMsg string
				errMsg = fmt.Sprintf("\nNot equal, expected :%v\nGot :%v, diff: %v", test.expected, output, diff)
				t.Errorf(errMsg)
			}
		})
	}
}
