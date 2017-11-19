package main

import (
	"fmt"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
	"os"
	"sync"
	"time"
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
	jctx *jcontext
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
	st.Lock()
	defer st.Unlock()

	switch s.(type) {
	case *stats.InHeader:
		st.totalInHeaderWireLength += uint64(s.(*stats.InHeader).WireLength)
	case *stats.OutHeader:
	case *stats.OutPayload:
	case *stats.InPayload:
		st.totalInPayloadLength += uint64(s.(*stats.InPayload).Length)
		st.totalInPayloadWireLength += uint64(s.(*stats.InPayload).WireLength)

		if h.jctx.cfg.CStats.csv_stats {
			switch v := (s.(*stats.InPayload).Payload).(type) {
			case *na_pb.OpenConfigData:
				updateStats(v, false)
				for _, kv := range v.Kv {
					updateStatsKV(h.jctx, false)
					switch kvvalue := kv.Value.(type) {
					case *na_pb.KeyValue_UintValue:
						if kv.Key == "__timestamp__" {
							emitLog(fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d\n",
								v.Path, v.SequenceNumber, v.ComponentId, v.SubComponentId, s.(*stats.InPayload).Length, v.Timestamp, kvvalue.UintValue))
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

func updateStats(ocData *na_pb.OpenConfigData, need_lock bool) {
	if need_lock {
		st.Lock()
		defer st.Unlock()
	}

	st.totalIn++

	if *lcheck {
		now := time.Now()
		nanos := now.UnixNano()
		millis := nanos / 1000000
		if millis > int64(ocData.Timestamp) {
			st.totalLatency += uint64((millis - int64(ocData.Timestamp)))
			st.totalLatencyPkt++
		}
	}
}

func updateStatsKV(jctx *jcontext, need_lock bool) {
	if need_lock {
		st.Lock()
		defer st.Unlock()
	}
	st.totalKV++

	if *maxKV != 0 && st.totalKV >= *maxKV {
		st.Unlock()
		printSummary(jctx, *pstats)
		os.Exit(0)
	}
}

func periodicStats(pstats int64) {
	if pstats == 0 {
		return
	}

	i := 0
	for {
		tickChan := time.NewTicker(time.Second * time.Duration(pstats)).C
		<-tickChan

		// Do nothing if we haven't heard back anything from the device
		st.Lock()
		if st.totalIn == 0 {
			st.Unlock()
			continue
		}
		st.Unlock()

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

		st.Lock()
		if *lcheck && st.totalLatencyPkt != 0 {
			fmt.Printf("| %s | %18v | %18v | %18v | %18v | %15v |\n", time.Now().Format(time.UnixDate),
				st.totalKV,
				st.totalIn,
				st.totalInPayloadLength,
				st.totalInPayloadWireLength,
				st.totalLatency/st.totalLatencyPkt)
		} else {
			fmt.Printf("| %s | %18v | %18v | %18v | %18v |\n", time.Now().Format(time.UnixDate),
				st.totalKV,
				st.totalIn,
				st.totalInPayloadLength,
				st.totalInPayloadWireLength)
		}
		st.Unlock()
		i++
	}
}

func printSummary(jctx *jcontext, pstats int64) {

	if *dcheck == true {
		printDropDS(jctx)
	}
	st.Lock()
	endTime := time.Since(st.startTime)
	stmap := make(map[string]interface{})

	s := fmt.Sprintf("\nCollector Stats (Run time : %s)\n", endTime)
	stmap["run-time"] = endTime
	s += fmt.Sprintf("%-12v : in-packets\n", st.totalIn)
	stmap["in-packets"] = st.totalIn
	s += fmt.Sprintf("%-12v : data points (KV pairs)\n", st.totalKV)
	stmap["kv"] = st.totalKV

	s += fmt.Sprintf("%-12v : in-header wirelength (bytes)\n", st.totalInHeaderWireLength)
	stmap["in-header-wire-length"] = st.totalInHeaderWireLength
	s += fmt.Sprintf("%-12v : in-payload length (bytes)\n", st.totalInPayloadLength)
	stmap["in-payload-length-bytes"] = st.totalInPayloadLength
	s += fmt.Sprintf("%-12v : in-payload wirelength (bytes)\n", st.totalInPayloadWireLength)
	stmap["in-payload-wirelength-bytes"] = st.totalInPayloadWireLength
	if endTime.Seconds() != 0 {
		s += fmt.Sprintf("%-12v : throughput (bytes per seconds)\n", st.totalInPayloadLength/uint64(endTime.Seconds()))
		stmap["throughput"] = st.totalInPayloadLength / uint64(endTime.Seconds())
	}

	if *lcheck && st.totalLatencyPkt != 0 {
		s += fmt.Sprintf("%-12v : latency sample packets\n", st.totalLatencyPkt)
		stmap["latency-sample-packets"] = st.totalLatencyPkt
		s += fmt.Sprintf("%-12v : latency (ms)\n", st.totalLatency)
		stmap["total-latency"] = st.totalLatency
		s += fmt.Sprintf("%-12v : average latency (ms)\n", st.totalLatency/st.totalLatencyPkt)
		stmap["average-latency"] = st.totalLatency / st.totalLatencyPkt
	}

	if *dcheck == true {
		s += fmt.Sprintf("%-12v : total packet drops\n", st.totalDdrops)
		stmap["total-drops"] = st.totalDdrops
	}

	s += fmt.Sprintf("\n")
	fmt.Printf("\n%s\n", s)
	st.Unlock()

	addIDBSummary(jctx, stmap)

	if *td == true {
		influxDBQueryString(jctx)
	}
}
