package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	gnmi_dialout "github.com/nileshsimaria/jtimon/gnmi/dialout"
	"github.com/nileshsimaria/jtimon/gnmi/gnmi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	MAX_DIALOUT_RPCS             = 16
	SUBSCRIBER_DIALOUTSUBSCRIBER = "/Subscriber/DialOutSubscriber" // Only one such RPC per device is ensured
)

type dialoutServerT struct {
	registeredRpcs []string
	devices        map[string]*deviceInfoT
	kafkaClient    *sarama.Client
	configConsumer *sarama.Consumer
	dataProducer   *sarama.AsyncProducer
	jctx           *JCtx
}

type deviceInfoT struct {
	device     string
	rpcIdSpace int16
	rpcs       [MAX_DIALOUT_RPCS]*rpcInfoT
	jctx       *JCtx // Currently inherited from server, so use this handle as read-only

	server *dialoutServerT
}

type rpcInfoT struct {
	rpc        string
	rpcId      int8
	running    bool
	cfgChannel chan *dialOutConfigT
	config     *dialOutConfigT
	jctx       *JCtx // Currently inherited from device, so use this handle as read-only

	device *deviceInfoT
}

type dialOutConfigT struct {
	Device  string        `json:"device"`
	RpcType string        `json:"rpc-type"`
	Paths   []PathsConfig `json:"paths"`
}

func newDialOutServer(rpcs []string) *dialoutServerT {
	s := &dialoutServerT{devices: map[string]*deviceInfoT{}}

	// Create kafka Client
	kafkaCfg := sarama.NewConfig()
	kafkaClient, err := sarama.NewClient([]string{*kafkaBroker}, kafkaCfg)
	if err != nil {
		log.Fatalf("Not able to connect to Kafka broker at %v: %v", *kafkaBroker, err)
	}
	s.kafkaClient = &kafkaClient

	kafkaConsumer, err := sarama.NewConsumerFromClient(*s.kafkaClient)
	if err != nil {
		log.Fatalf("Not able to create consumer: %v", err)
	}
	s.configConsumer = &kafkaConsumer

	kafkaProducer, err := sarama.NewAsyncProducerFromClient(*s.kafkaClient)
	if err != nil {
		log.Fatalf("Not able to create producer: %v", err)
	}
	s.dataProducer = &kafkaProducer

	jctx := JCtx{
		config: Config{
			Host: "GnmiOutboundServer",
			EOS:  true,
			Log: LogConfig{
				Verbose: true,
			},
		},
	}
	logInit(&jctx)
	s.jctx = &jctx

	s.registeredRpcs = append(s.registeredRpcs, rpcs...)

	// TODO: Vivek Take config topic from cmd line
	go populateAllConfig(s, "gnmi-config")
	return s
}

func createRpc(device *deviceInfoT, name string, id int) (*rpcInfoT, error) {
	// Sarama's default channel size for consuming messages is 256
	cfgChannel := make(chan *dialOutConfigT, 512)
	return &rpcInfoT{rpc: name, rpcId: int8(id), cfgChannel: cfgChannel, jctx: device.jctx, device: device}, nil
}

func getUnusedRpcId(rpcIdSpace int16, rpcType string) (int, error) {
	ids := rpcIdSpace
	var i int
	switch rpcType {
	case SUBSCRIBER_DIALOUTSUBSCRIBER:
		i = 0 // Always first rpc
	default:
		spacelen := MAX_DIALOUT_RPCS
		i = 1
		for {
			if i >= spacelen {
				return -1, errors.New(fmt.Sprintf("RPC ID space exhausted for %s", rpcType))
			}

			bitpos := (int16(1) << i)
			if ids&bitpos == 0 {
				break
			}

			i++
		}
	}

	return i + 1, nil
}

func createOrUpdateDeviceWithNewRpc(s *dialoutServerT, cn string, rpcType string) (*deviceInfoT, *rpcInfoT, error) {
	device, ok := s.devices[cn]
	if !ok {
		device = &deviceInfoT{device: cn}
		device.jctx = s.jctx
		device.server = s
	}

	rpcId, err := getUnusedRpcId(device.rpcIdSpace, rpcType)
	if err != nil {
		jLog(s.jctx, fmt.Sprintf("[%v, DialOutSubscriber] Not able to get an unused rpc id : %v", cn, err))
		return nil, nil, err
	}
	device.rpcIdSpace |= (int16(1) << (rpcId - 1))
	log.Printf("[%v, DialOutSubscriber]rcp unused id: %v\n\n", cn, rpcId)

	rpc := device.rpcs[rpcId-1]
	if device.rpcs[rpcId-1] == nil {
		log.Printf("[%v, DialOutSubscriber]create rpc id: %v\n\n", cn, rpcId)
		rpc, err = createRpc(device, rpcType, rpcId)
		if err != nil {
			jLog(s.jctx, fmt.Sprintf("[%v, DialOutSubscriber] Not able to create rpc: %v", cn, err))
			return nil, nil, err
		}
		device.rpcs[rpc.rpcId-1] = rpc
	}

	s.devices[cn] = device
	return device, rpc, nil
}

