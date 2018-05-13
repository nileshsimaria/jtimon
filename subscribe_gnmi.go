package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	gnmipb "github.com/nileshsimaria/jtimon/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func getEncoding(encoding string) gnmipb.Encoding {
	switch strings.ToLower(encoding) {
	case "proto":
		return gnmipb.Encoding_PROTO
	case "json":
		return gnmipb.Encoding_JSON
	case "bytes":
		return gnmipb.Encoding_BYTES
	case "ascii":
		return gnmipb.Encoding_ASCII
	case "json-ietf":
		return gnmipb.Encoding_JSON_IETF
	default:
		log.Fatalf("unsupported encoding: please use proto|json|bytes|ascii|json-ietf\n")
	}
	return gnmipb.Encoding_PROTO
}

func getSMode(mode string) gnmipb.SubscriptionMode {
	switch strings.ToLower(mode) {
	case "target-defined":
		return gnmipb.SubscriptionMode_TARGET_DEFINED
	case "sample":
		return gnmipb.SubscriptionMode_SAMPLE
	case "on-change":
		return gnmipb.SubscriptionMode_ON_CHANGE
	default:
		log.Fatalf("unsupported subscription mode\n")
	}
	return gnmipb.SubscriptionMode_SAMPLE
}

func getMode(mode string) gnmipb.SubscriptionList_Mode {
	switch strings.ToLower(mode) {
	case "once":
		return gnmipb.SubscriptionList_ONCE
	case "stream":
		return gnmipb.SubscriptionList_STREAM
	case "poll":
		return gnmipb.SubscriptionList_POLL
	default:
		log.Fatalf("unsupported mode: please use stream | once | poll \n")
	}
	return gnmipb.SubscriptionList_STREAM
}

func subscribeGNMI(conn *grpc.ClientConn, jctx *JCtx) {
	cfg := jctx.config

	s := &gnmipb.SubscribeRequest_Subscribe{
		Subscribe: &gnmipb.SubscriptionList{
			Mode:     getMode(*gnmiMode),
			Prefix:   &gnmipb.Path{Target: ""},
			Encoding: getEncoding(*gnmiEncoding),
		},
	}

	for i := range cfg.Paths {
		fmt.Printf("ocPath: %s\n", cfg.Paths[i].Path)
		gpath, err := xpathToGNMIPath(cfg.Paths[i].Path)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fmt.Printf("gnmiPath: (%T) %q\n", gpath, gpath)
		fmt.Printf("pathToString: %q\n", pathToString(gpath))
		pp, err := StringToPath(pathToString(gpath), StructuredPath, StringSlicePath)
		if err != nil {
			log.Fatalf("Invalid path: %v", err)
		}
		s.Subscribe.Subscription = append(s.Subscribe.Subscription,
			&gnmipb.Subscription{
				Path:           pp,
				Mode:           getSMode(cfg.Paths[i].Mode),
				SampleInterval: cfg.Paths[i].Freq,
			})
		fmt.Printf("gNMIPath: (%T) %q\n", pp, pp)
	}

	req := &gnmipb.SubscribeRequest{Request: s}
	fmt.Printf("\n\n%q\n\n", req)
	subSendAndReceiveGNMI(conn, jctx, req)
}

func processGNMIResponse(resp *gnmipb.SubscribeResponse) {
	if notif := resp.GetUpdate(); notif != nil {
		//fmt.Printf("update: %q\n", notif)
	}
	if syncResp := resp.GetSyncResponse(); syncResp {
		fmt.Printf("Received sync-response\n")
		if false {
			os.Exit(0)
		}
	}
	if err := resp.GetError(); err != nil {
		fmt.Printf("Received error: %q\n", err)
	}
}

func subSendAndReceiveGNMI(conn *grpc.ClientConn, jctx *JCtx, req *gnmipb.SubscribeRequest) {
	var ctx context.Context
	c := gnmipb.NewGNMIClient(conn)

	if jctx.config.Meta == true {
		md := metadata.New(map[string]string{"username": jctx.config.User, "password": jctx.config.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctx = context.Background()
	}

	client, err := c.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Error invoking gnmi.subscribe(): %q", err)
	}
	if err := client.Send(req); err != nil {
		log.Fatalf("Error sending(): %q", err)
	}

	for {
		var resp *gnmipb.SubscribeResponse
		resp, err := client.Recv()
		if err != nil {
			log.Fatalf("Recv error: %s\n", err)
		}
		processGNMIResponse(resp)
	}
}

func xpathToGNMIPath(input string) ([]string, error) {
	path := strings.Trim(input, "/")
	var buf []rune
	inKey := false
	null := rune(0)
	for _, r := range path {
		switch r {
		case '[':
			if inKey {
				return nil, fmt.Errorf("malformed path, nested '[': %q ", path)
			}
			inKey = true
		case ']':
			if !inKey {
				return nil, fmt.Errorf("malformed path, unmatched ']': %q", path)
			}
			inKey = false
		case '/':
			if !inKey {
				buf = append(buf, null)
				continue
			}
		}
		buf = append(buf, r)
	}
	if inKey {
		return nil, fmt.Errorf("malformed path, missing trailing ']': %q", path)
	}
	return strings.Split(string(buf), string(null)), nil
}

func pathToString(q []string) string {
	qq := make([]string, len(q))
	copy(qq, q)
	for i, e := range qq {
		qq[i] = strings.Replace(e, "/", "\\/", -1)
	}
	return strings.Join(qq, "/")
}
