package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
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
	running   bool
	testMeta  *os.File
	testBytes *os.File
	testExp   *os.File
	testRes   *os.File
}

type workerCtx struct {
	signalch chan<- os.Signal
	err      error
}

func work(jctx *JCtx, statusch chan bool) {
	var retry bool
	var opts []grpc.DialOption

	vendor, err := getVendor(jctx)
	if *conTestData && vendor.consumeTestData != nil {
		vendor.consumeTestData(jctx)
		statusch <- false
		return
	}

	if opts, err = getGPRCDialOptions(jctx, vendor); err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		statusch <- false
		return
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
			goto connect
		}
	}

	if vendor.subscribe == nil {
		panic(fmt.Sprintf("could not found subscribe implementation for vendor %s", vendor.name))
	}
	res := vendor.subscribe(conn, jctx, statusch)

	// close the current connection and retry
	conn.Close()

	if res == SubRcSighupRestart {
		jLog(jctx, fmt.Sprintf("restarting the connection for config changes"))
	} else {
		jLog(jctx, fmt.Sprintf("subscribe returns, retrying the connection"))
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

	if *genTestData {
		var errf error
		jctx.testMeta, errf = os.Create(file + ".testmeta")
		if errf != nil {
			log.Printf("Could not create 'testmeta' for file %s\n", file+".testmeta")
		}

		jctx.testBytes, errf = os.Create(file + ".testbytes")
		if errf != nil {
			log.Printf("Could not create 'testbytes' for file %s\n", file+".testbytes")
		}

		jctx.testExp, errf = os.Create(file + ".testexp")
		if errf != nil {
			log.Printf("Could not create 'testexp' for file %s\n", file+".testexp")
		}
	}

	err := ConfigRead(&jctx, true)
	if err != nil {
		log.Println(err)
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
					if *genTestData {
						if jctx.testMeta != nil {
							jctx.testMeta.Close()
						}
						if jctx.testBytes != nil {
							jctx.testBytes.Close()
						}
						if jctx.testExp != nil {
							jctx.testExp.Close()
						}
					}

					jctx.wg.Done()
				case syscall.SIGHUP:
					// handle SIGHUP if the streaming is happening.
					// running will not be set when the connection is
					// not establihsed and it is trying to connect.
					// ConfigRead will re-parse the config and updates jctx so
					// when we retry Dial, it will do it with updated config
					err := ConfigRead(&jctx, false)
					if err != nil {
						jLog(&jctx, fmt.Sprintln(err))
					} else if jctx.running {
						jctx.pause.subch <- struct{}{}
						jctx.running = false
					} else {
						jLog(&jctx, fmt.Sprintf("config re-parse, data streaming has not started yet"))
					}
				case syscall.SIGCONT:
					go work(&jctx, statusch)
				}
			case status := <-statusch:
				switch status {
				case false:
					// worker must have encountered error
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
