package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	auth_pb "github.com/nileshsimaria/jtimon/authentication"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// DefaultGRPCWindowSize is the default GRPC Window Size
	DefaultGRPCWindowSize = 1048576
)

var (
	cfgFile        = flag.StringSlice("config", make([]string, 0, 0), "Config file name(s)")
	expConfig      = flag.Bool("explore-config", false, "Explore full config of JTIMON and exit")
	print          = flag.Bool("print", false, "Print Telemetry data")
	mr             = flag.Int64("max-run", 0, "Max run time in seconds")
	stateHandler   = flag.Bool("stats-handler", false, "Use GRPC statshandler")
	ver            = flag.Bool("version", false, "Print version and build-time of the binary and exit")
	compression    = flag.String("compression", "", "Enable HTTP/2 compression (gzip, deflate)")
	latencyProfile = flag.Bool("latency-profile", false, "Profile latencies. Place them in TSDB")
	gnmi           = flag.Bool("gnmi", false, "Use gnmi proto")
	gnmiMode       = flag.String("gnmi-mode", "stream", "Mode of gnmi (stream | once | poll")
	gnmiEncoding   = flag.String("gnmi-encoding", "proto", "gnmi encoding (proto | json | bytes | ascii | ietf-json")
	prom           = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	promPort       = flag.Int32("prometheus-port", 8090, "Prometheus port")
	prefixCheck    = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	apiControl     = flag.Bool("api", false, "Receive HTTP commands when running")
	pProf          = flag.Bool("pprof", false, "Profile JTIMON")
	pProfPort      = flag.Int32("pprof-port", 6060, "Profile port")
	gtrace         = flag.Bool("gtrace", false, "Collect GRPC traces")
	grpcHeaders    = flag.Bool("grpc-headers", false, "Add grpc headers in DB")

	version   = "version-not-available"
	buildTime = "build-time-not-available"
	gmutex    = &sync.Mutex{}
)

// JCtx is JTIMON run time context
type JCtx struct {
	config    Config
	file      string
	index     int
	wg        *sync.WaitGroup
	dMap      map[uint32]map[uint32]map[string]dropData
	influxCtx InfluxCtx
	stats     statsCtx
	pause     struct {
		pch  chan int64
		upch chan struct{}
	}
}

type workerCtx struct {
	ch  chan bool
	err error
}

// A worker function is the one who gets job done.
func worker(file string, idx int, wg *sync.WaitGroup) (chan bool, error) {
	ch := make(chan bool)
	jctx := JCtx{
		file:  file,
		index: idx,
		wg:    wg,
		stats: statsCtx{
			startTime: time.Now(),
		},
	}

	var err error
	jctx.config, err = NewJTIMONConfig(file)
	if err != nil {
		fmt.Printf("\nConfig parsing error for %s[%d]: %v\n", file, idx, err)
		return ch, fmt.Errorf("config parsing (json Unmarshal) error for %s[%d]: %v", file, idx, err)
	}

	logInit(&jctx)
	b, err := json.MarshalIndent(jctx.config, "", "    ")
	if err != nil {
		return ch, fmt.Errorf("Config parsing error (json Marshal) for %s[%d]: %v", file, idx, err)
	}
	l(true, &jctx, fmt.Sprintf("\nRunning config of JTIMON:\n %s\n", string(b)))

	go periodicStats(&jctx)
	influxInit(&jctx)
	dropInit(&jctx)
	go apiInit(&jctx)

	if *grpcHeaders {
		pmap := make(map[string]interface{})
		for i := range jctx.config.Paths {
			pmap["path"] = jctx.config.Paths[i].Path
			pmap["reporting-rate"] = float64(jctx.config.Paths[i].Freq)
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
						var retry bool
						var opts []grpc.DialOption

						if jctx.config.TLS.CA != "" {
							certificate, err := tls.LoadX509KeyPair(jctx.config.TLS.ClientCrt, jctx.config.TLS.ClientKey)

							certPool := x509.NewCertPool()
							bs, err := ioutil.ReadFile(jctx.config.TLS.CA)
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
								ServerName:   jctx.config.TLS.ServerName,
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

						ws := jctx.config.GRPC.WS
						opts = append(opts, grpc.WithInitialWindowSize(ws))

						hostname := jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
						if hostname == ":0" {
							return
						}
					connect:
						if retry {
							l(true, &jctx, fmt.Sprintf("Reconnecting to %s", hostname))
						} else {
							l(true, &jctx, fmt.Sprintf("Connecting to %s", hostname))
						}
						conn, err := grpc.Dial(hostname, opts...)
						if err != nil {
							l(true, &jctx, fmt.Sprintf("[%d] Could not dial: %v\n", idx, err))
							time.Sleep(10 * time.Second)
							retry = true
							goto connect
						}

						if jctx.config.User != "" && jctx.config.Password != "" {
							user := jctx.config.User
							pass := jctx.config.Password
							if jctx.config.Meta == false {
								lc := auth_pb.NewLoginClient(conn)
								dat, err := lc.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: user, Password: pass, ClientId: jctx.config.CID})
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
						// If we are here we must try to reconnect again.
						// Reconnect after 10 seconds.
						time.Sleep(10 * time.Second)
						retry = true
						goto connect
					}()
				}
			}
		}
	}()

	return ch, nil
}

func main() {
	flag.Parse()
	if *pProf {
		go func() {
			addr := fmt.Sprintf("localhost:%d", *pProfPort)
			fmt.Println(http.ListenAndServe(addr, nil))
		}()
	}

	if *prom {
		go func() {
			addr := fmt.Sprintf("localhost:%d", promPort)
			http.Handle("/metrics", promhttp.Handler())
			fmt.Println(http.ListenAndServe(addr, nil))
		}()

	}
	startGtrace(*gtrace)

	fmt.Printf("Version: %s BuildTime %s\n", version, buildTime)
	if *ver {
		return
	}

	if *expConfig {
		config, err := ExploreConfig()
		if err == nil {
			fmt.Printf("\n%s\n", config)
		} else {
			fmt.Printf("Can not generate config\n")
		}
		return
	}

	n := len(*cfgFile)
	if n == 0 {
		fmt.Println("Can not run without any config file")
		return
	}

	var wg sync.WaitGroup
	wg.Add(n)
	wList := make([]*workerCtx, n, n)

	for idx, file := range *cfgFile {
		ch, err := worker(file, idx, &wg)
		if err != nil {
			wg.Done()
		}
		wList[idx] = &workerCtx{
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
