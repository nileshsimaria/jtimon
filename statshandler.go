package main

import (
	"fmt"
	"sync"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

type statsType struct {
	sync.Mutex               // guarding following stats
	startTime                time.Time
	totalIn                  uint64
	totalKV                  uint64
	totalInPayloadLength     uint64
	totalInPayloadWireLength uint64
	totalInHeaderWireLength  uint64
	totalLatency             uint64
	totalLatencyPkt          uint64
	totalDdrops              uint64
}

type statshandler struct {
	jctx *JCtx
}

func (h *statshandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

func (h *statshandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *statshandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
	case *stats.ConnEnd:
	default:
	}
}

func (h *statshandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	h.jctx.st.Lock()
	defer h.jctx.st.Unlock()

	switch s.(type) {
	case *stats.InHeader:
		h.jctx.st.totalInHeaderWireLength += uint64(s.(*stats.InHeader).WireLength)
	case *stats.OutHeader:
	case *stats.OutPayload:
	case *stats.InPayload:
		h.jctx.st.totalInPayloadLength += uint64(s.(*stats.InPayload).Length)
		h.jctx.st.totalInPayloadWireLength += uint64(s.(*stats.InPayload).WireLength)

		if h.jctx.cfg.CStats.csvStats {
			switch v := (s.(*stats.InPayload).Payload).(type) {
			case *na_pb.OpenConfigData:
				updateStats(h.jctx, v, false)
				for idx, kv := range v.Kv {
					updateStatsKV(h.jctx, false)
					switch kvvalue := kv.Value.(type) {
					case *na_pb.KeyValue_UintValue:
						if kv.Key == "__timestamp__" {
							var reCTS uint64
							var rePGetTS uint64
							if len(v.Kv) > idx+2 {
								nextKV := v.Kv[idx+1]
								if nextKV.Key == "__junos_re_stream_creation_timestamp__" {
									reCTS = nextKV.GetUintValue()
								}
								nextnextKV := v.Kv[idx+2]
								if nextnextKV.Key == "__junos_re_payload_get_timestamp__" {
									rePGetTS = nextnextKV.GetUintValue()
								}
							}
							emitLog(fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d\n",
								v.Path, v.SequenceNumber, v.ComponentId, v.SubComponentId, s.(*stats.InPayload).Length, v.Timestamp, kvvalue.UintValue, reCTS, rePGetTS))
						}
					}
				}
			}
		}

	case *stats.InTrailer:
	case *stats.End:
	default:
	}
}

func updateStats(jctx *JCtx, ocData *na_pb.OpenConfigData, needLock bool) {
	if needLock {
		jctx.st.Lock()
		defer jctx.st.Unlock()
	}

	jctx.st.totalIn++

	if *lcheck {
		now := time.Now()
		nanos := now.UnixNano()
		millis := nanos / 1000000
		if millis > int64(ocData.Timestamp) {
			jctx.st.totalLatency += uint64((millis - int64(ocData.Timestamp)))
			jctx.st.totalLatencyPkt++
		}
	}
}

func updateStatsKV(jctx *JCtx, needLock bool) {
	if needLock {
		jctx.st.Lock()
		defer jctx.st.Unlock()
	}
	jctx.st.totalKV++
}

