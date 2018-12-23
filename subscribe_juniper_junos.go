package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"encoding/json"

	"github.com/golang/protobuf/proto"
	auth_pb "github.com/nileshsimaria/jtimon/authentication"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SubErrorCode to define the type of errors
type SubErrorCode int

// Error Codes for Subscribe Send routines
const (
	SubRcConnRetry = iota
	SubRcSighupRestart
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

// subSendAndReceive handles the following
// 		- Opens up a stream for receiving the telemetry data
//		- Handles SIGHUP by terminating the current stream and requests the
//		  	caller to restart the streaming by setting the corresponding return
//			code
//		- In case of an error, Set the error code to restart the connection.
func subSendAndReceive(conn *grpc.ClientConn, jctx *JCtx,
	subReqM na_pb.SubscriptionRequest,
	statusch chan<- bool) SubErrorCode {

	var ctx context.Context
	c := na_pb.NewOpenConfigTelemetryClient(conn)
	if jctx.config.Meta {
		md := metadata.New(map[string]string{"username": jctx.config.User, "password": jctx.config.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctx = context.Background()
	}
	stream, err := c.TelemetrySubscribe(ctx, &subReqM)

	if err != nil {
		return SubRcConnRetry
	}

	hdr, errh := stream.Header()
	if errh != nil {
		jLog(jctx, fmt.Sprintf("Failed to get header for stream: %v", errh))
	}

	jLog(jctx, fmt.Sprintf("gRPC headers from host %s:%d\n", jctx.config.Host, jctx.config.Port))
	for k, v := range hdr {
		jLog(jctx, fmt.Sprintf("  %s: %s", k, v))
	}

	if jctx.config.Log.CSVStats {
		jLog(jctx, fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			"sensor-path", "sequence-number", "component-id", "sub-component-id", "packet-size", "p-ts", "e-ts", "re-stream-creation-ts", "re-payload-get-ts"))
	}

	datach := make(chan struct{})

	// inform the caller that streaming has been started
	statusch <- true
	go func() {
		// Go Routine which actually starts the streaming connection and receives the data
		jLog(jctx, fmt.Sprintf("Receiving telemetry data from %s:%d\n", jctx.config.Host, jctx.config.Port))

		for {
			ocData, err := stream.Recv()
			if err == io.EOF {
				printSummary(jctx)
				datach <- struct{}{}
				return
			}
			if err != nil {
				jLog(jctx, fmt.Sprintf("%v.TelemetrySubscribe(_) = _, %v", conn, err))
				datach <- struct{}{}
				return
			}

			if *genTestData {
				if ocDataM, err := proto.Marshal(ocData); err == nil {
					generateTestData(jctx, ocDataM)
				} else {
					jLog(jctx, fmt.Sprintf("%v", err))
				}
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

			if *noppgoroutines {
				addIDB(ocData, jctx, rtime)
			} else {
				go addIDB(ocData, jctx, rtime)
			}

			if *prom {
				go addPrometheus(ocData, jctx)
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
	}()
	for {
		// The below select loop will handle the following
		// 		1. Tell the caller that streaming has started
		//		2. Tell the caller that there is no incoming data
		select {
		case <-jctx.pause.subch:
			// config has been updated restart the streaming. this will make
			// subscribe call to return and after that it will retry with
			// newly parsed config
			return SubRcSighupRestart
		case <-datach:
			// data is not received, retry the connection
			return SubRcConnRetry
		}
	}
}

// subscribe routine constructs the subscription paths and calls
// the function to start the streaming connection.
//
// In case of SIGHUP, the paths are formed again and streaming
// is restarted.
func subscribeJunos(conn *grpc.ClientConn, jctx *JCtx, statusch chan<- bool) SubErrorCode {
	var subReqM na_pb.SubscriptionRequest
	var additionalConfigM na_pb.SubscriptionAdditionalConfig

	cfg := &jctx.config
	for i := range cfg.Paths {
		var pathM na_pb.Path
		pathM.Path = cfg.Paths[i].Path
		pathM.SampleFrequency = uint32(cfg.Paths[i].Freq)
		subReqM.PathList = append(subReqM.PathList, &pathM)
	}
	additionalConfigM.NeedEos = jctx.config.EOS
	subReqM.AdditionalConfig = &additionalConfigM

	res := subSendAndReceive(conn, jctx, subReqM, statusch)
	if res == SubRcSighupRestart {
		jLog(jctx, fmt.Sprintf("Restarting the connection for sighup\n"))
	} else {
		jLog(jctx, fmt.Sprintf("Restarting the connection"))
	}
	return res
}

func loginCheckJunos(jctx *JCtx, conn *grpc.ClientConn) error {
	if jctx.config.User != "" && jctx.config.Password != "" {
		user := jctx.config.User
		pass := jctx.config.Password
		if !jctx.config.Meta {
			lc := auth_pb.NewLoginClient(conn)
			dat, err := lc.LoginCheck(context.Background(),
				&auth_pb.LoginRequest{UserName: user,
					Password: pass, ClientId: jctx.config.CID})
			if err != nil {
				return fmt.Errorf("[%s] Could not login: %v", jctx.config.Host, err)
			}
			if !dat.Result {
				return fmt.Errorf("[%s] LoginCheck failed", jctx.config.Host)
			}
		}
	}
	return nil
}
