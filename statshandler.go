package main

import (
	"fmt"
	"sync"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

type statsCtx struct {
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
	h.jctx.stats.Lock()
	defer h.jctx.stats.Unlock()

	switch s.(type) {
	case *stats.InHeader:
		h.jctx.stats.totalInHeaderWireLength += uint64(s.(*stats.InHeader).WireLength)
	case *stats.OutHeader:
	case *stats.OutPayload:
	case *stats.InPayload:
		h.jctx.stats.totalInPayloadLength += uint64(s.(*stats.InPayload).Length)
		h.jctx.stats.totalInPayloadWireLength += uint64(s.(*stats.InPayload).WireLength)

		if h.jctx.config.Log.CSVStats {
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
							jLog(h.jctx, fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d\n",
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
	if !*stateHandler {
		return
	}
	if needLock {
		jctx.stats.Lock()
		defer jctx.stats.Unlock()
	}

	jctx.stats.totalIn++

	if jctx.config.Log.LatencyCheck {
		now := time.Now()
		nanos := now.UnixNano()
		millis := nanos / 1000000
		if millis > int64(ocData.Timestamp) {
			jctx.stats.totalLatency += uint64((millis - int64(ocData.Timestamp)))
			jctx.stats.totalLatencyPkt++
		}
	}
}

func updateStatsKV(jctx *JCtx, needLock bool) {
	if !*stateHandler {
		return
	}

	if needLock {
		jctx.stats.Lock()
		defer jctx.stats.Unlock()
	}
	jctx.stats.totalKV++
}

func periodicStats(jctx *JCtx) {
	if !*stateHandler {
		return
	}
	pstats := jctx.config.Log.PeriodicStats
	if pstats == 0 {
		return
	}

	headerCounter := 0
	for {
		tickChan := time.NewTicker(time.Second * time.Duration(pstats)).C
		<-tickChan

		// Do nothing if we haven't heard back anything from the device

		jctx.stats.Lock()
		if jctx.stats.totalIn == 0 {
			jctx.stats.Unlock()
			continue
		}

		s := fmt.Sprintf("\n")

		// print header
		if headerCounter%100 == 0 {
			if jctx.config.Log.LatencyCheck {
				s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+-----------------+\n"
				s += "|         Timestamp            |        KV          |      Packets       |       Bytes        |     Bytes(wire)    | Average Latency |\n"
				s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+-----------------+\n"
			} else {
				s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n"
				s += "|         Timestamp            |        KV          |      Packets       |       Bytes        |     Bytes(wire)    |\n"
				s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n"
			}
		}

		if jctx.config.Log.LatencyCheck && jctx.stats.totalLatencyPkt != 0 {
			s += fmt.Sprintf("| %s | %18v | %18v | %18v | %18v | %15v |\n", time.Now().Format(time.UnixDate),
				jctx.stats.totalKV,
				jctx.stats.totalIn,
				jctx.stats.totalInPayloadLength,
				jctx.stats.totalInPayloadWireLength,
				jctx.stats.totalLatency/jctx.stats.totalLatencyPkt)
		} else {
			s += fmt.Sprintf("| %s | %18v | %18v | %18v | %18v |\n", time.Now().Format(time.UnixDate),
				jctx.stats.totalKV,
				jctx.stats.totalIn,
				jctx.stats.totalInPayloadLength,
				jctx.stats.totalInPayloadWireLength)
		}
		jctx.stats.Unlock()
		headerCounter++
		if s != "" {
			jLog(jctx, fmt.Sprintf("%s\n", s))
		}
	}
}

func printSummary(jctx *JCtx) {
	if !*stateHandler {
		return
	}

	if jctx.config.Log.CSVStats && jctx.config.Log.DropCheck {
		dropCheckCSV(jctx)
	}

	if jctx.config.Log.DropCheck {
		printDropDS(jctx)
	}

	endTime := time.Since(jctx.stats.startTime)
	stmap := make(map[string]interface{})

	s := fmt.Sprintf("\nCollector Stats for %s:%d (Run time : %s)\n", jctx.config.Host, jctx.config.Port, endTime)
	stmap["run-time"] = float64(endTime)
	s += fmt.Sprintf("%-12v : in-packets\n", jctx.stats.totalIn)
	stmap["in-packets"] = float64(jctx.stats.totalIn)
	s += fmt.Sprintf("%-12v : data points (KV pairs)\n", jctx.stats.totalKV)
	stmap["kv"] = float64(jctx.stats.totalKV)

	s += fmt.Sprintf("%-12v : in-header wirelength (bytes)\n", jctx.stats.totalInHeaderWireLength)
	stmap["in-header-wire-length"] = float64(jctx.stats.totalInHeaderWireLength)
	s += fmt.Sprintf("%-12v : in-payload length (bytes)\n", jctx.stats.totalInPayloadLength)
	stmap["in-payload-length-bytes"] = float64(jctx.stats.totalInPayloadLength)
	s += fmt.Sprintf("%-12v : in-payload wirelength (bytes)\n", jctx.stats.totalInPayloadWireLength)
	stmap["in-payload-wirelength-bytes"] = float64(jctx.stats.totalInPayloadWireLength)
	if uint64(endTime.Seconds()) != 0 {
		s += fmt.Sprintf("%-12v : throughput (bytes per seconds)\n", jctx.stats.totalInPayloadLength/uint64(endTime.Seconds()))
		stmap["throughput"] = float64(jctx.stats.totalInPayloadLength / uint64(endTime.Seconds()))
	}

	if jctx.config.Log.LatencyCheck && jctx.stats.totalLatencyPkt != 0 {
		s += fmt.Sprintf("%-12v : latency sample packets\n", jctx.stats.totalLatencyPkt)
		stmap["latency-sample-packets"] = float64(jctx.stats.totalLatencyPkt)
		s += fmt.Sprintf("%-12v : latency (ms)\n", jctx.stats.totalLatency)
		stmap["total-latency"] = float64(jctx.stats.totalLatency)
		s += fmt.Sprintf("%-12v : average latency (ms)\n", jctx.stats.totalLatency/jctx.stats.totalLatencyPkt)
		stmap["average-latency"] = float64(jctx.stats.totalLatency / jctx.stats.totalLatencyPkt)
	}

	if jctx.config.Log.DropCheck {
		s += fmt.Sprintf("%-12v : total packet drops\n", jctx.stats.totalDdrops)
		stmap["total-drops"] = float64(jctx.stats.totalDdrops)
	}

	s += fmt.Sprintf("\n")
	jLog(jctx, fmt.Sprintf("\n%s\n", s))

	addIDBSummary(jctx, stmap)
}
