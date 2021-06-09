package main

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"

	google_protobuf "github.com/golang/protobuf/ptypes/any"
	gnmi "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	gnmi_ext1 "github.com/nileshsimaria/jtimon/gnmi/gnmi_ext"
	gnmi_juniper_header "github.com/nileshsimaria/jtimon/gnmi/gnmi_juniper_header"
	gnmi_juniper_header_ext "github.com/nileshsimaria/jtimon/gnmi/gnmi_juniper_header_ext"
)

func TestXPathTognmiPath(t *testing.T) {
	tests := []struct {
		name  string
		xpath string
		err   bool
		path  *gnmi.Path
	}{
		{
			name:  "multi-level-multi-keys",
			xpath: "/interfaces/interface[k1=\"foo\"]/subinterfaces/subinterface[k1=\"foo1\" and k2=\"bar1\"]",
			err:   false,
			path: &gnmi.Path{Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
		},
		{
			name:  "multi-level-multi-keys-err",
			xpath: "/interfaces/interface[k1=\"foo\"]/subinterfaces/subinterface[k1=\"foo1\" and k2=\"bar1\"]",
			err:   true,
			path: &gnmi.Path{Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1"}},
				},
			},
		},
		{
			name:  "multi-level-multi-keys-xpath-err",
			xpath: "/interfaces/interface[k1=\"foo]/subinterfaces/subinterface[k1=\"foo1\" and k2=\"bar1\"]",
			err:   true,
			path: &gnmi.Path{Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1"}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gnmiPath, err := xPathTognmiPath(test.xpath)
			if !test.err {
				if err != nil || !reflect.DeepEqual(*test.path, *gnmiPath) {
					var errMsg string
					if err == nil {
						errMsg = fmt.Sprintf("\nexpected:%v\nGot:%v", *test.path, *gnmiPath)
					} else {
						errMsg = fmt.Sprintf("Not an error, but got an error: %v", err)
					}
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if err == nil && reflect.DeepEqual(*test.path, *gnmiPath) {
					var errMsg string
					errMsg = fmt.Sprintf("want error but got nil\n")
					errMsg += fmt.Sprintf("\nexpected:%v\nGot:%v\n", *test.path, *gnmiPath)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

func TestGnmiMode(t *testing.T) {
	tests := []struct {
		name    string
		inMode  string
		err     bool
		outMode gnmi.SubscriptionMode
	}{
		{
			name:    "gnmi-mode-on-change",
			inMode:  "on-change",
			err:     false,
			outMode: gnmi.SubscriptionMode_ON_CHANGE,
		},
		{
			name:    "gnmi-mode-sample",
			inMode:  "",
			err:     false,
			outMode: gnmi.SubscriptionMode_SAMPLE,
		},
		{
			name:    "gnmi-mode-err",
			inMode:  "target-defined",
			err:     true,
			outMode: gnmi.SubscriptionMode_SAMPLE,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outMode := gnmiMode(test.inMode)
			if !test.err {
				if test.outMode != outMode {
					var errMsg string
					errMsg = fmt.Sprintf("\nexpected:%v\nGot:%v", test.outMode, outMode)
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if test.outMode == outMode {
					var errMsg string
					errMsg = fmt.Sprintf("\nexpected:%v\nGot:%v", test.outMode, outMode)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

func TestGnmiFreq(t *testing.T) {
	tests := []struct {
		name    string
		inMode  gnmi.SubscriptionMode
		inFreq  uint64
		err     bool
		outMode gnmi.SubscriptionMode
		outFreq uint64
	}{
		{
			name:    "gnmi-freq-sample",
			inMode:  gnmi.SubscriptionMode_SAMPLE,
			inFreq:  30000,
			err:     false,
			outMode: gnmi.SubscriptionMode_SAMPLE,
			outFreq: 30000000000,
		},
		{
			name:    "gnmi-freq-on-change",
			inMode:  gnmi.SubscriptionMode_ON_CHANGE,
			inFreq:  30,
			err:     false,
			outMode: gnmi.SubscriptionMode_ON_CHANGE,
			outFreq: 0,
		},
		{
			name:    "gnmi-freq-sample-taken-as-on-change",
			inMode:  gnmi.SubscriptionMode_SAMPLE,
			inFreq:  0,
			err:     false,
			outMode: gnmi.SubscriptionMode_ON_CHANGE,
			outFreq: 0,
		},
		{
			name:    "gnmi-freq-target-defined",
			inMode:  gnmi.SubscriptionMode_TARGET_DEFINED,
			inFreq:  30,
			err:     false,
			outMode: gnmi.SubscriptionMode_TARGET_DEFINED,
			outFreq: gGnmiFreqMin, // default
		},
		{
			name:    "gnmi-freq-sample-err",
			inMode:  gnmi.SubscriptionMode_SAMPLE,
			inFreq:  30,
			err:     true,
			outMode: gnmi.SubscriptionMode_SAMPLE,
			outFreq: 30000000000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outMode, outFreq := gnmiFreq(test.inMode, test.inFreq)
			if !test.err {
				if test.outMode != outMode && test.outFreq != outFreq {
					var errMsg string
					errMsg = fmt.Sprintf("\nexpected:(%v, %v)\nGot:(%v, %v)", test.outMode, test.outFreq, outMode, outFreq)
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if test.outMode == outMode && test.outFreq == outFreq {
					var errMsg string
					errMsg = fmt.Sprintf("\nexpected:(%v, %v)\nGot:(%v, %v)", test.outMode, test.outFreq, outMode, outFreq)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

func TestGnmiParseUpdates(t *testing.T) {
	tests := []struct {
		name        string
		parseOrigin bool
		prefix      *gnmi.Path
		updates     []*gnmi.Update
		parseOutput *gnmiParseOutputT
		err         bool
		output      *gnmiParseOutputT
		enableUint  bool
	}{
		{
			name:        "updates-valid-no-prefix",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "description"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_StringVal{StringVal: "Hello"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"/interfaces/interface/@k1": "foo",
					"/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"/interfaces/interface/subinterfaces/subinterface/state/description": "Hello"},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-with-prefix",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "description"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_StringVal{StringVal: "Hello"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs:    map[string]string{"/a/b/@k1": "foo", "/a/b/c/d/@k1": "foo1", "/a/b/c/d/@k2": "bar1"},
				xpaths:     map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs: map[string]string{"/a/b/@k1": "foo",
					"/a/b/c/d/@k1": "foo1",
					"/a/b/c/d/@k2": "bar1"},
				xpaths: map[string]interface{}{"/a/b/c/d/state/description": "Hello"},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-with-misc-simple-types",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "mtu"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_IntVal{IntVal: 1500},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "in-octets"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_UintVal{UintVal: 40000},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-dec64"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_DecimalVal{DecimalVal: &gnmi.Decimal64{Digits: 9007199254740992, Precision: 15}},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-float"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_FloatVal{FloatVal: 32.45},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "power-inf-float"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_FloatVal{FloatVal: math.MaxFloat32 + 1},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs:    map[string]string{"/a/b/@k1": "foo", "/a/b/c/d/@k1": "foo1", "/a/b/c/d/@k2": "bar1"},
				xpaths:     map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs: map[string]string{"/a/b/@k1": "foo",
					"/a/b/c/d/@k1": "foo1",
					"/a/b/c/d/@k2": "bar1"},
				xpaths: map[string]interface{}{
					"/a/b/c/d/state/mtu":                       int64(1500),
					"/a/b/c/d/state/counters/in-octets":        float64(40000),
					"/a/b/c/d/state/counters/out-octets-dec64": float64(9.007199254740992),
					"/a/b/c/d/state/counters/out-octets-float": float64(float32(32.45)), // Guess this may not always work..
					"/a/b/c/d/state/counters/power-inf-float":  float64(3.40282346638528859811704183484516925440e+38),
				},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-scalar-array",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "different-mtus"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_LeaflistVal{
							LeaflistVal: &gnmi.ScalarArray{
								Element: []*gnmi.TypedValue{
									{
										Value: &gnmi.TypedValue_IntVal{IntVal: 1500},
									},
									{
										Value: &gnmi.TypedValue_IntVal{IntVal: 1499},
									},
									{
										Value: &gnmi.TypedValue_IntVal{IntVal: 1501},
									},
								},
							},
						},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"/interfaces/interface/@k1": "foo",
					"/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"/interfaces/interface/subinterfaces/subinterface/state/different-mtus": []int64{1500, 1499, 1501}},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-with-misc-simple-types-json",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "mtu"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`1500`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "in-octets"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`"40000000000000"`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-dec64"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`9.007199254740992`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "enabled"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`true`)},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs:    map[string]string{"/a/b/@k1": "foo", "/a/b/c/d/@k1": "foo1", "/a/b/c/d/@k2": "bar1"},
				xpaths:     map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs: map[string]string{"/a/b/@k1": "foo",
					"/a/b/c/d/@k1": "foo1",
					"/a/b/c/d/@k2": "bar1"},
				xpaths: map[string]interface{}{
					"/a/b/c/d/state/mtu":                       int64(1500),
					"/a/b/c/d/state/counters/in-octets":        "40000000000000",
					"/a/b/c/d/state/counters/out-octets-dec64": 9.007199254740992,
					"/a/b/c/d/state/enabled":                   true,
				},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-with-misc-simple-types-json_ietf",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "mtu"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`1500`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "in-octets"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`"40000000000000"`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-dec64"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`"9.007199254740992"`)},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "enabled"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte(`true`)},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs:    map[string]string{"/a/b/@k1": "foo", "/a/b/c/d/@k1": "foo1", "/a/b/c/d/@k2": "bar1"},
				xpaths:     map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs: map[string]string{"/a/b/@k1": "foo",
					"/a/b/c/d/@k1": "foo1",
					"/a/b/c/d/@k2": "bar1"},
				xpaths: map[string]interface{}{
					"/a/b/c/d/state/mtu":                       int64(1500),
					"/a/b/c/d/state/counters/in-octets":        "40000000000000",
					"/a/b/c/d/state/counters/out-octets-dec64": "9.007199254740992",
					"/a/b/c/d/state/enabled":                   true,
				},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-no-prefix-with-origin--config--donotParseOrigin",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "openconfig",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "description"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_StringVal{StringVal: "Hello"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"/interfaces/interface/@k1": "foo",
					"/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"/interfaces/interface/subinterfaces/subinterface/state/description": "Hello"},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-no-prefix-with-origin--config--parseOrigin",
			err:         false,
			parseOrigin: true,
			prefix: &gnmi.Path{
				Origin: "openconfig",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "description"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_StringVal{StringVal: "Hello"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "openconfig:/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"openconfig:/interfaces/interface/@k1": "foo",
					"openconfig:/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"openconfig:/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"openconfig:/interfaces/interface/subinterfaces/subinterface/state/description": "Hello"},
			},
			enableUint: false,
		},
		{
			name:        "updates-valid-with-misc-simple-types-uint-enabled",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			updates: []*gnmi.Update{
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "mtu"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_IntVal{IntVal: 1500},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "in-octets"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_UintVal{UintVal: 40000},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-dec64"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_DecimalVal{DecimalVal: &gnmi.Decimal64{Digits: 9007199254740992, Precision: 15}},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "out-octets-float"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_FloatVal{FloatVal: 32.45},
					},
				},
				{
					Path: &gnmi.Path{
						Origin: "",
						Elem: []*gnmi.PathElem{
							{Name: "state"},
							{Name: "counters"},
							{Name: "power-inf-float"},
						},
					},
					Val: &gnmi.TypedValue{
						Value: &gnmi.TypedValue_FloatVal{FloatVal: math.MaxFloat32 + 1},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs:    map[string]string{"/a/b/@k1": "foo", "/a/b/c/d/@k1": "foo1", "/a/b/c/d/@k2": "bar1"},
				xpaths:     map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/a/b/c/d",
				kvpairs: map[string]string{"/a/b/@k1": "foo",
					"/a/b/c/d/@k1": "foo1",
					"/a/b/c/d/@k2": "bar1"},
				xpaths: map[string]interface{}{
					"/a/b/c/d/state/mtu":                       int64(1500),
					"/a/b/c/d/state/counters/in-octets":        uint64(40000),
					"/a/b/c/d/state/counters/out-octets-dec64": float64(9.007199254740992),
					"/a/b/c/d/state/counters/out-octets-float": float64(float32(32.45)), // Guess this may not always work..
					"/a/b/c/d/state/counters/power-inf-float":  float64(3.40282346638528859811704183484516925440e+38),
				},
			},
			enableUint: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parseOutput, err := gnmiParseUpdates(test.parseOrigin, test.prefix, test.updates, test.parseOutput, test.enableUint)
			if !test.err {
				if err != nil || !reflect.DeepEqual(*test.output, *parseOutput) {
					var errMsg string
					if err == nil {
						errMsg = fmt.Sprintf("\nexpected :%v\nGot:%v", *test.output, *parseOutput)
					} else {
						errMsg = fmt.Sprintf("Not an error, but got an error: %v", err)
					}
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if err == nil && reflect.DeepEqual(*test.output, *parseOutput) {
					var errMsg string
					errMsg = fmt.Sprintf("want error but got nil\n")
					errMsg += fmt.Sprintf("\nexpected:%v\nGot:%v\n", *test.output, *parseOutput)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

func TestGnmiParseDeletes(t *testing.T) {
	tests := []struct {
		name        string
		parseOrigin bool
		prefix      *gnmi.Path
		deletes     []*gnmi.Path
		parseOutput *gnmiParseOutputT
		err         bool
		output      *gnmiParseOutputT
	}{
		{
			name:        "deletes-valid-no-prefix",
			err:         false,
			parseOrigin: false,
			prefix: &gnmi.Path{
				Origin: "",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			deletes: []*gnmi.Path{
				{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{Name: "state"},
						{Name: "description"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"/interfaces/interface/@k1": "foo",
					"/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"/interfaces/interface/subinterfaces/subinterface/state/description": nil},
			},
		},
		{
			name:        "deletes-valid-no-prefix-with-origin",
			err:         false,
			parseOrigin: true,
			prefix: &gnmi.Path{
				Origin: "dummy",
				Elem: []*gnmi.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"k1": "foo"}},
					{Name: "subinterfaces"},
					{Name: "subinterface", Key: map[string]string{"k1": "foo1", "k2": "bar1"}},
				},
			},
			deletes: []*gnmi.Path{
				{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{Name: "state"},
						{Name: "description"},
					},
				},
			},
			parseOutput: &gnmiParseOutputT{
				kvpairs: map[string]string{},
				xpaths:  map[string]interface{}{},
			},
			output: &gnmiParseOutputT{
				prefixPath: "dummy:/interfaces/interface/subinterfaces/subinterface",
				kvpairs: map[string]string{"dummy:/interfaces/interface/@k1": "foo",
					"dummy:/interfaces/interface/subinterfaces/subinterface/@k1": "foo1",
					"dummy:/interfaces/interface/subinterfaces/subinterface/@k2": "bar1"},
				xpaths: map[string]interface{}{"dummy:/interfaces/interface/subinterfaces/subinterface/state/description": nil},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parseOutput, err := gnmiParseDeletes(test.parseOrigin, test.prefix, test.deletes, test.parseOutput)
			if !test.err {
				if err != nil || !reflect.DeepEqual(*test.output, *parseOutput) {
					var errMsg string
					if err == nil {
						errMsg = fmt.Sprintf("\nexpected:%v\nGot:%v", *test.output, *parseOutput)
					} else {
						errMsg = fmt.Sprintf("Not an error, but got an error: %v", err)
					}
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if err == nil && reflect.DeepEqual(*test.output, *parseOutput) {
					var errMsg string
					errMsg = fmt.Sprintf("want error but got nil\n")
					errMsg += fmt.Sprintf("\nexpected:%v\nGot:%v\n", *test.output, *parseOutput)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

func TestFormJuniperTelemetryHdr(t *testing.T) {
	var hdrInputXpath = gnmi_juniper_header.GnmiJuniperTelemetryHeader{
		SystemId: "my-device", ComponentId: 65535, SubComponentId: 0,
		Path: "sensor_1:/a/:/z/:my-app", SequenceNumber: 100,
	}

	hdrInputXpathBytes, err := proto.Marshal(&hdrInputXpath)
	if err != nil {
		t.Errorf("Error marshalling header for xpath case: %v", err)
	}

	var hdrInputExt = gnmi_juniper_header_ext.GnmiJuniperTelemetryHeaderExtension{
		SystemId: "my-device", ComponentId: 65535, SubComponentId: 0,
		SensorName: "sensor_1:/a/:/z/:my-app", StreamedPath: "/a/", SubscribedPath: "/z/",
		Component: "my-app", SequenceNumber: 100,
	}

	hdrInputExtBytes, err := proto.Marshal(&hdrInputExt)
	if err != nil {
		t.Errorf("Error marshalling header for xpath case: %v", err)
	}

	tests := []struct {
		name      string
		jXpaths   *jnprXpathDetails
		exts      []*gnmi_ext1.Extension
		err       bool
		isJuniper bool
		output    *juniperGnmiHeaderDetails
	}{
		{
			name: "juniper-gnmi-header-in-updates-parsed-as-xpaths",
			err:  false,
			jXpaths: &jnprXpathDetails{
				xPaths: map[string]interface{}{"/a/b/c/d/__juniper_telemetry_header__": &google_protobuf.Any{
					TypeUrl: "type.googleapis.com/GnmiJuniperTelemetryHeader",
					Value:   hdrInputXpathBytes,
				}},
				hdrXpath:       "/a/b/c/d/__juniper_telemetry_header__",
				publishTsXpath: "",
			},
			isJuniper: true,
			output: &juniperGnmiHeaderDetails{
				hdr: &hdrInputXpath,
			},
		},
		{
			name: "juniper-gnmi-header-in-extension",
			err:  false,
			exts: []*gnmi_ext1.Extension{
				{
					Ext: &gnmi_ext1.Extension_RegisteredExt{
						RegisteredExt: &gnmi_ext1.RegisteredExtension{
							Id:  gnmi_ext1.ExtensionID_EID_JUNIPER_TELEMETRY_HEADER,
							Msg: hdrInputExtBytes,
						},
					},
				},
			},
			isJuniper: true,
			output: &juniperGnmiHeaderDetails{
				hdrExt: &hdrInputExt,
			},
		},
		{
			name: "other-vendor",
			err:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputHdr, isJnpr, err := formJuniperTelemetryHdr(test.jXpaths, test.exts)
			if !test.err {
				if err != nil || test.isJuniper != isJnpr || !reflect.DeepEqual(*test.output, *outputHdr) {
					var errMsg string
					if err == nil {
						if test.isJuniper == isJnpr {
							errMsg = fmt.Sprintf("\nexpected:%v\nGot:%v\n", *test.output, *outputHdr)
						} else {
							errMsg = fmt.Sprintf("expected jnpr header?%v, Is it present?%v\n", test.isJuniper, isJnpr)
						}
					} else {
						errMsg = fmt.Sprintf("Not an error, but got an error: %v\n", err)
					}
					t.Errorf(errMsg)
				}
			}

			if test.err {
				if err == nil && test.isJuniper == isJnpr && reflect.DeepEqual(*test.output, *outputHdr) {
					var errMsg string
					errMsg = fmt.Sprintf("want error but got nil\n")
					errMsg += fmt.Sprintf("expected jnpr header?%v, Is it present?%v\n", test.isJuniper, isJnpr)
					errMsg += fmt.Sprintf("\nexpected:%v\nGot:%v", *test.output, *outputHdr)
					t.Errorf(errMsg)
				}
			}
		})
	}
}

// The below functions should have been already covered by now, so no need to UT them
func TestGnmiParsePath(t *testing.T) {
}

func TestGnmiParseValue(t *testing.T) {
}
