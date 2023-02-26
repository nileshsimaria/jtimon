package jtisim

import (
	"fmt"
	"log"
	"net"
	"strings"

	apb "github.com/nileshsimaria/jtimon/authentication"
	gnmipb "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	tpb "github.com/nileshsimaria/jtimon/telemetry"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	// server size compression
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

// JTISim is JTI Simulator
type JTISim struct {
	host    string
	port    int32
	random  bool
	descDir string
}

// NewJTISim to create new jti simulator
func NewJTISim(host string, port int32, random bool, descDir string) *JTISim {
	return &JTISim{
		host:    host,
		port:    port,
		random:  random,
		descDir: descDir,
	}
}

// Start the simulator
func (s *JTISim) Start() error {
	if lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port)); err == nil {
		grpcServer := grpc.NewServer()
		authServer := &authServer{}

		apb.RegisterLoginServer(grpcServer, authServer)
		tpb.RegisterOpenConfigTelemetryServer(grpcServer, &server{s})
		gnmipb.RegisterGNMIServer(grpcServer, &server{s})

		grpcServer.Serve(lis)
	} else {
		return err
	}
	return nil
}

type server struct {
	jtisim *JTISim
}
type authServer struct {
}

func (s *authServer) LoginCheck(ctx context.Context, req *apb.LoginRequest) (*apb.LoginReply, error) {
	// allow everyone
	rep := &apb.LoginReply{
		Result: true,
	}
	return rep, nil
}

func (s *server) TelemetrySubscribe(req *tpb.SubscriptionRequest, stream tpb.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Println("Client metadata:")
		log.Println(md)
	}

	// send metadata to client
	header := metadata.Pairs("jtisim", "yes")
	stream.SendHeader(header)

	plist := req.GetPathList()
	ch := make(chan *tpb.OpenConfigData)
	for _, path := range plist {
		pname := path.GetPath()
		switch {
		case strings.HasPrefix(pname, "/interfaces"):
			go s.streamInterfaces(ch, path)
		case strings.HasPrefix(pname, "/bgp"):
			go s.streamBGP(ch, path)
		case strings.HasPrefix(pname, "/lldp"):
			go s.streamLLDP(ch, path)
		default:
			log.Fatalf("Sensor (%s) is not yet supported", pname)
		}
	}

	for {
		select {
		case data := <-ch:
			if err := stream.Send(data); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func (s *server) CancelTelemetrySubscription(ctx context.Context, req *tpb.CancelSubscriptionRequest) (*tpb.CancelSubscriptionReply, error) {
	return nil, nil
}

func (s *server) GetTelemetrySubscriptions(ctx context.Context, req *tpb.GetSubscriptionsRequest) (*tpb.GetSubscriptionsReply, error) {
	return nil, nil
}

func (s *server) GetTelemetryOperationalState(ctx context.Context, req *tpb.GetOperationalStateRequest) (*tpb.GetOperationalStateReply, error) {
	return nil, nil
}

func (s *server) GetDataEncodings(ctx context.Context, req *tpb.DataEncodingRequest) (*tpb.DataEncodingReply, error) {
	return nil, nil
}

func (s *server) Capabilities(context.Context, *gnmipb.CapabilityRequest) (*gnmipb.CapabilityResponse, error) {
	return nil, nil
}
func (s *server) Get(context.Context, *gnmipb.GetRequest) (*gnmipb.GetResponse, error) {
	return nil, nil
}
func (s *server) Set(context.Context, *gnmipb.SetRequest) (*gnmipb.SetResponse, error) {
	return nil, nil
}
func (s *server) Subscribe(stream gnmipb.GNMI_SubscribeServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Println("Client metadata:")
		log.Println(md)
	}

	// send metadata to client
	header := metadata.Pairs("jtisim", "yes")
	stream.SendHeader(header)

	req, err := stream.Recv()
	if err != nil {
		log.Fatalf("Recv failed: %v", err)
	}

	subReq := req.GetSubscribe()
	if subReq == nil {
		log.Fatalf("Invalid subscribe request, received %v", req.GetRequest())
	}

	if subReq.GetEncoding() != gnmipb.Encoding_PROTO {
		log.Fatalf("Only PROTO encoding supported, received %v", subReq.GetEncoding())
	}

	if subReq.GetMode() != gnmipb.SubscriptionList_STREAM {
		log.Fatalf("Only STREAM mode supported, received %v", subReq.GetMode())
	}

	if subReq.GetUseAliases() {
		log.Fatalf("Aliases not supported, received %v", subReq.GetUseAliases())
	}

	stream.Send(&gnmipb.SubscribeResponse{Response: &gnmipb.SubscribeResponse_SyncResponse{SyncResponse: true}})
	ch := make(chan *gnmipb.SubscribeResponse)
	for _, sub := range subReq.GetSubscription() {
		sub.GetSampleInterval()
		gnmiPath := sub.GetPath()
		pname, _, _ := gnmiParsePath("", gnmiPath.GetElem(), nil, nil)
		switch {
		case strings.HasPrefix(pname, "/interfaces"):
			go s.gnmiStreamInterfaces(ch, pname, sub)
		default:
			log.Fatalf("Sensor (%s) is not yet supported", pname)
		}
	}

	for {
		select {
		case data := <-ch:
			if err := stream.Send(data); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}