func removeRpcFromDevice(rpc *rpcInfoT) {
	device := rpc.device

	device.rpcIdSpace = device.rpcIdSpace & (^(1 << rpc.rpcId))
	device.rpcs[rpc.rpcId] = nil
}

func (s *dialoutServerT) DialOutSubscriber(stream gnmi_dialout.Subscriber_DialOutSubscriberServer) error {
	// Get client info - UUID
	cn := *myListeningIP

	fmt.Printf("[%v, DialOutSubscriber]: Rpc begin", cn)
	if !*skipVerify {
		peer, ok := peer.FromContext(stream.Context())
		if ok && (peer.AuthInfo != nil) {
			tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
			cn = tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
		}
	} else {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if ok {
			log.Println("Client metadata:")
			log.Println(md)
		}

		if md.Len() != 0 {
			values, ok := md["server"]
			if ok {
				cn = values[0]
			}
		}
	}

	jLog(s.jctx, fmt.Sprintf("[%v, DialOutSubscriber]: Rpc begin", cn))

	_, rpc, err := createOrUpdateDeviceWithNewRpc(s, cn, SUBSCRIBER_DIALOUTSUBSCRIBER)
	if err != nil {
		jLog(s.jctx, fmt.Sprintf("[%v, DialOutSubscriber]: Not able to create or update device: %v", cn, err))
		return err
	}
	rpc.running = true
	defer func() {
		rpc.running = false
	}()

	i := 0
	length := len(rpc.cfgChannel)
	var dialOutCfg *dialOutConfigT
	// Drain n - 1 message to get latest config
	for i < (length - 1) {
		<-rpc.cfgChannel
		i++
	}
	log.Printf("Waiting for config..")
	dialOutCfg = <-rpc.cfgChannel
	log.Printf("Read config.. : %v", *dialOutCfg)

	req := &gnmi.SubscribeRequest{}
	req = &gnmi.SubscribeRequest{Request: &gnmi.SubscribeRequest_Subscribe{
		Subscribe: &gnmi.SubscriptionList{
			Mode:     gnmi.SubscriptionList_STREAM,
			Encoding: gnmi.Encoding_PROTO,
		},
	}}
	subReq := req.Request.(*gnmi.SubscribeRequest_Subscribe)
	subReq.Subscribe.Subscription, err = xPathsTognmiSubscription(dialOutCfg.Paths)
	if err != nil {
		//log.Printf("Host: %v, Invalid path config: %s", cn, cfg)
		jLog(s.jctx, fmt.Sprintf("Host: %v, Invalid path config: %s", cn, "foo...."))
		return err
	}

	// Subscribe to the device
	stream.Send(req)

	for {
		select {
		case dialOutCfg = <-rpc.cfgChannel:
			removeRpcFromDevice(rpc)
			stream.Context().Done()
		case producerError := <-(*s.dataProducer).Errors():
			log.Printf("Sending failed for %s, err: %v\n", fmt.Sprintf("%s", producerError.Msg.Value), producerError.Error())
		default:
			rspFromDevice, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			var rspString string
			rspString = fmt.Sprintf("%s", rspFromDevice)
			log.Printf("Rsp from device: %s, len: %v\n", rspString, len(rspString))
			if len(rspString) == 0 {
				continue
			}

			payload, err := proto.Marshal(rspFromDevice)
			if err != nil {
				log.Printf("Marshalling failed for %s, len: %v\n", rspString, len(rspString))
				continue
			}

			// TODO: Vivek Take data topic from cmd line
			(*s.dataProducer).Input() <- &sarama.ProducerMessage{Topic: "gnmi-data", Key: sarama.ByteEncoder(cn), Value: sarama.ByteEncoder(payload)}
		}
	}
	return err
}

