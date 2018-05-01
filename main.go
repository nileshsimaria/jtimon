package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	auth_pb "github.com/nileshsimaria/jtimon/authentication"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	cfgFile      = flag.StringSlice("config", make([]string, 0, 0), "Config file name(s)")
	gnmiMode     = flag.String("gnmi-mode", "stream", "Mode of gnmi (stream | once | poll")
	gnmiEncoding = flag.String("gnmi-encoding", "proto", "gnmi encoding (proto | json | bytes | ascii | ietf-json")
	gtrace       = flag.Bool("gtrace", false, "Collect GRPC traces")
	stateHandler = flag.Bool("stats-handler", false, "Use GRPC statshandler")
	grpcHeaders  = flag.Bool("grpc-headers", false, "Add grpc headers in DB")
	ver          = flag.Bool("version", false, "Print version and build-time of the binary and exit")
	gnmi         = flag.Bool("gnmi", false, "Use gnmi proto")
	prometheus   = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	print        = flag.Bool("print", false, "Print Telemetry data")
	prefixCheck  = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	apiControl   = flag.Bool("api", false, "Receive HTTP commands when running")
	mr           = flag.Int64("max-run", 0, "Max run time in seconds")
	compression  = flag.String("compression", "", "Enable HTTP/2 compression (gzip, deflate)")

	version   = "version-not-available"
	buildTime = "build-time-not-available"
	gmutex    = &sync.Mutex{}
)

// JCtx is JTIMON Context
type JCtx struct {
	cfg   config
	file  string
	idx   int
	wg    *sync.WaitGroup
	dMap  map[uint32]map[uint32]map[string]dropData
	iFlux iFluxCtx
	st    statsType
	pause struct {
		pch  chan int64
		upch chan struct{}
	}
}

type winfo struct {
	ch  chan bool
	err error
}

func worker(file string, idx int, wg *sync.WaitGroup) (chan bool, error) {
	ch := make(chan bool)
	jctx := JCtx{
		file: file,
		idx:  idx,
		wg:   wg,
		st: statsType{
			startTime: time.Now(),
		},
	}

	var err error
	jctx.cfg, err = configInit(file)
	if err != nil {
		fmt.Printf("\nConfig parsing error for %s[%d]: %v\n", file, idx, err)
		return ch, fmt.Errorf("config parsing error for %s[%d]: %v", file, idx, err)
	}

	logInit(&jctx)
	go prometheusHandler(*prometheus)
	go periodicStats(&jctx)
	jctx.iFlux.influxc = influxInit(jctx.cfg)
	dropInit(&jctx)
	go apiInit(&jctx)

	if *grpcHeaders {
		pmap := make(map[string]interface{})
		for i := range jctx.cfg.Paths {
			pmap["path"] = jctx.cfg.Paths[i].Path
			pmap["reporting-rate"] = float64(jctx.cfg.Paths[i].Freq)
			addGRPCHeader(&jctx, pmap)
		}
	}

	go func() {
		for {
			select {
			case ctrl := <-ch:
				switch ctrl {
				case false:
					printSummary(&jctx)
					jctx.wg.Done()
				case true:
					go func() {
						var opts []grpc.DialOption
						if jctx.cfg.TLS.CA != "" {
							certificate, err := tls.LoadX509KeyPair(jctx.cfg.TLS.ClientCrt, jctx.cfg.TLS.ClientKey)

							certPool := x509.NewCertPool()
							bs, err := ioutil.ReadFile(jctx.cfg.TLS.CA)
							if err != nil {
								l(true, &jctx, fmt.Sprintf("[%d] Failed to read ca cert: %s\n", idx, err))
								return
							}

							ok := certPool.AppendCertsFromPEM(bs)
							if !ok {
								l(true, &jctx, fmt.Sprintf("[%d] Failed to append certs\n", idx))
								return
							}

							transportCreds := credentials.NewTLS(&tls.Config{
								Certificates: []tls.Certificate{certificate},
								ServerName:   jctx.cfg.TLS.ServerName,
								RootCAs:      certPool,
							})
							opts = append(opts, grpc.WithTransportCredentials(transportCreds))
						} else {
							opts = append(opts, grpc.WithInsecure())
						}

						if *stateHandler {
							opts = append(opts, grpc.WithStatsHandler(&statshandler{jctx: &jctx}))
						}
						if *compression != "" {
							var dc grpc.Decompressor
							if *compression == "gzip" {
								dc = grpc.NewGZIPDecompressor()
							} else if *compression == "deflate" {
								dc = newDEFLATEDecompressor()
							}
							compressionOpts := grpc.Decompressor(dc)
							opts = append(opts, grpc.WithDecompressor(compressionOpts))
						}
						ws := jctx.cfg.Grpc.Ws
						if ws == 0 {
							ws = 1048576
						}
						opts = append(opts, grpc.WithInitialWindowSize(ws))

						hostname := jctx.cfg.Host + ":" + strconv.Itoa(jctx.cfg.Port)
						conn, err := grpc.Dial(hostname, opts...)
						if err != nil {
							l(true, &jctx, fmt.Sprintf("[%d] Could not dial: %v\n", idx, err))
							return
						}

						if jctx.cfg.User != "" && jctx.cfg.Password != "" {
							user := jctx.cfg.User
							pass := jctx.cfg.Password
							if jctx.cfg.Meta == false {
								lc := auth_pb.NewLoginClient(conn)
								dat, err := lc.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: user, Password: pass, ClientId: jctx.cfg.Cid})
								if err != nil {
									l(true, &jctx, fmt.Sprintf("[%d] Could not login: %v\n", idx, err))
									return
								}
								if dat.Result == false {
									l(true, &jctx, fmt.Sprintf("[%d] LoginCheck failed", idx))
									return
								}
							}
						}

						if *gnmi {
							subscribeGNMI(conn, &jctx)
						} else {
							subscribe(conn, &jctx)
						}
					}()
				}
			}
		}
	}()

	return ch, nil
}

func main() {
	flag.Parse()
	startGtrace(*gtrace)

	fmt.Printf("Version: %s BuildTime %s\n", version, buildTime)
	if *ver {
		return
	}

	n := len(*cfgFile)
	if n == 0 {
		fmt.Println("Can not run without any config file")
		return
	}

	var wg sync.WaitGroup
	wg.Add(n)
	wList := make([]*winfo, n, n)

	for idx, file := range *cfgFile {
		ch, err := worker(file, idx, &wg)
		if err != nil {
			wg.Done()
		}
		wList[idx] = &winfo{
			ch:  ch,
			err: err,
		}
	}

	for _, worker := range wList {
		if worker.err == nil {
			worker.ch <- true
		}
	}

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		for _, worker := range wList {
			if worker.err == nil {
				worker.ch <- false
			}
		}
	}()

	go func() {
		if *mr == 0 {
			return
		}
		tickChan := time.NewTimer(time.Second * time.Duration(*mr)).C
		<-tickChan
		for _, worker := range wList {
			if worker.err == nil {
				worker.ch <- false
			}
		}
	}()
	wg.Wait()
	fmt.Printf("All done ... exiting!\n")
}
