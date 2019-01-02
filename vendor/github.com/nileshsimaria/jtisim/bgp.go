package jtisim

import (
	"log"
	"time"

	tpb "github.com/nileshsimaria/jtimon/telemetry"
)

func (s *server) streamBGP(ch chan *tpb.OpenConfigData, path *tpb.Path) {
	pname := path.GetPath()
	freq := path.GetSampleFrequency()
	log.Println(pname, freq)

	seq := uint64(0)
	for {
		kv := []*tpb.KeyValue{
			{Key: "__prefix__", Value: &tpb.KeyValue_StrValue{StrValue: "/bgp"}},
			{Key: "state/foo", Value: &tpb.KeyValue_UintValue{UintValue: 1111}},
		}

		d := &tpb.OpenConfigData{
			SystemId:       "jtisim",
			ComponentId:    2,
			Timestamp:      uint64(MakeMSTimestamp()),
			SequenceNumber: seq,
			Kv:             kv,
		}
		ch <- d
		time.Sleep(time.Duration(freq) * time.Millisecond)
		seq++
	}
}
