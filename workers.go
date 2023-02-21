package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/golang/protobuf/proto"
	"github.com/nileshsimaria/jtimon/dialout"
	"github.com/nileshsimaria/jtimon/gnmi/gnmi"
	tpb "github.com/openconfig/grpctunnel/proto/tunnel"
	"github.com/openconfig/grpctunnel/tunnel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JCtx is JTIMON run time context
type JCtx struct {
	config          Config
	file            string
	wg              *sync.WaitGroup
	influxCtx       InfluxCtx
	esCtx           ElasticCtx
	stats           statsCtx
	pExporter       *jtimonPExporter
	control         chan os.Signal
	running         bool
	alias           *Alias
	testMeta        *os.File
	testBytes       *os.File
	testExp         *os.File
	testRes         *os.File
	receivedSyncRsp bool
}

// JWorkers holds worker
type JWorkers struct {
	m          map[string]*JWorker
	devices    map[string]*JWorker
	wg         sync.WaitGroup
	mr         int64
	files      []string
	fileList   string
	sigchan    chan os.Signal
	statusChan chan string
}

// NewJWorkers to create new workers
func NewJWorkers(files []string, fileList string, mr int64) *JWorkers {
	return &JWorkers{
		m:          make(map[string]*JWorker),
		mr:         mr,
		files:      files,
		fileList:   fileList,
		sigchan:    make(chan os.Signal, 10),
		statusChan: make(chan string, 1),
	}
}

// EndWorkers by sending them SIGINT
func (ws *JWorkers) EndWorkers() {
	ws.sigchan <- os.Interrupt
}

// SIGHUPWorkers for reconfig
func (ws *JWorkers) SIGHUPWorkers() {
	ws.sigchan <- syscall.SIGHUP
}

// StartWorkers to start all of the workers.
// - add all of workers
// - ask workers to start actual work i.e. send syscall.SIGCONT
// - start signal handler and max run handler go routines
func (ws *JWorkers) StartWorkers() {
	ws.AddWorkers(ws.files)
	if !*dialOut {
		for _, v := range ws.m {
			v.signalch <- syscall.SIGCONT
		}
	}
	go ws.signalHandler(ws.fileList)
	go ws.maxRunHandler(ws.mr)
}

// Wait for all the workers to finish
func (ws *JWorkers) Wait() {
	ws.wg.Wait()
}

// AddWorkers to add all the workers
func (ws *JWorkers) AddWorkers(files []string) {
	for _, file := range files {
		ws.AddWorker(file)
	}
}

// StartWorker is to start one worker
// - add worker
// - ask worker to start actual work i.e. send syscall.SIGCONT
func (ws *JWorkers) StartWorker(file string) {
	ws.AddWorker(file)
	for k, v := range ws.m {
		if k == file {
			v.signalch <- syscall.SIGCONT
			return
		}
	}
}

// AddWorker is to add new worker in set of (actually map of) workers
func (ws *JWorkers) AddWorker(file string) {
	if w, err := NewJWorker(file, &ws.wg, ws.statusChan); err == nil {
		ws.m[file] = w
		ws.wg.Add(1)
	}
}

func (ws *JWorkers) maxRunHandler(maxRunTime int64) {
	if maxRunTime > 0 {
		// mr - Max run time in seconds
		// Subscription is configured for a certain time period
		// Once the time expires, interrupt worker go routines.
		tickChan := time.NewTimer(time.Second * time.Duration(maxRunTime)).C
		<-tickChan
		for _, w := range ws.m {
			w.signalch <- os.Interrupt
		}
	}
}

