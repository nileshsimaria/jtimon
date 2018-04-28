package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strconv"
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
	logFile      = flag.String("log", "", "Log file name")
	gtrace       = flag.Bool("gtrace", false, "Collect GRPC traces")
	version      = flag.Bool("version", false, "Print version and build-time of the binary and exit")
	gnmi         = flag.Bool("gnmi", false, "Use gnmi proto")
	td           = flag.Bool("time-diff", false, "Time Diff for sensor analysis using InfluxDB")
	dcheck       = flag.Bool("drop-check", false, "Check for packet drops")
	lcheck       = flag.Bool("latency-check", false, "Check for latency")
	prometheus   = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	print        = flag.Bool("print", false, "Print Telemetry data")
	prefixCheck  = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	sleep        = flag.Int64("sleep", 0, "Sleep after each read (ms)")
	mr           = flag.Int64("max-run", 0, "Max run time in seconds")
	maxKV        = flag.Uint64("max-kv", 0, "Max kv")
	pstats       = flag.Int64("stats", 0, "Print collected stats periodically")
	csvStats     = flag.Bool("csv-stats", false, "Capture size of each telemetry packet")
	compression  = flag.String("compression", "", "Enable HTTP/2 compression (gzip, deflate)")
)

var (
	// Version (Version of the binary, supplied at the build time)
	Version = "version-not-available"
	// BuildTime (Time of the build, supplied at the build time)
	BuildTime = "build-time-not-available"
)

type jcontext struct {
	cfg   config
	file  string
	idx   int
	dMap  map[uint32]map[uint32]map[string]dropData
	iFlux iFluxCtx
	st    statsType
	pause struct {
		pch  chan int64
		upch chan struct{}
	}
}

func main() {
	flag.Parse()

	fmt.Println("Version:   ", Version)
	fmt.Println("BuildTime: ", BuildTime)
	if *version {
		return
	}

	if len(*cfgFile) == 0 {
		fmt.Println("Can not run JTIMON without any config file")
		return
	}

	for idx, file := range *cfgFile {
		fmt.Printf("Starting go-routine for %s[%d]\n", file, idx)

		go func(file string, idx int) {
			jctx := jcontext{}
			jctx.file = file
			jctx.idx = idx
			jctx.st.startTime = time.Now()
			jctx.cfg = configInit(file)
			jctx.cfg.CStats.pStats = *pstats
			jctx.cfg.CStats.csvStats = *csvStats

			logInit(&jctx, *logFile)
			go prometheusHandler(*prometheus)
			startGtrace(*gtrace)
			go maxRun(&jctx, *mr)
			go periodicStats(&jctx, *pstats)

			jctx.iFlux.influxc = influxInit(jctx.cfg)
			configValidation(&jctx)

			dropInit(&jctx)
			go apiInit(&jctx)

			pmap := make(map[string]interface{})
			for i := range jctx.cfg.Paths {
				pmap["path"] = jctx.cfg.Paths[i].Path
				pmap["reporting-rate"] = float64(jctx.cfg.Paths[i].Freq)
				addGRPCHeader(&jctx, pmap)
			}

			var opts []grpc.DialOption
			if jctx.cfg.TLS.CA != "" {
				certificate, err := tls.LoadX509KeyPair(jctx.cfg.TLS.ClientCrt, jctx.cfg.TLS.ClientKey)

				certPool := x509.NewCertPool()
				bs, err := ioutil.ReadFile(jctx.cfg.TLS.CA)
				if err != nil {
					fmt.Printf("[%d] Failed to read ca cert: %s\n", idx, err)
					return
				}

				ok := certPool.AppendCertsFromPEM(bs)
				if !ok {
					fmt.Printf("[%d] Failed to append certs\n", idx)
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

			opts = append(opts, grpc.WithStatsHandler(&statshandler{jctx: &jctx}))
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

			if jctx.cfg.Grpc.Ws != 0 {
				opts = append(opts, grpc.WithInitialWindowSize(jctx.cfg.Grpc.Ws))
			}

			hostname := jctx.cfg.Host + ":" + strconv.Itoa(jctx.cfg.Port)
			conn, err := grpc.Dial(hostname, opts...)
			if err != nil {
				fmt.Printf("[%d] Could not connect: %v\n", idx, err)
				return
			}
			defer conn.Close()

			if jctx.cfg.User != "" && jctx.cfg.Password != "" {
				user := jctx.cfg.User
				pass := jctx.cfg.Password
				if jctx.cfg.Meta == false {
					l := auth_pb.NewLoginClient(conn)
					dat, err := l.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: user, Password: pass, ClientId: jctx.cfg.Cid})
					if err != nil {
						fmt.Printf("[%d] Could not login: %v\n", idx, err)
						return
					}
					if dat.Result == false {
						fmt.Printf("[%d] LoginCheck failed", idx)
						return
					}
				}
			}

			if *gnmi {
				subscribeGNMI(conn, &jctx)
			} else {
				subscribe(conn, &jctx)
			}
		}(file, idx)
	}
	select {}
}