func periodicStats(jctx *JCtx, pstats int64) {
	if pstats == 0 {
		return
	}

	i := 0
	for {
		tickChan := time.NewTicker(time.Second * time.Duration(pstats)).C
		<-tickChan

		// Do nothing if we haven't heard back anything from the device
		jctx.st.Lock()
		if jctx.st.totalIn == 0 {
			jctx.st.Unlock()
			continue
		}
		jctx.st.Unlock()

		// print header
		if i%100 == 0 {
			if *lcheck {
				fmt.Printf("%s", "+------------------------------+--------------------+--------------------+--------------------+--------------------+-----------------+\n")
				fmt.Printf("%s", "|         Timestamp            |        KV          |      Packets       |       Bytes        |     Bytes(wire)    | Average Latency |\n")
				fmt.Printf("%s", "+------------------------------+--------------------+--------------------+--------------------+--------------------+-----------------+\n")
			} else {
				fmt.Printf("%s", "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n")
				fmt.Printf("%s", "|         Timestamp            |        KV          |      Packets       |       Bytes        |     Bytes(wire)    |\n")
				fmt.Printf("%s", "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n")
			}
		}

		jctx.st.Lock()
		if *lcheck && jctx.st.totalLatencyPkt != 0 {
			fmt.Printf("| %s | %18v | %18v | %18v | %18v | %15v |\n", time.Now().Format(time.UnixDate),
				jctx.st.totalKV,
				jctx.st.totalIn,
				jctx.st.totalInPayloadLength,
				jctx.st.totalInPayloadWireLength,
				jctx.st.totalLatency/jctx.st.totalLatencyPkt)
		} else {
			fmt.Printf("| %s | %18v | %18v | %18v | %18v |\n", time.Now().Format(time.UnixDate),
				jctx.st.totalKV,
				jctx.st.totalIn,
				jctx.st.totalInPayloadLength,
				jctx.st.totalInPayloadWireLength)
		}
		jctx.st.Unlock()
		i++
	}
}

func printSummary(jctx *JCtx, pstats int64) {
	gmutex.Lock()
	defer gmutex.Unlock()

	if *dcheck == true {
		printDropDS(jctx)
	}

	endTime := time.Since(jctx.st.startTime)
	stmap := make(map[string]interface{})

	s := fmt.Sprintf("\nCollector Stats for %s[%d] (Run time : %s)\n", jctx.file, jctx.idx, endTime)
	stmap["run-time"] = float64(endTime)
	s += fmt.Sprintf("%-12v : in-packets\n", jctx.st.totalIn)
	stmap["in-packets"] = float64(jctx.st.totalIn)
	s += fmt.Sprintf("%-12v : data points (KV pairs)\n", jctx.st.totalKV)
	stmap["kv"] = float64(jctx.st.totalKV)

	s += fmt.Sprintf("%-12v : in-header wirelength (bytes)\n", jctx.st.totalInHeaderWireLength)
	stmap["in-header-wire-length"] = float64(jctx.st.totalInHeaderWireLength)
	s += fmt.Sprintf("%-12v : in-payload length (bytes)\n", jctx.st.totalInPayloadLength)
	stmap["in-payload-length-bytes"] = float64(jctx.st.totalInPayloadLength)
	s += fmt.Sprintf("%-12v : in-payload wirelength (bytes)\n", jctx.st.totalInPayloadWireLength)
	stmap["in-payload-wirelength-bytes"] = float64(jctx.st.totalInPayloadWireLength)
	if uint64(endTime.Seconds()) != 0 {
		s += fmt.Sprintf("%-12v : throughput (bytes per seconds)\n", jctx.st.totalInPayloadLength/uint64(endTime.Seconds()))
		stmap["throughput"] = float64(jctx.st.totalInPayloadLength / uint64(endTime.Seconds()))
	}

	if *lcheck && jctx.st.totalLatencyPkt != 0 {
		s += fmt.Sprintf("%-12v : latency sample packets\n", jctx.st.totalLatencyPkt)
		stmap["latency-sample-packets"] = float64(jctx.st.totalLatencyPkt)
		s += fmt.Sprintf("%-12v : latency (ms)\n", jctx.st.totalLatency)
		stmap["total-latency"] = float64(jctx.st.totalLatency)
		s += fmt.Sprintf("%-12v : average latency (ms)\n", jctx.st.totalLatency/jctx.st.totalLatencyPkt)
		stmap["average-latency"] = float64(jctx.st.totalLatency / jctx.st.totalLatencyPkt)
	}

	if *dcheck == true {
		s += fmt.Sprintf("%-12v : total packet drops\n", jctx.st.totalDdrops)
		stmap["total-drops"] = float64(jctx.st.totalDdrops)
	}

	s += fmt.Sprintf("\n")
	fmt.Printf("\n%s\n", s)

	addIDBSummary(jctx, stmap)
}
