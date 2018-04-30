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

func handleOnePacket(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	fmt.Println("handleOnePacket")
	updateStats(jctx, ocData, true)

	if *print || (jctx.cfg.Log.Verbose && !*print) {
		l(false, jctx, fmt.Sprintf("system_id: %s\n", ocData.SystemId))
		l(false, jctx, fmt.Sprintf("component_id: %d\n", ocData.ComponentId))
		l(false, jctx, fmt.Sprintf("sub_component_id: %d\n", ocData.SubComponentId))
		l(false, jctx, fmt.Sprintf("path: %s\n", ocData.Path))
		l(false, jctx, fmt.Sprintf("sequence_number: %d\n", ocData.SequenceNumber))
		l(false, jctx, fmt.Sprintf("timestamp: %d\n", ocData.Timestamp))
		l(false, jctx, fmt.Sprintf("sync_response: %v\n", ocData.SyncResponse))
		if ocData.SyncResponse {
			if *print || (jctx.cfg.Log.Verbose && !*print) {
				l(false, jctx, "Received sync_response\n")
			}
		}

		del := ocData.GetDelete()
		for _, d := range del {
			if *print || (jctx.cfg.Log.Verbose && !*print) {
				l(false, jctx, fmt.Sprintf("Delete: %s\n", d.GetPath()))
			}
		}
	}

	prefixSeen := false
	for _, kv := range ocData.Kv {
		updateStatsKV(jctx, true)

		if *print || (jctx.cfg.Log.Verbose && !*print) {
			l(false, jctx, fmt.Sprintf("  key: %s\n", kv.Key))
			switch value := kv.Value.(type) {
			case *na_pb.KeyValue_DoubleValue:
				l(false, jctx, fmt.Sprintf("  double_value: %v\n", value.DoubleValue))
			case *na_pb.KeyValue_IntValue:
				l(false, jctx, fmt.Sprintf("  int_value: %d\n", value.IntValue))
			case *na_pb.KeyValue_UintValue:
				l(false, jctx, fmt.Sprintf("  uint_value: %d\n", value.UintValue))
			case *na_pb.KeyValue_SintValue:
				l(false, jctx, fmt.Sprintf("  sint_value: %d\n", value.SintValue))
			case *na_pb.KeyValue_BoolValue:
				l(false, jctx, fmt.Sprintf("  bool_value: %v\n", value.BoolValue))
			case *na_pb.KeyValue_StrValue:
				l(false, jctx, fmt.Sprintf("  str_value: %s\n", value.StrValue))
			case *na_pb.KeyValue_BytesValue:
				l(false, jctx, fmt.Sprintf("  bytes_value: %s\n", value.BytesValue))
			default:
				l(false, jctx, fmt.Sprintf("  default: %v\n", value))
			}
		}

		if kv.Key == "__prefix__" {
			prefixSeen = true
		} else if !strings.HasPrefix(kv.Key, "__") {
			if !prefixSeen && !strings.HasPrefix(kv.Key, "/") {
				if *prefixCheck {
					l(false, jctx, fmt.Sprintf("Missing prefix for sensor: %s\n", ocData.Path))
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
		l(true, jctx, fmt.Sprintf("Could not send RPC: %v\n", err))
		return
	}

	hdr, errh := stream.Header()
	if errh != nil {
		l(true, jctx, fmt.Sprintf("Failed to get header for stream: %v", errh))
	}

	if !jctx.cfg.Log.CSVStats {
		gmutex.Lock()
		l(false, jctx, fmt.Sprintf("gRPC headers from host %s:%d\n", jctx.cfg.Host, jctx.cfg.Port))
		for k, v := range hdr {
			l(false, jctx, fmt.Sprintf("  %s: %s\n", k, v))
		}
		l(false, jctx, fmt.Sprintf("Receiving telemetry data from %s:%d\n", jctx.cfg.Host, jctx.cfg.Port))
		gmutex.Unlock()
	}

	if jctx.cfg.Log.CSVStats {
		l(true, jctx, fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			"sensor-path", "sequence-number", "component-id", "sub-component-id", "packet-size", "p-ts", "e-ts", "re-stream-creation-ts", "re-payload-get-ts"))
	}

	for {
		ocData, err := stream.Recv()
		if err == io.EOF {
			printSummary(jctx)
			break
		}
		if err != nil {
			l(true, jctx, fmt.Sprintf("%v.TelemetrySubscribe(_) = _, %v", conn, err))
			return
		}

		rtime := time.Now()

		if jctx.cfg.Log.DropCheck && !jctx.cfg.Log.CSVStats {
			dropCheck(jctx, ocData)
		}

		if *print || *stateHandler || jctx.cfg.Log.Verbose {
			gmutex.Lock()
			handleOnePacket(ocData, jctx)
			gmutex.Unlock()
		}

		if jctx.iFlux.influxc != nil {
			go addIDB(ocData, jctx, rtime)
		}

		if *apiControl {
			select {
			case pfor := <-jctx.pause.pch:
				l(true, jctx, fmt.Sprintf("Pausing for %v seconds\n", pfor))
				t := time.NewTimer(time.Second * time.Duration(pfor))
				select {
				case <-t.C:
					l(true, jctx, fmt.Sprintf("Done pausing for %v seconds\n", pfor))
				case <-jctx.pause.upch:
					t.Stop()
				}
			default:
			}
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