func (ws *JWorkers) handleConfigChanges() {
	// we support config changes through sighup only for config file list
	// perform following on sighup:
	// 	  Add new worker if needed
	//	  delete worker if not in new list
	//    otherwise, send sighup to worker to restart streaming with new config
	if configfilelist, err := NewJTIMONConfigFilelist(ws.fileList); err == nil {
		for _, file := range configfilelist.Filenames {
			if w, ok := ws.m[file]; ok {
				// signal to the worker if they are running. upon receiving sighup,
				// the worker'd stop current streaming, parse new config and make new
				// connection (grpc dial) to the device to get new streams of data
				log.Printf("sending sighup to the worker for %v", file)
				w.signalch <- syscall.SIGHUP
			} else {
				// new worker
				log.Printf("adding a new worker for %v", file)
				ws.StartWorker(file)
			}
		}
		// handle deletions
		for file, w := range ws.m {
			if StringInSlice(file, configfilelist.Filenames) == false {
				// kill the worker go routine and remove it from the map
				log.Printf("deleting worker for %v", file)
				w.signalch <- os.Interrupt
				delete(ws.m, file)
			}
		}
	} else {
		log.Printf("error in parsing the new config file, continuing with older config")
	}
}

func (ws *JWorkers) signalHandler(configFileList string) {
	sigchan := make(chan os.Signal, 10)
	ws.sigchan = sigchan
	// handle interrupt and sighup
	signal.Notify(sigchan, os.Interrupt, syscall.SIGHUP)
	for {
		select {
		case s := <-sigchan:
			switch s {
			case syscall.SIGHUP:
				// propagate the signal to workers and continue waiting for signals
				if len(ws.fileList) != 0 {
					ws.handleConfigChanges()
				} else {
					// SIGHUP the workers spawned when --config option is used
					for _, w := range ws.m {
						w.signalch <- s
					}
				}
			case os.Interrupt:
				for _, w := range ws.m {
					w.signalch <- s
				}
				return
			}
		case file := <-ws.statusChan:
			// worker must have encountered error
			log.Printf("Worker removed for %v", file)
			delete(ws.m, file)
		}
	}
}

// JWorker is one worker
type JWorker struct {
	jctx     *JCtx
	signalch chan os.Signal
}

func getDialOutRequest(jctx *JCtx) *dialout.DialOutRequest {
	// Form the context for dialout
	dialOutCfg := jctx.config
	influxServer, ok := os.LookupEnv("MY_NAME")
	if dialOutCfg.Influx.Server != "" {
		if !ok {
			influxServer = jctx.config.Influx.Server
		}
		dialOutCfg.Influx.Server = influxServer
		if dialOutCfg.Influx.Dbname == "" {
			dialOutCfg.Influx.Dbname = jctx.config.Host
		}
	}

	esServer, ok := os.LookupEnv("MY_ESNAME")
	if dialOutCfg.Es.Server != "" {
		if !ok {
			esServer = jctx.config.Es.Server
		}
		dialOutCfg.Es.Server = esServer
		if dialOutCfg.Es.Idxname == "" {
			dialOutCfg.Es.Idxname = jctx.config.Host
		}

	}
	cfgPayload, err := json.Marshal(&dialOutCfg)
	if err != nil {
		// TODO: Vivek Will  jctx.config.Host reflect the right host ?
		jLog(jctx, fmt.Sprintf("Marshalling configuration failed for %v, err: %v", jctx.config.Host, err))
	}

	// Form the dialout request
	var req dialout.DialOutRequest
	req.Device = dialOutCfg.Host
	for _, tmp := range dialOutCfg.Paths {
		var path dialout.PathsConfig
		path.Path = tmp.Path
		path.Frequency = tmp.Freq
		path.Mode = tmp.Mode

		req.Paths = append(req.Paths, &path)
	}
	req.DialOutContext = cfgPayload
	return &req
}

