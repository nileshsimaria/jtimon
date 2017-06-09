package main

import (
	auth_pb "github.com/nileshsimaria/jtimon/authentication"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"strconv"
	"time"
)

var (
	cert               = flag.String("cert", "", "CA certificate file")
	cfgFile            = flag.String("config", "", "Config file name")
	logFile            = flag.String("log", "", "Log file name")
	gtrace             = flag.Bool("gtrace", false, "Collect GRPC traces")
	td                 = flag.Bool("time-diff", false, "Time Diff for sensor analysis using InfluxDB")
	dcheck             = flag.Bool("drop-check", false, "Check for packet drops")
	lcheck             = flag.Bool("latency-check", false, "Check for latency")
	prometheus         = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	print              = flag.Bool("print", false, "Print Telemetry data")
	prefixCheck        = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	sleep              = flag.Int64("sleep", 0, "Sleep after each read (ms)")
	mr                 = flag.Int64("max-run", 0, "Max run time in seconds")
	maxKV              = flag.Int64("max-kv", 0, "Max kv")
	pstats             = flag.Int64("stats", 0, "Collect and Print statistics periodically")
	compression        = flag.String("compression", "", "Enable HTTP/2 compression (gzip, deflate)")
	serverHostOverride = flag.String("server-host-override", "", "ServerName used to verify the hostname")
	tls                = flag.Bool("tls", false, "Connection uses TLS")
	st                 statsType
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

	jctx := jcontext{}
	jctx.cfg = configInit(*cfgFile)

	go prometheusHandler(*prometheus)
	start_gtrace(*gtrace)
	go maxRun(&jctx, *mr)
	go periodicStats(*pstats)
	logInit(*logFile)

	jctx.iFlux.influxc = influxInit(jctx.cfg)

	dropInit(&jctx)
	go apiInit(&jctx)

	var opts []grpc.DialOption
	if *tls {
		var sn string
		if *serverHostOverride != "" {
			sn = *serverHostOverride
		}
		if *cert != "" {
			creds, err := credentials.NewClientTLSFromFile(*cert, sn)
			if err != nil {
				log.Fatalf("%v", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(creds))
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if *pstats != 0 {
		opts = append(opts, grpc.WithStatsHandler(&statshandler{}))
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
		l := auth_pb.NewLoginClient(conn)
		dat, err := l.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: user, Password: pass, ClientId: jctx.cfg.Cid})
		if err != nil {
			log.Fatalf("Could not login: %v", err)
		}
		if dat.Result == false {
			log.Fatalf("LoginCheck failed\n")
		}
	}

	subscribe(conn, &jctx)
}