func consumePartition(server *dialoutServerT, topic string, partition int32, offset int64, deviceName string) error {
	var err error
	jctx := server.jctx

	partitionConsumer, err := (*server.configConsumer).ConsumePartition(topic, partition, offset)
	if err != nil {
		errMsg := fmt.Sprintf("Not able to consume topic %v, partition %v, offset %v, err: %v", topic, partition, offset, err)
		jLog(jctx, errMsg)
		return errors.New(errMsg)
	}
	defer partitionConsumer.Close()

	var tmpDeviceName string
	for msg := range partitionConsumer.Messages() {
		log.Printf("topic: %v, partition: %v, offset: %v, msg key: %v, msg val: %v", topic, partition, offset, string(msg.Key), string(msg.Value))
		var dialOutCfg dialOutConfigT
		err = json.Unmarshal(msg.Value, &dialOutCfg)
		log.Printf("dialOutCfg: %v", dialOutCfg.Paths[0].Path)
		if err != nil {
			jLog(jctx, fmt.Sprintf("Unmarshalling dialout config failed, ignoring"))
			continue
		}

		if len(deviceName) != 0 {
			if !*skipVerify && dialOutCfg.Device != deviceName {
				continue
			}
			tmpDeviceName = deviceName
		} else {
			tmpDeviceName = dialOutCfg.Device
		}

		if len(dialOutCfg.RpcType) == 0 {
			dialOutCfg.RpcType = SUBSCRIBER_DIALOUTSUBSCRIBER
		}
		rpcName := dialOutCfg.RpcType
		switch rpcName {
		case SUBSCRIBER_DIALOUTSUBSCRIBER:
			var device *deviceInfoT
			var rpc *rpcInfoT
			var ok bool
			if device, ok = server.devices[tmpDeviceName]; !ok {
				_, rpc, err = createOrUpdateDeviceWithNewRpc(server, tmpDeviceName, SUBSCRIBER_DIALOUTSUBSCRIBER)
				if err != nil {
					errMsg := fmt.Sprintf("[%v, DialOutSubscriber]: Not able to create or update device: %v", tmpDeviceName, err)
					jLog(jctx, errMsg)
					return errors.New(errMsg)
				}
			} else {
				rpcId, _ := getUnusedRpcId(device.rpcIdSpace, SUBSCRIBER_DIALOUTSUBSCRIBER)
				rpc = device.rpcs[rpcId-1]
				if rpc == nil {
					_, rpc, err = createOrUpdateDeviceWithNewRpc(server, tmpDeviceName, SUBSCRIBER_DIALOUTSUBSCRIBER)
					if err != nil {
						errMsg := fmt.Sprintf("[%v, DialOutSubscriber]: Not able to create or update device: %v", tmpDeviceName, err)
						jLog(jctx, errMsg)
						return errors.New(errMsg)
					}
				}
			}
			fmt.Printf("Writing to %v's channel, device: %v, , dialOutCfg: %v", *rpc, *rpc.device, dialOutCfg)
			rpc.cfgChannel <- &dialOutCfg
		default:
			errMsg := fmt.Sprintf("[%v, DialOutSubscriber]: Unimplemented rpc %v, ignoring", tmpDeviceName, err)
			jLog(jctx, errMsg)
		}
	}

	return nil
}

func populateAllConfig(server *dialoutServerT, topic string) {
	partitions := []int32{}
	var err error
	jctx := server.jctx

	for {
		partitions, err = (*server.configConsumer).Partitions(topic)
		if err != nil {
			jLog(jctx, fmt.Sprintf("Not able to fetch partitions: %v", err))
			time.Sleep(2 * time.Second)
			continue
		}
		for i, p := range partitions {
			offset, err := (*server.kafkaClient).GetOffset(topic, p, sarama.OffsetOldest)
			if err != nil {
				jLog(jctx, fmt.Sprintf("Not able to fetch offset for topic %v, partition %v", topic, p))
				continue
			}

			if i == len(partitions)-1 {
				consumePartition(server, topic, p, offset, "")
			} else {
				go consumePartition(server, topic, p, offset, "")
			}
		}
	}
}

func startDialOutServer(host *string, port *int) {
	var transportCreds credentials.TransportCredentials

	lis, err := net.Listen("tcp", *host+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("Failed to create listener for %v", *host+":"+strconv.Itoa(*port))
	}

	// TODO: Vivek - Talk to PAPI for certs
	if !*skipVerify {
		if *myCert == "" {
			log.Fatalf("Cert not provided")
		}
		myCertFile, err := filepath.Abs(*myCert)
		if err != nil {
			log.Fatalf("Cert %v not found", *myCert)
		}

		if *myKey == "" {
			log.Fatalf("Key not provided")
		}
		myKeyFile, err := filepath.Abs(*myKey)
		if err != nil {
			log.Fatalf("Key %v not found", *myKey)
		}

		cert, err := tls.LoadX509KeyPair(myCertFile, myKeyFile)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Error loading certificate: %s", err))
		}

		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(*myCACert)
		if err != nil {
			log.Fatalf("Failed to read ca cert: %s", err)
		}

		if ok := certPool.AppendCertsFromPEM(bs); !ok {
			log.Fatalf("Failed to append certs")
		}

		transportCreds = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		})
	}

	grpcServer := grpc.NewServer(grpc.Creds(transportCreds))

	dialOutServer := newDialOutServer([]string{SUBSCRIBER_DIALOUTSUBSCRIBER})
	gnmi_dialout.RegisterSubscriberServer(grpcServer, dialOutServer)
	grpcServer.Serve(lis)
}
