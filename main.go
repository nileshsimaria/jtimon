package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	auth_pb "github.com/nileshsimaria/jtimon/authentication"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	cfgFile      = flag.String("config", "", "Config file name")
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
	st           statsType
)

var (
	Version   = "version-not-available"
	BuildTime = "build-time-not-available"
)

type jcontext struct {
	cfg   config
	dMap  map[uint32]map[uint32]map[string]dropData
	iFlux iFluxCtx
	pause struct {
		pch  chan int64
		upch chan struct{}
	}
}

func main() {
	st.startTime = time.Now()
	flag.Parse()
	if *version {
		fmt.Println("Version:   ", Version)
		fmt.Println("BuildTime: ", BuildTime)
		return
	}
	fmt.Println("Version:   ", Version)
	fmt.Println("BuildTime: ", BuildTime)

	jctx := jcontext{}
	jctx.cfg = configInit(*cfgFile)
	jctx.cfg.CStats.pstats = *pstats
	jctx.cfg.CStats.csv_stats = *csvStats

	logInit(&jctx, *logFile)
	go prometheusHandler(*prometheus)
	start_gtrace(*gtrace)
	go maxRun(&jctx, *mr)
	go periodicStats(*pstats)

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
	if jctx.cfg.Tls.Ca != "" {
		certificate, err := tls.LoadX509KeyPair(jctx.cfg.Tls.ClientCrt, jctx.cfg.Tls.ClientKey)

		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(jctx.cfg.Tls.Ca)
		if err != nil {
			log.Fatalf("failed to read ca cert: %s", err)
		}

		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			log.Fatal("failed to append certs")
		}

		transportCreds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			ServerName:   jctx.cfg.Tls.ServerName,
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
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	if jctx.cfg.User != "" && jctx.cfg.Password != "" {
		user := jctx.cfg.User
		pass := jctx.cfg.Password
		if jctx.cfg.Meta == false {
			l := auth_pb.NewLoginClient(conn)
			dat, err := l.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: user, Password: pass, ClientId: jctx.cfg.Cid})
			if err != nil {
				log.Fatalf("Could not login: %v", err)
			}
			if dat.Result == false {
				log.Fatalf("LoginCheck failed\n")
			}
		}
	}

	if *gnmi {
		subscribe_gnmi(conn, &jctx)
	} else {
		subscribe(conn, &jctx)
	}
}
