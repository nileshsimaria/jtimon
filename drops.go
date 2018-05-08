package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

type dropData struct {
	seq      uint64
	received uint64
	drop     uint64
}

func dropInit(jctx *JCtx) {
	jctx.dMap = make(map[uint32]map[uint32]map[string]dropData)
}

func dropCheckCSV(jctx *JCtx) {
	if !jctx.config.Log.CSVStats {
		return
	}
	f := jctx.config.Log.FileHandle
	if jctx.config.Log.FileHandle == nil {
		return
	}
	f.Seek(0, 0)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "sensor-path,sequence-number,component-id,sub-component-id,packet-size,p-ts,e-ts") {
			tokens := strings.Split(line, ",")
			//fmt.Printf("\n%s + %s + %s + %s + %s + %s + %s", tokens[0], tokens[1], tokens[2], tokens[3], tokens[4], tokens[5], tokens[6])
			cid, _ := strconv.ParseUint(tokens[2], 10, 32)
			scid, _ := strconv.ParseUint(tokens[3], 10, 32)
			seqNum, _ := strconv.ParseUint(tokens[1], 10, 32)

			dropCheckWork(jctx, uint32(cid), uint32(scid), tokens[0], seqNum)
		}

	}
}

func dropCheckWork(jctx *JCtx, cid uint32, scid uint32, path string, seq uint64) {
	var last dropData
	var new dropData
	var ok bool

	_, ok = jctx.dMap[cid]
	if ok == false {
		jctx.dMap[cid] = make(map[uint32]map[string]dropData)
	}

	_, ok = jctx.dMap[cid][scid]
	if ok == false {
		jctx.dMap[cid][scid] = make(map[string]dropData)
	}

	last, ok = jctx.dMap[cid][scid][path]
	if ok == false {
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

func dropCheck(jctx *JCtx, ocData *na_pb.OpenConfigData) {
	dropCheckWork(jctx, ocData.ComponentId, ocData.SubComponentId, ocData.Path, ocData.SequenceNumber)
}

func printDropDS(jctx *JCtx) {
	jctx.stats.Lock()
	s := fmt.Sprintf("\nDrops Distribution for %s:%d", jctx.config.Host, jctx.config.Port)
	s += fmt.Sprintf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	s += fmt.Sprintf("\n| CID |SCID| Drops | Received | %-120s|", "Sensor Path")
	s += fmt.Sprintf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	s += fmt.Sprintf("\n")
	for cid, sdMap := range jctx.dMap {
		for scid, pathM := range sdMap {
			for path, dData := range pathM {
				if path != "" {
					s += fmt.Sprintf("|%5v|%4v| %6v| %8v | %-120s| \n", cid, scid, dData.drop, dData.received, path)
					jctx.stats.totalDdrops += dData.drop
				}
			}
		}
	}
	s += fmt.Sprintf("+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	// Log it. The caller printSummary() has already taken global mutex log so dont try to get it again here.
	l(false, jctx, s)
	jctx.stats.Unlock()
}
