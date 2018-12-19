package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	auth_pb "github.com/nileshsimaria/jtimon/authentication"
)

// JCtx is JTIMON run time context
type JCtx struct {
	config    Config
	file      string
	wg        *sync.WaitGroup
	dMap      map[uint32]map[uint32]map[string]dropData
	influxCtx InfluxCtx
	stats     statsCtx
	pExporter *jtimonPExporter
	pause     struct {
		pch   chan int64
		upch  chan struct{}
		subch chan struct{}
	}
	running bool
}

type workerCtx struct {
	signalch chan<- os.Signal
	err      error
}

func getSecurityOptions(jctx *JCtx) (grpc.DialOption, error) {
	var bs []byte
	var err error

	if jctx.config.TLS.CA == "" {
		return grpc.WithInsecure(), nil
	}

	certificate, _ := tls.LoadX509KeyPair(jctx.config.TLS.ClientCrt, jctx.config.TLS.ClientKey)
	certPool := x509.NewCertPool()
	if bs, err = ioutil.ReadFile(jctx.config.TLS.CA); err != nil {
		return nil, fmt.Errorf("[%s] failed to read ca cert: %s", jctx.config.Host, err)
	}

	if ok := certPool.AppendCertsFromPEM(bs); !ok {
		return nil, fmt.Errorf("[%s] failed to append certs", jctx.config.Host)
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   jctx.config.TLS.ServerName,
		RootCAs:      certPool,
	})

	return grpc.WithTransportCredentials(transportCreds), nil
}

func work(jctx *JCtx, statusch chan bool) {
	var retry bool
	var opts []grpc.DialOption
	vendor, err := getVendor(jctx)

	if err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		statusch <- false
		return
	}

	securityOpt, err := getSecurityOptions(jctx)
	if err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		statusch <- false
		return
	}
	opts = append(opts, securityOpt)

	if *stateHandler {
		opts = append(opts, grpc.WithStatsHandler(&statshandler{jctx: jctx}))
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

	if vendor.dialExt != nil {
		opt := vendor.dialExt(jctx)
		if opt != nil {
			opts = append(opts, opt)
		}
	}

	hostname := jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
	if hostname == ":0" {
		statusch <- false
		return
	}
connect:
	if retry {
		jLog(jctx, fmt.Sprintf("Reconnecting to %s", hostname))
	} else {
		jLog(jctx, fmt.Sprintf("Connecting to %s", hostname))
	}
	conn, err := grpc.Dial(hostname, opts...)
	if err != nil {
		jLog(jctx, fmt.Sprintf("[%s] Could not dial: %v\n", jctx.config.Host, err))
		time.Sleep(10 * time.Second)
		retry = true
		goto connect
	}

	// Close the connection on return
	defer conn.Close()

	if jctx.config.User != "" && jctx.config.Password != "" {
		user := jctx.config.User
		pass := jctx.config.Password
		if vendor.loginCheckRequired && !jctx.config.Meta {
			lc := auth_pb.NewLoginClient(conn)
			dat, err := lc.LoginCheck(context.Background(),
				&auth_pb.LoginRequest{UserName: user,
					Password: pass, ClientId: jctx.config.CID})
			if err != nil {
				jLog(jctx, fmt.Sprintf("[%s] Could not login: %v", jctx.config.Host, err))
				time.Sleep(10 * time.Second)
				retry = true
				goto connect
			}
			if !dat.Result {
				jLog(jctx, fmt.Sprintf("[%s] LoginCheck failed", jctx.config.Host))
				time.Sleep(10 * time.Second)
				retry = true
				goto connect
			}
		}
	}

	var res SubErrorCode
	if vendor.subscribe == nil {
		panic("Could not found subscribe implementation")
	}
	res = vendor.subscribe(conn, jctx, statusch)

	// Close the current connection and retry
	conn.Close()
	if res == SubRcSighupRestart {
		jLog(jctx, fmt.Sprintf("Restarting the connection for config changes\n"))
	} else {
		jLog(jctx, fmt.Sprintf("Retrying the connection"))
		// If we are here we must try to reconnect again.
		// Reconnect after 10 seconds.
		time.Sleep(10 * time.Second)
	}
	retry = true
	goto connect
}

// A worker function is the one who gets job done.
func worker(file string, wg *sync.WaitGroup) (chan<- os.Signal, error) {
	signalch := make(chan os.Signal)
	statusch := make(chan bool)
	jctx := JCtx{
		file:      file,
		wg:        wg,
		pExporter: exporter,
		stats: statsCtx{
			startTime: time.Now(),
		},
	}

	err := ConfigRead(&jctx, true)
	if err != nil {
		fmt.Println(err)
		return signalch, err
	}

	go func() {
		for {
			select {
			case sig := <-signalch:
				switch sig {
				case os.Interrupt:
					// we are asked to stop
					printSummary(&jctx)
					jLog(&jctx, fmt.Sprintf("Streaming for host %s has been stopped (SIGINT)", jctx.config.Host))
					jctx.wg.Done()
				case syscall.SIGHUP:
					// handle SIGHUP if the streaming is happening.
					//
					// running will not be set when the connection is
					// not establihsed and it is trying to connect.
					err := ConfigRead(&jctx, false)
					if err != nil {
						jLog(&jctx, fmt.Sprintln(err))
					} else if jctx.running {
						jctx.pause.subch <- struct{}{}
						jctx.running = false
					} else {
						jLog(&jctx, fmt.Sprintf("Config Re-Read, Data streaming has not started yet"))
					}
				case syscall.SIGCONT:
					go work(&jctx, statusch)
				}
			case status := <-statusch:
				switch status {
				case false:
					// Exited with error
					printSummary(&jctx)
					jctx.wg.Done()
				case true:
					jctx.running = true
				}
			}
		}
	}()
	return signalch, nil
}