// NewJWorker is to create new worker
func NewJWorker(file string, wg *sync.WaitGroup, wsChan chan string) (*JWorker, error) {
	w := &JWorker{}

	signalch := make(chan os.Signal)
	statusch := make(chan struct{})
	dataCh := make(chan *gnmi.SubscribeResponse, 20000)
	jctx := JCtx{
		file:      file,
		wg:        wg,
		pExporter: exporter,
		stats: statsCtx{
			startTime: time.Now(),
		},
	}
	w.jctx = &jctx
	w.signalch = signalch

	if *genTestData {
		testSetup(&jctx)
	}

	err := ConfigRead(&jctx, true, nil)
	if err != nil {
		log.Println(err)
		return w, err
	}
	log.Printf("%v, jctx.config.Kafka.producer: %v", jctx.config.Host, jctx.config.Kafka)
	if alias, err := NewAlias(jctx.config.Alias); err == nil {
		jctx.alias = alias
        } else {
		jLog(&jctx, fmt.Sprintf("New alias creation failed for %v, err: %v", jctx.config.Host, err))
        }
	go func() {
		if *dialOut {
			// Publish the initial config for dial-out
			req := getDialOutRequest(&jctx)
			ctxPayload, err := proto.Marshal(req)
			if err != nil {
				// TODO: Vivek Will  jctx.config.Host reflect the right host ?
				jLog(&jctx, fmt.Sprintf("Marshalling dial-out context failed for %v, err: %v", jctx.config.Host, err))
			}
			// TODO: Vivek Make topic configurable.
			topic := "gnmi-config"
			p, o, err := (*jctx.config.Kafka.producer).SendMessage(
				&sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.ByteEncoder(ctxPayload),
				},
			)
			jLog(&jctx, fmt.Sprintf("Initial configuration for %v published to topic %v, partition %v, offset %v, err: %v", jctx.config.Host, topic, p, o, err))
		}
		for {
			select {
			case sig := <-signalch:
				switch sig {
				case os.Interrupt:
					// we are asked to stop
					printSummary(&jctx)
					jLog(&jctx, fmt.Sprintf("Streaming for host %s will be stopped (SIGINT)", jctx.config.Host))
					if *genTestData {
						testTearDown(&jctx)
					}
					jctx.wg.Done()
					// let the downstream subscribe go routines know we are done and no need to restart
					jctx.control <- os.Interrupt
					logStop(&jctx)
					return
				case syscall.SIGHUP:
					// handle SIGHUP if the streaming is happening.
					// running will not be set when the connection is
					// not establihsed and it is trying to connect.
					// ConfigRead will re-parse the config and updates jctx so
					// when we retry Dial, it will do it with updated config
					restart := false
					err := ConfigRead(&jctx, false, &restart)
					if err != nil {
						jLog(&jctx, fmt.Sprintln(err))
					}
					if !*dialOut {
						if restart {
							jctx.control <- syscall.SIGHUP
						} else {
							jLog(&jctx, fmt.Sprintf("config re-parse, data streaming has not started yet"))
						}
					} else {
						req := getDialOutRequest(&jctx)
						ctxPayload, err := proto.Marshal(req)
						if err != nil {
							// TODO: Vivek Will  jctx.config.Host reflect the right host ?
							jLog(&jctx, fmt.Sprintf("Marshalling dial-out context failed for %v, err: %v", jctx.config.Host, err))
						}
						// TODO: Vivek Make topic configurable.
						topic := "gnmi-config"
						p, o, err := (*jctx.config.Kafka.producer).SendMessage(
							&sarama.ProducerMessage{
								Topic: topic,
								Value: sarama.ByteEncoder(ctxPayload),
							},
						)
						jLog(&jctx, fmt.Sprintf("Configuration for %v published to topic %v, partition %v, offset %v, err: %v", jctx.config.Host, topic, p, o, err))
					}
				case syscall.SIGCONT:
					// Do not start the subscribe worker for dialout..
					if !*dialOut {
						go work(&jctx, statusch)
					}
				}
			case <-statusch:
				// worker must have encountered error
				printSummary(&jctx)
				wsChan <- jctx.file
				jctx.wg.Done()
				logStop(&jctx)
				return
			case rsp := <-dataCh:
				err := gnmiHandleResponse(&jctx, rsp)
				if err != nil && strings.Contains(err.Error(), gGnmiJtimonIgnoreErrorSubstr) {
					jLog(&jctx, fmt.Sprintf("gnmiHandleResponse failed: %v", err))
					continue
				}
			}
		}
	}()
	return w, nil
}

