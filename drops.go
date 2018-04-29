package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	// Create a map for key ComponentID
	jctx.dMap = make(map[uint32]map[uint32]map[string]dropData)
}

func dropCheckCSV(jctx *JCtx) {
	if !jctx.cfg.CStats.csvStats {
		return
	}

	if jctx.cfg.Log.File == "" {
		return
	}
	if err := jctx.cfg.Log.FileHandle.Close(); err != nil {
		log.Fatalf("Could not close csv data log file(%s): %v\n", jctx.cfg.Log.File, err)
	}

	f, err := os.Open(jctx.cfg.Log.File)
	if err != nil {
		log.Fatalf("Could not open csv data log file(%s) for drop-check: %v\n", jctx.cfg.Log.File, err)
	}
	defer f.Close()

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

func dropCheck(jctx *JCtx, ocData *na_pb.OpenConfigData) {
	dropCheckWork(jctx, ocData.ComponentId, ocData.SubComponentId, ocData.Path, ocData.SequenceNumber)
}

func printDropDS(jctx *JCtx) {
	jctx.st.Lock()
	s := fmt.Sprintf("\nDrops Distribution for %s:%d", jctx.cfg.Host, jctx.cfg.Port)
	s += fmt.Sprintf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	s += fmt.Sprintf("\n| CID |SCID| Drops | Received | %-120s|", "Sensor Path")
	s += fmt.Sprintf("\n+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	s += fmt.Sprintf("\n")
	for cid, sdMap := range jctx.dMap {
		for scid, pathM := range sdMap {
			for path, dData := range pathM {
				if path != "" {
					s += fmt.Sprintf("|%5v|%4v| %6v| %8v | %-120s| \n", cid, scid, dData.drop, dData.received, path)
					jctx.st.totalDdrops += dData.drop
				}
			}
		}
	}
	s += fmt.Sprintf("+----+-----+-------+----------+%s+", strings.Repeat("-", 121))
	l(false, jctx, s)
	jctx.st.Unlock()
}
