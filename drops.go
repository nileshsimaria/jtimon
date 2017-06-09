package main

import (
	"fmt"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"strings"
)

type dropData struct {
	seq      uint64
	received uint64
	drop     uint64
}

func dropInit(jctx *jcontext) {
	// Create a map for key ComponentID
	jctx.dMap = make(map[uint32]map[uint32]map[string]dropData)
}

func dropCheck(jctx *jcontext, ocData *na_pb.OpenConfigData) {
	var last dropData
	var new dropData
	var ok bool

	cid := ocData.ComponentId
	scid := ocData.SubComponentId
	path := ocData.Path
	seq := ocData.SequenceNumber

	_, ok = jctx.dMap[cid]
	if ok == false {
		// Create a map for key SubComponentID
		jctx.dMap[cid] = make(map[uint32]map[string]dropData)
	}

	_, ok = jctx.dMap[cid][scid]
	if ok == false {
		// Create a map for key path (sensor)
		jctx.dMap[cid][scid] = make(map[string]dropData)
	}

	last, ok = jctx.dMap[cid][scid][path]
	if ok == false {
		// A combination of (cid, scid, path) not found, create new dropData
		new.seq = seq
		new.received = 1
		new.drop = 0
		jctx.dMap[cid][scid][path] = new
	} else {
		new.seq = seq
		new.received = last.received + 1
		new.drop = last.drop
		if seq > last.seq && seq-last.seq != 1 {
			fmt.Printf("Packet Drop: path: %-120v cid: %-5v scid: %v seq: %v-%v=%v\n", path, cid, scid, seq, last.seq, seq-last.seq)
			new.drop += (seq - last.seq)
		}
		jctx.dMap[cid][scid][path] = new
	}
}

func printDropDS(jctx *jcontext) {
	st.Lock()
	fmt.Printf("\n Drops Distribution")
	fmt.Printf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	fmt.Printf("\n| CID |SCID| Drops | Received | %-120s|", "Sensor Path")
	fmt.Printf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	fmt.Printf("\n")
	for cid, sdMap := range jctx.dMap {
		for scid, pathM := range sdMap {
			for path, dData := range pathM {
				if path != "" {
					fmt.Printf("|%5v|%4v| %6v| %8v | %-120s| \n", cid, scid, dData.drop, dData.received, path)
					st.totalDdrops += dData.drop
				}
			}
		}
	}
	fmt.Printf("+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	st.Unlock()
}
