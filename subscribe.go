package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func printPacket(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	updateStats(jctx, ocData, true)

	emitLog(jctx, fmt.Sprintf("system_id: %s\n", ocData.SystemId))
	emitLog(jctx, fmt.Sprintf("component_id: %d\n", ocData.ComponentId))
	emitLog(jctx, fmt.Sprintf("sub_component_id: %d\n", ocData.SubComponentId))
	emitLog(jctx, fmt.Sprintf("path: %s\n", ocData.Path))
	emitLog(jctx, fmt.Sprintf("sequence_number: %d\n", ocData.SequenceNumber))
	emitLog(jctx, fmt.Sprintf("timestamp: %d\n", ocData.Timestamp))
	emitLog(jctx, fmt.Sprintf("sync_response: %v\n", ocData.SyncResponse))
	if ocData.SyncResponse {
		fmt.Printf("Received sync_response\n")
	}

	del := ocData.GetDelete()
	for _, d := range del {
		emitLog(jctx, fmt.Sprintf("Delete: %s\n", d.GetPath()))
	}

	prefixSeen := false
	for _, kv := range ocData.Kv {
		updateStatsKV(jctx, true)

		emitLog(jctx, fmt.Sprintf("  key: %s\n", kv.Key))
		switch value := kv.Value.(type) {
		case *na_pb.KeyValue_DoubleValue:
			emitLog(jctx, fmt.Sprintf("  double_value: %v\n", value.DoubleValue))
		case *na_pb.KeyValue_IntValue:
			emitLog(jctx, fmt.Sprintf("  int_value: %d\n", value.IntValue))
		case *na_pb.KeyValue_UintValue:
			emitLog(jctx, fmt.Sprintf("  uint_value: %d\n", value.UintValue))
		case *na_pb.KeyValue_SintValue:
			emitLog(jctx, fmt.Sprintf("  sint_value: %d\n", value.SintValue))
		case *na_pb.KeyValue_BoolValue:
			emitLog(jctx, fmt.Sprintf("  bool_value: %v\n", value.BoolValue))
		case *na_pb.KeyValue_StrValue:
			emitLog(jctx, fmt.Sprintf("  str_value: %s\n", value.StrValue))
		case *na_pb.KeyValue_BytesValue:
			emitLog(jctx, fmt.Sprintf("  bytes_value: %s\n", value.BytesValue))
		default:
			emitLog(jctx, fmt.Sprintf("  default: %v\n", value))
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

func subSendAndReceive(conn *grpc.ClientConn, jctx *JCtx, subReqM na_pb.SubscriptionRequest) {
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
		SafePrint(fmt.Sprintf("Could not send RPC: %v\n", err))
		return
	}

	hdr, errh := stream.Header()
	if errh != nil {
		SafePrint(fmt.Sprintf("Failed to get header for stream: %v", errh))
	}

	gmutex.Lock()
	fmt.Printf("gRPC headers from Junos for %s[%d]\n", jctx.file, jctx.idx)
	for k, v := range hdr {
		fmt.Printf("  %s: %s\n", k, v)
	}
	gmutex.Unlock()

	if jctx.cfg.CStats.csvStats {
		emitLog(jctx, fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			"sensor-path", "sequence-number", "component-id", "sub-component-id", "packet-size", "p-ts", "e-ts", "re-stream-creation-ts", "re-payload-get-ts"))
	}

	for {
		ocData, err := stream.Recv()
		if err == io.EOF {
			printSummary(jctx, *pstats)
			break
		}
		if err != nil {
			SafePrint(fmt.Sprintf("%v.TelemetrySubscribe(_) = _, %v", conn, err))
			return
		}

		rtime := time.Now()

		if *dcheck == true && !jctx.cfg.CStats.csvStats {
			dropCheck(jctx, ocData)
		}

		gmutex.Lock()
		printPacket(ocData, jctx)
		gmutex.Unlock()

		if jctx.iFlux.influxc != nil && !jctx.cfg.CStats.csvStats {
			go addIDB(ocData, jctx, rtime)
		}

		if *sleep != 0 {
			time.Sleep(time.Duration(*sleep) * time.Millisecond)
		}

		select {
		case pfor := <-jctx.pause.pch:
			SafePrint(fmt.Sprintf("Pausing for %v seconds\n", pfor))
			t := time.NewTimer(time.Second * time.Duration(pfor))
			select {
			case <-t.C:
				SafePrint(fmt.Sprintf("Done pausing for %v seconds\n", pfor))
			case <-jctx.pause.upch:
				t.Stop()
			}
		default:
		}
	}
}

func subscribe(conn *grpc.ClientConn, jctx *JCtx) {
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