// Conn is a wraper as a net.Conn interface.
type Conn struct {
	io.ReadWriteCloser
}

// LocalAddr is trivial implementation, in order to match interface net.Conn.
func (tc *Conn) LocalAddr() net.Addr { return nil }

// RemoteAddr is trivial implementation, in order to match interface net.Conn.
func (tc *Conn) RemoteAddr() net.Addr { return nil }

// SetDeadline is trivial implementation, in order to match interface net.Conn.
func (tc *Conn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline is trivial implementation, in order to match interface net.Conn.
func (tc *Conn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline is trivial implementation, in order to match interface net.Conn.
func (tc *Conn) SetWriteDeadline(t time.Time) error { return nil }

const (
	retryBaseDelay     = time.Second
	retryMaxDelay      = time.Minute
	retryRandomization = 0.5
)

func getBackOff() *backoff.ExponentialBackOff {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0 // Retry Subscribe indefinitely.
	bo.InitialInterval = retryBaseDelay
	bo.MaxInterval = retryMaxDelay
	bo.RandomizationFactor = retryRandomization
	return bo
}

func workTunnel(jctx *JCtx, statusch chan struct{}) error {
	certFile := jctx.config.GRPC.TunnelServer.TLS.ClientCrt
	keyFile := jctx.config.GRPC.TunnelServer.TLS.ClientKey
	caFile := jctx.config.GRPC.TunnelServer.TLS.CA
	insecure := jctx.config.GRPC.TunnelServer.TLS.Insecure

	tunnelAddress := jctx.config.GRPC.TunnelServer.Address
	dialTargetCfg := jctx.config.GRPC.TunnelServer.DialTarget
	dialTargetType := jctx.config.GRPC.TunnelServer.DialTargetType

	var opts []grpc.DialOption
	var err error

	if insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		if len(certFile) == 0 || len(keyFile) == 0 {
			opts, err = tunnel.DialTLSCredsOpts(caFile)
		} else {
			opts, err = tunnel.DialmTLSCredsOpts(certFile, keyFile, caFile)
		}
	}

	if err != nil {
		return err
	}
	clientConn, err := grpc.Dial(tunnelAddress, opts...)
	if err != nil {
		return fmt.Errorf("grpc dial error: %v", err)
	}
	defer clientConn.Close()

	peers := make(map[tunnel.Target]struct{})
	var peerMux sync.Mutex

	peerAddHandler := func(t tunnel.Target) error {
		peerMux.Lock()
		defer peerMux.Unlock()
		peers[t] = struct{}{}
		log.Printf("peer target %s added\n", t)
		return nil
	}

	peerDelHandler := func(t tunnel.Target) error {
		peerMux.Lock()
		defer peerMux.Unlock()
		if _, ok := peers[t]; ok {
			delete(peers, t)
			log.Printf("peer target %s deleted\n", t)
		}
		return nil
	}

	targets := make(map[tunnel.Target]struct{})
	client, err := tunnel.NewClient(tpb.NewTunnelClient(clientConn), tunnel.ClientConfig{
		PeerAddHandler: peerAddHandler,
		PeerDelHandler: peerDelHandler,
		Subscriptions:  []string{dialTargetType},
	}, targets)

	if err != nil {
		return fmt.Errorf("failed to create tunnel client: %v", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
		if err := client.Register(ctx); err != nil {
			errCh <- err
			return
		}
		client.Start(ctx)
		if err := client.Error(); err != nil {
			errCh <- err
		}
	}()

	dialTarget := tunnel.Target{ID: dialTargetCfg, Type: dialTargetType}
	foundDialTarget := func() bool {
		peerMux.Lock()
		defer peerMux.Unlock()
		_, ok := peers[dialTarget]
		return ok
	}

	go func() {
		bo := getBackOff()
		for !foundDialTarget() {
			wait := bo.NextBackOff()
			log.Printf("dial target %s (type: %s) not found. reconnecting in %s (all targets found: %s) \n", dialTargetCfg, dialTargetType, wait, peers)
			time.Sleep(wait)
		}

		session, err := client.NewSession(dialTarget)
		if err != nil {
			log.Printf("error from new session: %v", err)
			errCh <- err
			return
		}
		log.Printf("new session established for target: %s\n", dialTarget)

		var (
			retry   bool
			opts    []grpc.DialOption
			tryGnmi bool
		)

		if jctx.config.Vendor.Gnmi != nil {
			tryGnmi = true
		}

	connect:
		vendor, err := getVendor(jctx, tryGnmi)
		if err != nil {
			log.Fatalf("Could not get the vendor: %v", err)
		}

		if opts, err = getGPRCDialOptions(jctx, vendor); err != nil {
			jLog(jctx, fmt.Sprintf("%v", err))
			statusch <- struct{}{}
			errCh <- err
			return
		}

		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))

		md := metadata.New(map[string]string{"username": jctx.config.User, "password": jctx.config.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)

		withContextDialer := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return &Conn{session}, nil
		})
		opts = append(opts, withContextDialer)

		if retry {
			jLog(jctx, fmt.Sprintf("Reconnecting to target %s", dialTargetCfg))
		} else {
			jLog(jctx, fmt.Sprintf("Connecting to target %s", dialTargetCfg))
		}

		conn, err := grpc.DialContext(ctx, "use_tunnel:0", opts...)
		if err != nil {
			log.Fatalf("Could not dial target: %v", err)
		}
		fmt.Printf("Successfully dialed target %s(%s)", dialTargetCfg, dialTargetType)

		// we are able to Dial grpc, now let's begin by sending LoginCheck
		// if required.
		if vendor.loginCheckRequired {
			if err := vendor.sendLoginCheck(jctx, conn); err != nil {
				jLog(jctx, fmt.Sprintf("%v", err))
				time.Sleep(10 * time.Second)
				retry = true
				conn.Close()
				if jctx.config.Vendor.Gnmi != nil {
					tryGnmi = true
				} else {
					tryGnmi = false
				}
				goto connect
			}
		}

		if vendor.subscribe == nil {
			panic(fmt.Sprintf("could not found subscribe implementation for vendor %s", vendor.name))
		}
		fmt.Println("Calling subscribe() :::", jctx.file)
		code := vendor.subscribe(conn, jctx)
		fmt.Println("Returns subscribe() :::", jctx.file, "CODE ::: ", code)

		// close the current connection and retry
		conn.Close()
		switch code {
		case SubRcSighupRestart:
			jLog(jctx, fmt.Sprintf("sighup detected, reconnect with new config for worker %s", jctx.file))
			retry = true
			if jctx.config.Vendor.Gnmi != nil {
				tryGnmi = true
			} else {
				tryGnmi = false
			}
			goto connect
		case SubRcRPCFailedNoRetry:
			jLog(jctx, fmt.Sprintf("RPC failed and reconnecting with fallback RPC if available %s", jctx.file))
			retry = true
			if tryGnmi {
				tryGnmi = false // fallback to vendor mode
			}
			goto connect
		case SubRcConnRetry:
			jLog(jctx, fmt.Sprintf("subscribe returns, reconnecting after 10s for worker %s", jctx.file))
			time.Sleep(10 * time.Second)
			retry = true
			if jctx.config.Vendor.Gnmi != nil {
				tryGnmi = true
			} else {
				tryGnmi = false
			}
			goto connect
		case SubRcSighupNoRestart:
			jLog(jctx, fmt.Sprintf("not reconnecting for worker %s", jctx.file))
			statusch <- struct{}{}
			errCh <- err
			return
		}

	}()

	// Listen for any request to create a new session.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return fmt.Errorf("exiting: %s", err)
	}
}

