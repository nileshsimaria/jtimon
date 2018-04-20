package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func handleOneTelemetryPkt(ocData *na_pb.OpenConfigData, jctx *jcontext) {
	updateStats(ocData, true)

	emitLog(fmt.Sprintf("system_id: %s\n", ocData.SystemId))
	emitLog(fmt.Sprintf("component_id: %d\n", ocData.ComponentId))
	emitLog(fmt.Sprintf("sub_component_id: %d\n", ocData.SubComponentId))
	emitLog(fmt.Sprintf("path: %s\n", ocData.Path))
	emitLog(fmt.Sprintf("sequence_number: %d\n", ocData.SequenceNumber))
	emitLog(fmt.Sprintf("timestamp: %d\n", ocData.Timestamp))
	emitLog(fmt.Sprintf("sync_response: %d\n", ocData.SyncResponse))
	if ocData.SyncResponse {
		fmt.Printf("Received sync_response\n")
	}

	del := ocData.GetDelete()
	for _, d := range del {
		emitLog(fmt.Sprintf("Delete: %s\n", d.GetPath()))
	}

	prefixSeen := false
	for _, kv := range ocData.Kv {
		updateStatsKV(jctx, true)

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
				if *prefixCheck {
					fmt.Printf("Missing prefix for sensor: %s\n", ocData.Path)
				}
			}
		}
	}
}

func subSendAndReceive(conn *grpc.ClientConn, jctx *jcontext, subReqM na_pb.SubscriptionRequest) {
	var ctx context.Context
	c := na_pb.NewOpenConfigTelemetryClient(conn)
	if jctx.cfg.Meta == true {
		md := metadata.New(map[string]string{"username": jctx.cfg.User, "password": jctx.cfg.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctx = context.Background()
	}
	stream, err := c.TelemetrySubscribe(ctx, &subReqM)

	if err != nil {
		log.Fatalf("Could not send RPC: %v\n", err)
	}

	hdr, errh := stream.Header()
	if errh != nil {
		log.Fatalf("Failed to get header for stream: %v", errh)
	}
	fmt.Println("gRPC headers from Junos:")
	for k, v := range hdr {
		fmt.Printf("  %s: %s\n", k, v)
	}
	if jctx.cfg.Log.LogFileName != "" {
		fmt.Printf("\nlogging Junos Telemetry data in %s ...", jctx.cfg.Log.LogFileName)
		fmt.Printf("\nto stop receiving, press CTRL+C \n")
	}
	if jctx.cfg.CStats.csv_stats {
		emitLog(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			"sensor-path", "sequence-number", "component-id", "sub-component-id", "packet-size", "p-ts", "e-ts", "re-stream-creation-ts", "re-payload-get-ts"))
	}

	/*
	 * The following for loop will run forever as long as server is up and
	 * sending data in a stream. Setup signal handler to print summary when
	 * user pressed ctrl-c
	 */
	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		if *dcheck == true {
			dropCheckCSV(jctx)
		}
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

		if *dcheck == true && !jctx.cfg.CStats.csv_stats {
			dropCheck(jctx, ocData)
		}

		if !jctx.cfg.CStats.csv_stats {
			handleOneTelemetryPkt(ocData, jctx)
		}

		if jctx.iFlux.influxc != nil && !jctx.cfg.CStats.csv_stats {
			go addIDB(ocData, jctx, rtime)
		}

		if *sleep != 0 {
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
	var additionalConfigM na_pb.SubscriptionAdditionalConfig
	cfg := jctx.cfg

	for i := range cfg.Paths {
		var pathM na_pb.Path
		pathM.Path = cfg.Paths[i].Path
		pathM.SampleFrequency = uint32(cfg.Paths[i].Freq)

		subReqM.PathList = append(subReqM.PathList, &pathM)
	}
	additionalConfigM.NeedEos = jctx.cfg.Eos
	subReqM.AdditionalConfig = &additionalConfigM
	subSendAndReceive(conn, jctx, subReqM)
}
