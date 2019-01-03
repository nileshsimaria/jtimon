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
			s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n"
			s += "|         Timestamp            |        KV          |      Packets       |       Bytes        |     Bytes(wire)    |\n"
			s += "+------------------------------+--------------------+--------------------+--------------------+--------------------+\n"
		}

		s += fmt.Sprintf("| %s | %18v | %18v | %18v | %18v |\n", time.Now().Format(time.UnixDate),
			jctx.stats.totalKV,
			jctx.stats.totalIn,
			jctx.stats.totalInPayloadLength,
			jctx.stats.totalInPayloadWireLength)
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

	endTime := time.Since(jctx.stats.startTime)

	s := fmt.Sprintf("\nCollector Stats for %s:%d (Run time : %s)\n", jctx.config.Host, jctx.config.Port, endTime)
	s += fmt.Sprintf("%-12v : in-packets\n", jctx.stats.totalIn)
	s += fmt.Sprintf("%-12v : data points (KV pairs)\n", jctx.stats.totalKV)

	s += fmt.Sprintf("%-12v : in-header wirelength (bytes)\n", jctx.stats.totalInHeaderWireLength)
	s += fmt.Sprintf("%-12v : in-payload length (bytes)\n", jctx.stats.totalInPayloadLength)
	s += fmt.Sprintf("%-12v : in-payload wirelength (bytes)\n", jctx.stats.totalInPayloadWireLength)
	if uint64(endTime.Seconds()) != 0 {
		s += fmt.Sprintf("%-12v : throughput (bytes per seconds)\n", jctx.stats.totalInPayloadLength/uint64(endTime.Seconds()))
	}

	s += fmt.Sprintf("\n")
	jLog(jctx, fmt.Sprintf("\n%s\n", s))
}