func work(jctx *JCtx, statusch chan struct{}) {
	if jctx.config.GRPC.TunnelServer.Address != "" {
		if err := workTunnel(jctx, statusch); err != nil {
			jLog(jctx, fmt.Sprintf("tunnel worker failed: %v", err))
		}
		return
	}

	var (
		retry   bool
		opts    []grpc.DialOption
		tryGnmi bool
	)

	if jctx.config.Vendor.Gnmi != nil {
		tryGnmi = true
	}

connect:
	// Read the host-name and vendor from the config as they might be changed
	vendor, err := getVendor(jctx, tryGnmi)
	if err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		time.Sleep(10 * time.Second)
		retry = true
		goto connect
	}
	if opts, err = getGPRCDialOptions(jctx, vendor); err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		statusch <- struct{}{}
		return
	}

	hostname := jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
	if hostname == ":0" {
		statusch <- struct{}{}
		jLog(jctx, fmt.Sprintf("Not a valid host-name %s", hostname))
		return
	}

	// Check signals
	select {
	case s := <-jctx.control:
		switch s {
		case os.Interrupt:
			// we are done
			jLog(jctx, fmt.Sprintf("Connection for %s has been interrupted", hostname))
			statusch <- struct{}{}
			return
		}
		// Sighup need not be handled as the config is re-read for
		// each connection attempt
	default:
		// No signal recieved, Continue the connection attempt
	}

	if retry {
		jLog(jctx, fmt.Sprintf("Reconnecting to %s", hostname))
	} else {
		jLog(jctx, fmt.Sprintf("Connecting to %s", hostname))
	}
	conn, err := grpc.Dial(hostname, opts...)
	if err != nil {
		jLog(jctx, fmt.Sprintf("[%s] could not dial: %v", jctx.config.Host, err))
		time.Sleep(10 * time.Second)
		retry = true
		goto connect
	}

	// we are able to Dial grpc, now let's begin by sending LoginCheck
	// if required.
	if vendor.loginCheckRequired {
		if err := vendor.sendLoginCheck(jctx, conn); err != nil {
			jLog(jctx, fmt.Sprintf("%v", err))
			time.Sleep(10 * time.Second)
			retry = true
			conn.Close()
			if jctx.config.Vendor.Gnmi != nil {
				tryGnmi = true
			} else {
				tryGnmi = false
			}
			goto connect
		}
	}

	if vendor.subscribe == nil {
		panic(fmt.Sprintf("could not found subscribe implementation for vendor %s", vendor.name))
	}
	fmt.Println("Calling subscribe() :::", jctx.file)
	code := vendor.subscribe(conn, jctx)
	fmt.Println("Returns subscribe() :::", jctx.file, "CODE ::: ", code)

	// close the current connection and retry
	conn.Close()

	switch code {
	case SubRcSighupRestart:
		jLog(jctx, fmt.Sprintf("sighup detected, reconnect with new config for worker %s", jctx.file))
		retry = true
		if jctx.config.Vendor.Gnmi != nil {
			tryGnmi = true
		} else {
			tryGnmi = false
		}
		goto connect
	case SubRcRPCFailedNoRetry:
		jLog(jctx, fmt.Sprintf("RPC failed and reconnecting with fallback RPC if available %s", jctx.file))
		retry = true
		if tryGnmi {
			tryGnmi = false // fallback to vendor mode
		}
		goto connect
	case SubRcConnRetry:
		jLog(jctx, fmt.Sprintf("subscribe returns, reconnecting after 10s for worker %s", jctx.file))
		time.Sleep(10 * time.Second)
		retry = true
		if jctx.config.Vendor.Gnmi != nil {
			tryGnmi = true
		} else {
			tryGnmi = false
		}
		goto connect
	case SubRcSighupNoRestart:
		jLog(jctx, fmt.Sprintf("not reconnecting for worker %s", jctx.file))
		statusch <- struct{}{}
		return
	}
}
