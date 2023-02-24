package jtisim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	gnmi "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	gnmipb "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	gnmi_ext1 "github.com/nileshsimaria/jtimon/gnmi/gnmi_ext"
	gnmi_juniper_header_ext "github.com/nileshsimaria/jtimon/gnmi/gnmi_juniper_header_ext"
	tpb "github.com/nileshsimaria/jtimon/telemetry"
)

// IDesc Interface description structrue
type IDesc struct {
	Desc Description `json:"desc"`
	IFD  IFDCounters `json:"ifd-counters"`
	IFL  IFLCounters `json:"ifl-counters"`
}

// Description of interfaces
type Description struct {
	Media   string `json:"media"`
	FPC     int    `json:"fpc"`
	PIC     int    `json:"pic"`
	PORT    int    `json:"port"`
	Logical int    `json:"logical"`
}

// IFDCounters of interfaces
type IFDCounters struct {
	INPkts      int32 `json:"in-pkts"`
	INOctets    int32 `json:"in-octets"`
	AdminStatus bool  `json:"admin-status"`
	OperStatus  bool  `json:"oper-status"`
}

// IFLCounters of interfaces
type IFLCounters struct {
	INUnicastPkts   int32 `json:"in-unicast-pkts"`
	INMulticastPkts int32 `json:"in-multicast-pkts"`
}

func parseInterfacesJSON(dir string) *IDesc {
	file, err := ioutil.ReadFile(dir + "/interfaces.json")
	if err != nil {
		log.Fatalf("%v", err)
		os.Exit(1)
	}

	var iDesc IDesc
	if err := json.Unmarshal(file, &iDesc); err != nil {
		panic(err)
	}
	return &iDesc
}

type interfaces struct {
	desc *IDesc
	ifds []*ifd
}
type ifd struct {
	name        string
	inPkts      uint64
	inOctets    uint64
	adminStatus string
	operStatus  string
	ifls        []*ifl
}

type ifl struct {
	index   int
	inUPkts uint64
	inMPkts uint64
}

func generateIList(idesc *IDesc) *interfaces {
	fpc := idesc.Desc.FPC
	pic := idesc.Desc.PIC
	port := idesc.Desc.PORT
	media := idesc.Desc.Media
	logical := idesc.Desc.Logical

	interfaces := &interfaces{
		desc: idesc,
		ifds: make([]*ifd, fpc*pic*port),
	}

	cnt := 0
	for i := 0; i < fpc; i++ {
		for j := 0; j < pic; j++ {
			for k := 0; k < port; k++ {
				name := fmt.Sprintf("%s-%d/%d/%d", media, i, j, k)
				ifd := &ifd{
					name: name,
				}
				ifd.ifls = make([]*ifl, logical)

				for index := 0; index < logical; index++ {
					ifl := ifl{
						index: index,
					}
					ifd.ifls[index] = &ifl
				}

				interfaces.ifds[cnt] = ifd
				cnt++

			}
		}
	}
	return interfaces
}

