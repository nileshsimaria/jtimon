package main

import (
	"fmt"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func handleOneTelemetryPkt(ocData *na_pb.OpenConfigData, jctx *jcontext) {
	if *dcheck == true {
		dropCheck(jctx, ocData)
	}

	emitLog(fmt.Sprintf("system_id: %s\n", ocData.SystemId))
	emitLog(fmt.Sprintf("component_id: %d\n", ocData.ComponentId))
	emitLog(fmt.Sprintf("sub_component_id: %d\n", ocData.SubComponentId))
	emitLog(fmt.Sprintf("path: %s\n", ocData.Path))
	emitLog(fmt.Sprintf("sequence_number: %d\n", ocData.SequenceNumber))
	emitLog(fmt.Sprintf("timestamp: %d\n", ocData.Timestamp))

	updateStats(ocData)

	prefixSeen := false
	for _, kv := range ocData.Kv {
		updateStatsKV(jctx)

		emitLog(fmt.Sprintf("  key: %s\n", kv.Key))
		switch value := kv.Value.(type) {
		case *na_pb.KeyValue_DoubleValue:
			emitLog(fmt.Sprintf("  double_value: %d\n", value.DoubleValue))
		case *na_pb.KeyValue_IntValue:
			emitLog(fmt.Sprintf("  int_value: %d\n", value.IntValue))
		case *na_pb.KeyValue_UintValue:
			emitLog(fmt.Sprintf("  uint_value: %d\n", value.UintValue))
		case *na_pb.KeyValue_SintValue:
			emitLog(fmt.Sprintf("  sint_value: %d\n", value.SintValue))
		case *na_pb.KeyValue_BoolValue:
			emitLog(fmt.Sprintf("  bool_value: %s\n", value.BoolValue))
		case *na_pb.KeyValue_StrValue:
			emitLog(fmt.Sprintf("  str_value: %s\n", value.StrValue))
		case *na_pb.KeyValue_BytesValue:
			emitLog(fmt.Sprintf("  bytes_value: %s\n", value.BytesValue))
		default:
			emitLog(fmt.Sprintf("  default: %v\n", value))
		}

		if kv.Key == "__prefix__" {
			prefixSeen = true
		} else if !strings.HasPrefix(kv.Key, "__") {
			if !prefixSeen && !strings.HasPrefix(kv.Key, "/") {
				fmt.Printf("Missing prefix for sensor: %s\n", ocData.Path)
			}
		}
	}
}

func subSendAndReceive(conn *grpc.ClientConn, jctx *jcontext, subReqM na_pb.SubscriptionRequest) {
	c := na_pb.NewOpenConfigTelemetryClient(conn)
	stream, err := c.TelemetrySubscribe(
		context.Background(),
		&subReqM)

	if err != nil {
		log.Fatalf("Could not send RPC: %v\n", err)
	}

	hdr, errh := stream.Header()
	if errh != nil {
		log.Fatalf("Failed to get header for stream: %v", errh)
	}
	emitLog("\nHeaders:\n")
	for k, v := range hdr {
		emitLog(fmt.Sprintf("  %s: %s\n", k, v))
	}
	log.Printf("\n")

	/*
	 * The following for loop will run forever as long as server is up and
	 * sending data in a stream. Setup signal handler to print summary when
	 * user pressed ctrl-c
	 */
	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		printSummary(jctx, *pstats)
		os.Exit(0)
	}()

	for {
		ocData, err := stream.Recv()
		if err == io.EOF {
			printSummary(jctx, *pstats)
			break
		}
		if err != nil {
			log.Fatalf("%v.TelemetrySubscribe(_) = _, %v", conn, err)
		}

		rtime := time.Now()

		handleOneTelemetryPkt(ocData, jctx)
		if jctx.iFlux.influxc != nil {
			go addIDB(ocData, jctx, rtime)
		}

		if *sleep != 0 {
			fmt.Printf("Sleeping for %v Milliseconds\n", *sleep)
			time.Sleep(time.Duration(*sleep) * time.Millisecond)
		}

		select {
		case pfor := <-jctx.pause.pch:
			fmt.Printf("Pausing for %v seconds\n", pfor)
			t := time.NewTimer(time.Second * time.Duration(pfor))
			select {
			case <-t.C:
				fmt.Printf("Done pausing for %v seconds\n", pfor)
			case <-jctx.pause.upch:
				t.Stop()
			}
		default:
		}
	}
}

func subscribe(conn *grpc.ClientConn, jctx *jcontext) {
	var subReqM na_pb.SubscriptionRequest
	cfg := jctx.cfg

	for i := range cfg.Paths {
		var pathM na_pb.Path
		pathM.Path = cfg.Paths[i].Path
		pathM.SampleFrequency = cfg.Paths[i].Freq

		subReqM.PathList = append(subReqM.PathList, &pathM)
	}
	subSendAndReceive(conn, jctx, subReqM)
}
