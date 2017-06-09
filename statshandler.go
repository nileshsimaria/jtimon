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
	totalIn                  int64
	totalKV                  int64
	totalInPayloadLength     uint64
	totalInPayloadWireLength uint64
	totalInHeaderWireLength  uint64
	totalLatency             uint64
	totalLatencyPkt          uint64
	totalDdrops              uint64
}

type statshandler struct {
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

		if uint64(s.(*stats.InPayload).Length) > 16384 {
			//fmt.Printf("SH: Payload Length: %v\n", uint64(s.(*stats.InPayload).Length))
		}

	case *stats.InTrailer:
	case *stats.End:
	default:
	}
}

func updateStats(ocData *na_pb.OpenConfigData) {
	st.Lock()
	defer st.Unlock()

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

func updateStatsKV(jctx *jcontext) {
	st.Lock()
	defer st.Unlock()
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

	s := fmt.Sprintf("\nCollector Stats (Run time : %s)\n", endTime)
	s += fmt.Sprintf("%-12v : in-packets\n", st.totalIn)
	s += fmt.Sprintf("%-12v : data points (KV pairs)\n", st.totalKV)

	if pstats != 0 {
		s += fmt.Sprintf("%-12v : in-header wirelength (bytes)\n", st.totalInHeaderWireLength)
		s += fmt.Sprintf("%-12v : in-payload length (bytes)\n", st.totalInPayloadLength)
		s += fmt.Sprintf("%-12v : in-payload wirelength (bytes)\n", st.totalInPayloadWireLength)
		if endTime.Seconds() != 0 {
			s += fmt.Sprintf("%-12v : throughput (bytes per seconds)\n", st.totalInPayloadLength/uint64(endTime.Seconds()))
		}
	}

	if *lcheck && st.totalLatencyPkt != 0 {
		s += fmt.Sprintf("%-12v : latency sample packets\n", st.totalLatencyPkt)
		s += fmt.Sprintf("%-12v : latency (ms)\n", st.totalLatency)
		s += fmt.Sprintf("%-12v : average latency (ms)\n", st.totalLatency/st.totalLatencyPkt)
	}

	if *dcheck == true {
		s += fmt.Sprintf("%-12v : total packet drops\n", st.totalDdrops)
	}

	s += fmt.Sprintf("\n")
	fmt.Printf("\n%s\n", s)
	st.Unlock()

	if *td == true {
		influxDBQueryString(jctx)
	}
}