func getRandom(num int32, random bool) int32 {
	if random == false {
		return 100
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int31n(num)
}

func (s *server) streamInterfaces(ch chan *tpb.OpenConfigData, path *tpb.Path) {
	sysID := fmt.Sprintf("jtisim:%s:%d", s.jtisim.host, s.jtisim.port)
	pname := path.GetPath()
	freq := path.GetSampleFrequency()
	log.Println(pname, freq)

	nsFreq := time.Duration(freq) * 1000000
	iDesc := parseInterfacesJSON(s.jtisim.descDir)
	interfaces := generateIList(iDesc)

	seq := uint64(0)

	for {
		ifds := interfaces.ifds
		start := time.Now()
		for _, ifd := range ifds {
			prefixV := fmt.Sprintf("/interfaces/interface[name='%s']/", ifd.name)

			rValue := getRandom(interfaces.desc.IFD.INPkts, s.jtisim.random)
			inp := ifd.inPkts + uint64((uint32(rValue) * (freq / 1000)))
			ifd.inPkts = inp

			rValue = getRandom(interfaces.desc.IFD.INOctets, s.jtisim.random)
			ino := ifd.inOctets + uint64((uint32(rValue) * (freq / 1000)))
			ifd.inOctets = ino

			ops := "UP"
			ads := "DOWN"

			kv := []*tpb.KeyValue{
				{Key: "__prefix__", Value: &tpb.KeyValue_StrValue{StrValue: prefixV}},
				{Key: "name", Value: &tpb.KeyValue_StrValue{StrValue: ifd.name}},
				{Key: "state/oper-status", Value: &tpb.KeyValue_StrValue{StrValue: ops}},
				{Key: "state/admin-status", Value: &tpb.KeyValue_StrValue{StrValue: ads}},
				{Key: "state/counters/in-pkts", Value: &tpb.KeyValue_UintValue{UintValue: inp}},
				{Key: "state/counters/in-octets", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "init_time", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/counters/carrier-transitions", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/counters/last-clear", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/counters/out-octets", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/counters/out-pkts", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/counters/out-unicast-pkts", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/description", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/high-speed", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/ifindex", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/last-change", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/name", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/parent_ae_name", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				{Key: "state/type", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
			}

			d := &tpb.OpenConfigData{
				SystemId:       sysID,
				ComponentId:    1,
				Timestamp:      uint64(MakeMSTimestamp()),
				SequenceNumber: seq,
				Kv:             kv,
				SyncResponse:   false,
				Path:           "sensor_1000_1_1:/junos/system/linecard/interface/:/interfaces/:PFE",
			}
			seq++
			ch <- d

			for _, ifl := range ifd.ifls {
				prefixVifl := fmt.Sprintf("/interfaces/interface[name='%s']/subinterfaces/subinterface[index='%d']/", ifd.name, ifl.index)

				rValue := getRandom(interfaces.desc.IFL.INUnicastPkts, s.jtisim.random)
				inup := ifl.inUPkts + uint64((uint32(rValue) * (freq / 1000)))
				ifl.inUPkts = inup

				rValue = getRandom(interfaces.desc.IFL.INMulticastPkts, s.jtisim.random)
				inmp := ifl.inMPkts + uint64((uint32(rValue) * (freq / 1000)))
				ifl.inMPkts = inmp
				name := fmt.Sprintf("%s.%d", ifd.name, ifl.index)

				kvifl := []*tpb.KeyValue{
					{Key: "__prefix__", Value: &tpb.KeyValue_StrValue{StrValue: prefixVifl}},
					{Key: "index", Value: &tpb.KeyValue_UintValue{UintValue: uint64(ifl.index)}},
					{Key: "state/name", Value: &tpb.KeyValue_StrValue{StrValue: name}},
					{Key: "state/counters/in-unicast-pkts", Value: &tpb.KeyValue_UintValue{UintValue: inup}},
					{Key: "state/counters/in-multicast-pkts", Value: &tpb.KeyValue_UintValue{UintValue: inmp}},
					{Key: "state/oper-status", Value: &tpb.KeyValue_StrValue{StrValue: ops}},
					{Key: "state/admin-status", Value: &tpb.KeyValue_StrValue{StrValue: ads}},
					{Key: "ipv4/addresses/address/ipv4/state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/ipv4/unnumbered/interface-ref/state/interface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/ipv4/unnumbered/interface-ref/state/subinterface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/ipv4/unnumbered/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/state/ip", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/state/origin", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/addresses/address/state/prefix-length", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/@ip", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/ipv4/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/ipv4/state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/ipv4/unnumbered/interface-ref/state/interface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/ipv4/unnumbered/interface-ref/state/subinterface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/ipv4/unnumbered/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/expiry", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/host-name", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/interface-name", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/ip", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/is-publish", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/link-layer-address", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/logical-router-id", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/neighbor-state", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/origin", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/neighbors/neighbor/state/table-id", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/unnumbered/interface-ref/state/interface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/unnumbered/interface-ref/state/subinterface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv4/unnumbered/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/@ip", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/ipv6/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/ipv6/state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/ipv6/unnumbered/interface-ref/state/interface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/ipv6/unnumbered/interface-ref/state/subinterface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/ipv6/unnumbered/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/state/ip", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/state/origin", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/state/prefix-length", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/addresses/address/state/status", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/state/mtu", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/unnumbered/interface-ref/state/interface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/unnumbered/interface-ref/state/subinterface", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "ipv6/unnumbered/state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/counters/in-octets", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/counters/in-pkts", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/counters/out-octets", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/counters/out-pkts", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/description", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/enabled", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/ifindex", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/index", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
					{Key: "state/last-change", Value: &tpb.KeyValue_UintValue{UintValue: ino}},
				}

				d := &tpb.OpenConfigData{
					SystemId:       sysID,
					ComponentId:    1,
					Timestamp:      uint64(MakeMSTimestamp()),
					SequenceNumber: seq,
					Kv:             kvifl,
					SyncResponse:   false,
					Path:           "sensor_1013_1_1:/junos/system/linecard/interface/logical/usage/:/interfaces/:PFE",
				}
				seq++
				ch <- d

			}

		} //finish one wrap
		wrapDuration := time.Since(start)
		time.Sleep(nsFreq - wrapDuration)
	}
}

// Not the way JUNOS publishes the packets, but had to make sure to sync with OC way of simulation..
func (s *server) gnmiStreamInterfaces(ch chan *gnmipb.SubscribeResponse, pname string, sub *gnmipb.Subscription) {
	sysID := fmt.Sprintf("jtisim:%s:%d", s.jtisim.host, s.jtisim.port)
	freq := sub.GetSampleInterval()
	log.Println(pname, freq)

	nsFreq := time.Duration(freq)
	iDesc := parseInterfacesJSON(s.jtisim.descDir)
	interfaces := generateIList(iDesc)

	seq := uint64(0)

	for {
		ifds := interfaces.ifds
		start := time.Now()
		for _, ifd := range ifds {
			rValue := getRandom(interfaces.desc.IFD.INPkts, s.jtisim.random)
			inp := ifd.inPkts + (uint64(rValue) * (freq * 1000))
			ifd.inPkts = inp

			rValue = getRandom(interfaces.desc.IFD.INOctets, s.jtisim.random)
			ino := ifd.inOctets + (uint64(rValue) * (freq * 1000))
			ifd.inOctets = ino

			ops := "UP"
			ads := "DOWN"

			var hdrInputExt = gnmi_juniper_header_ext.GnmiJuniperTelemetryHeaderExtension{
				SystemId: sysID, ComponentId: 1, SubComponentId: 0,
				SensorName: "sensor_1000_1_1", SequenceNumber: seq, SubscribedPath: "/interfaces/",
				StreamedPath: "/junos/system/linecard/interface/", Component: "PFE",
			}

			hdrInputExtBytes, err := proto.Marshal(&hdrInputExt)
			if err != nil {
				log.Fatalf("Error marshalling header for ext case: %v", err)
			}

			rsp := &gnmi.SubscribeResponse{
				Extension: []*gnmi_ext1.Extension{
					{
						Ext: &gnmi_ext1.Extension_RegisteredExt{
							RegisteredExt: &gnmi_ext1.RegisteredExtension{
								Id:  gnmi_ext1.ExtensionID_EID_JUNIPER_TELEMETRY_HEADER,
								Msg: hdrInputExtBytes,
							},
						},
					},
				},
				Response: &gnmi.SubscribeResponse_Update{
					Update: &gnmi.Notification{
						Timestamp: time.Now().UnixNano(),
						Prefix: &gnmi.Path{
							Origin: "",
							Elem: []*gnmi.PathElem{
								{Name: "interfaces"},
								{Name: "interface", Key: map[string]string{"name": ifd.name}},
							},
						},
						Update: []*gnmi.Update{
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "name"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_StringVal{StringVal: ifd.name},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "oper-status"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_StringVal{StringVal: ops},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "admin-status"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_StringVal{StringVal: ads},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "in-pkts"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: inp},
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
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "init-time"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "carrier-transitions"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "last-clear"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "out-octets"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "out-pkts"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "counters"},
										{Name: "out-unicast-pkts"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "description"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
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
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "high-speed"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "ifindex"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "last-change"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "mtu"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "name"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "parent_ae_name"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
							{
								Path: &gnmi.Path{
									Origin: "",
									Elem: []*gnmi.PathElem{
										{Name: "state"},
										{Name: "type"},
									},
								},
								Val: &gnmi.TypedValue{
									Value: &gnmi.TypedValue_UintVal{UintVal: ino},
								},
							},
						},
					},
				},
			}

			seq++
			ch <- rsp

			hdrInputExt = gnmi_juniper_header_ext.GnmiJuniperTelemetryHeaderExtension{
				SystemId: sysID, ComponentId: 1, SubComponentId: 0,
				SensorName: "sensor_1013_1_1", SequenceNumber: seq, SubscribedPath: "/interfaces/",
				StreamedPath: "/junos/system/linecard/interface/logical/usage/", Component: "PFE",
			}

			hdrInputExtBytes, err = proto.Marshal(&hdrInputExt)
			if err != nil {
				log.Fatalf("Error marshalling header for ext case: %v", err)
			}

			for _, ifl := range ifd.ifls {
				rValue := getRandom(interfaces.desc.IFL.INUnicastPkts, s.jtisim.random)
				inup := ifl.inUPkts + (uint64(rValue) * (freq * 1000))
				ifl.inUPkts = inup

				rValue = getRandom(interfaces.desc.IFL.INMulticastPkts, s.jtisim.random)
				inmp := ifl.inMPkts + (uint64(rValue) * (freq * 1000))
				ifl.inMPkts = inmp
				name := fmt.Sprintf("%s.%d", ifd.name, ifl.index)

				rsp := &gnmi.SubscribeResponse{
					Extension: []*gnmi_ext1.Extension{
						{
							Ext: &gnmi_ext1.Extension_RegisteredExt{
								RegisteredExt: &gnmi_ext1.RegisteredExtension{
									Id:  gnmi_ext1.ExtensionID_EID_JUNIPER_TELEMETRY_HEADER,
									Msg: hdrInputExtBytes,
								},
							},
						},
					},
					Response: &gnmi.SubscribeResponse_Update{
						Update: &gnmi.Notification{
							Timestamp: time.Now().UnixNano(),
							Prefix: &gnmi.Path{
								Origin: "",
								Elem: []*gnmi.PathElem{
									{Name: "interfaces"},
									{Name: "interface", Key: map[string]string{"name": ifd.name}},
									{Name: "subinterfaces"},
									{Name: "subinterface", Key: map[string]string{"index": strconv.Itoa(ifl.index)}},
								},
							},
							Update: []*gnmi.Update{
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "index"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: uint64(ifl.index)},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "name"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_StringVal{StringVal: name},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "counters"},
											{Name: "in-unicast-pkts"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: inup},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "counters"},
											{Name: "in-multicast-pkts"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: inmp},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "oper-status"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_StringVal{StringVal: ops},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "admin-status"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_StringVal{StringVal: ads},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv4"},
											{Name: "state"},
											{Name: "mtu"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "interface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "subinterface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "ip"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "origin"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "prefix-length"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "@ip"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "ipv4"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "ipv4"},
											{Name: "state"},
											{Name: "mtu"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "interface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "subinterface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "expiry"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "host-name"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "interface-name"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "ip"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "is-publish"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "link-layer-address"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "logical-router-id"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "neighbor-state"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "origin"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "neighbors"},
											{Name: "neighbor"},
											{Name: "state"},
											{Name: "table-id"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "state"},
											{Name: "mtu"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "interface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "subinterface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv4"},
											{Name: "unnumbered"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addressesses"},
											{Name: "address"},
											{Name: "@ip"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addressesses"},
											{Name: "address"},
											{Name: "ipv6"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addressesses"},
											{Name: "address"},
											{Name: "ipv6"},
											{Name: "state"},
											{Name: "mtu"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "interface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "interface-ref"},
											{Name: "state"},
											{Name: "subinterface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "ip"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "origin"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "prefix-length"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "addresses"},
											{Name: "address"},
											{Name: "state"},
											{Name: "status"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "state"},
											{Name: "mtu"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "interface-red"},
											{Name: "state"},
											{Name: "interface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "interface-red"},
											{Name: "state"},
											{Name: "subinterface"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "ipv6"},
											{Name: "unnumbered"},
											{Name: "state"},
											{Name: "enabled"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
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
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "counters"},
											{Name: "in-pkts"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "counters"},
											{Name: "out-octets"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "counters"},
											{Name: "out-pkts"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "description"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
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
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "ifindex"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "index"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
								{
									Path: &gnmi.Path{
										Origin: "",
										Elem: []*gnmi.PathElem{
											{Name: "state"},
											{Name: "last-change"},
										},
									},
									Val: &gnmi.TypedValue{
										Value: &gnmi.TypedValue_UintVal{UintVal: ino},
									},
								},
							},
						},
					},
				}

				seq++
				ch <- rsp

			}

		} //finish one wrap
		wrapDuration := time.Since(start)
		time.Sleep(nsFreq - wrapDuration)
	}
}
