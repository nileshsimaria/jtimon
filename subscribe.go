package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"encoding/json"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func handleOnePacket(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	updateStats(jctx, ocData, true)

	s := ""

	if *print || (IsVerboseLogging(jctx) && !*print) {
		s += fmt.Sprintf("system_id: %s\n", ocData.SystemId)
		s += fmt.Sprintf("component_id: %d\n", ocData.ComponentId)
		s += fmt.Sprintf("sub_component_id: %d\n", ocData.SubComponentId)
		s += fmt.Sprintf("path: %s\n", ocData.Path)
		s += fmt.Sprintf("sequence_number: %d\n", ocData.SequenceNumber)
		s += fmt.Sprintf("timestamp: %d\n", ocData.Timestamp)
		s += fmt.Sprintf("sync_response: %v\n", ocData.SyncResponse)
		if ocData.SyncResponse {
			s += "Received sync_response\n"
		}

		del := ocData.GetDelete()
		for _, d := range del {
			s += fmt.Sprintf("Delete: %s\n", d.GetPath())
		}
	}

	prefixSeen := false
	for _, kv := range ocData.Kv {
		updateStatsKV(jctx, true)

		if *print || (IsVerboseLogging(jctx) && !*print) {
			s += fmt.Sprintf("  key: %s\n", kv.Key)
			switch value := kv.Value.(type) {
			case *na_pb.KeyValue_DoubleValue:
				s += fmt.Sprintf("  double_value: %v\n", value.DoubleValue)
			case *na_pb.KeyValue_IntValue:
				s += fmt.Sprintf("  int_value: %d\n", value.IntValue)
			case *na_pb.KeyValue_UintValue:
				s += fmt.Sprintf("  uint_value: %d\n", value.UintValue)
			case *na_pb.KeyValue_SintValue:
				s += fmt.Sprintf("  sint_value: %d\n", value.SintValue)
			case *na_pb.KeyValue_BoolValue:
				s += fmt.Sprintf("  bool_value: %v\n", value.BoolValue)
			case *na_pb.KeyValue_StrValue:
				s += fmt.Sprintf("  str_value: %s\n", value.StrValue)
			case *na_pb.KeyValue_BytesValue:
				s += fmt.Sprintf("  bytes_value: %s\n", value.BytesValue)
			default:
				s += fmt.Sprintf("  default: %v\n", value)
			}
		}

		if kv.Key == "__prefix__" {
			prefixSeen = true
		} else if !strings.HasPrefix(kv.Key, "__") {
			if !prefixSeen && !strings.HasPrefix(kv.Key, "/") {
				if *prefixCheck {
					s += fmt.Sprintf("Missing prefix for sensor: %s\n", ocData.Path)
				}
			}
		}
	}
	if s != "" {
		jLog(jctx, s)
	}
}

func subSendAndReceive(conn *grpc.ClientConn, jctx *JCtx, subReqM na_pb.SubscriptionRequest) {
	var ctx context.Context
	c := na_pb.NewOpenConfigTelemetryClient(conn)
	if jctx.config.Meta == true {
		md := metadata.New(map[string]string{"username": jctx.config.User, "password": jctx.config.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctx = context.Background()
	}

	stream, err := c.TelemetrySubscribe(ctx, &subReqM)

	if err != nil {
		jLog(jctx, fmt.Sprintf("Could not send RPC: %v\n", err))
		return
	}

	hdr, errh := stream.Header()
	if errh != nil {
		jLog(jctx, fmt.Sprintf("Failed to get header for stream: %v", errh))
	}

	jLog(jctx, fmt.Sprintf("gRPC headers from host %s:%d\n", jctx.config.Host, jctx.config.Port))
	for k, v := range hdr {
		jLog(jctx, fmt.Sprintf("  %s: %s\n", k, v))
	}

	if jctx.config.Log.CSVStats {
		jLog(jctx, fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			"sensor-path", "sequence-number", "component-id", "sub-component-id", "packet-size", "p-ts", "e-ts", "re-stream-creation-ts", "re-payload-get-ts"))
	}

	jLog(jctx, fmt.Sprintf("Receiving telemetry data from %s:%d\n", jctx.config.Host, jctx.config.Port))
	for {
		ocData, err := stream.Recv()
		if err == io.EOF {
			printSummary(jctx)
			break
		}
		if err != nil {
			jLog(jctx, fmt.Sprintf("%v.TelemetrySubscribe(_) = _, %v", conn, err))
			return
		}

		rtime := time.Now()

		if jctx.config.Log.DropCheck && !jctx.config.Log.CSVStats {
			dropCheck(jctx, ocData)
		}

		if *outJSON {
			if b, err := json.MarshalIndent(ocData, "", "  "); err == nil {
				jLog(jctx, fmt.Sprintf("%s\n", b))
			}
		}

		if *print || *stateHandler || IsVerboseLogging(jctx) {
			handleOnePacket(ocData, jctx)
		}

		if jctx.influxCtx.influxClient != nil {
			go addIDB(ocData, jctx, rtime)
		}

		if *apiControl {
			select {
			case pfor := <-jctx.pause.pch:
				jLog(jctx, fmt.Sprintf("Pausing for %v seconds\n", pfor))
				t := time.NewTimer(time.Second * time.Duration(pfor))
				select {
				case <-t.C:
					jLog(jctx, fmt.Sprintf("Done pausing for %v seconds\n", pfor))
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
	cfg := jctx.config

	for i := range cfg.Paths {
		var pathM na_pb.Path
		pathM.Path = cfg.Paths[i].Path
		pathM.SampleFrequency = uint32(cfg.Paths[i].Freq)

		subReqM.PathList = append(subReqM.PathList, &pathM)
	}
	additionalConfigM.NeedEos = jctx.config.EOS
	subReqM.AdditionalConfig = &additionalConfigM
	subSendAndReceive(conn, jctx, subReqM)
}
