package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
	influxCtx InfluxCtx
	stats     statsCtx
	pExporter *jtimonPExporter
	control   chan os.Signal
	running   bool
	alias     *Alias
	testMeta  *os.File
	testBytes *os.File
	testExp   *os.File
	testRes   *os.File
}

// JWorkers holds worker
type JWorkers struct {
	m        map[string]*JWorker
	wg       sync.WaitGroup
	mr       int64
	files    []string
	fileList string
	sigchan  chan os.Signal
}

// NewJWorkers to create new workers
func NewJWorkers(files []string, fileList string, mr int64) *JWorkers {
	return &JWorkers{
		m:        make(map[string]*JWorker),
		mr:       mr,
		files:    files,
		fileList: fileList,
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
	for _, v := range ws.m {
		v.signalch <- syscall.SIGCONT
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
	if w, err := NewJWorker(file, &ws.wg); err == nil {
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
		s := <-sigchan
		switch s {
		case syscall.SIGHUP:
			// propagate the signal to workers and continue waiting for signals
			if len(ws.fileList) != 0 {
				ws.handleConfigChanges()
			}
		case os.Interrupt:
			for _, w := range ws.m {
				w.signalch <- s
			}
			return
		}
	}
}

// JWorker is one worker
type JWorker struct {
	jctx     *JCtx
	signalch chan os.Signal
}

// NewJWorker is to create new worker
func NewJWorker(file string, wg *sync.WaitGroup) (*JWorker, error) {
	w := &JWorker{}

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
	if alias, err := NewAlias(jctx.config.Alias); err == nil {
		jctx.alias = alias
	}
	go func() {
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
					} else if jctx.running {
						if restart {
							jctx.control <- syscall.SIGHUP
							jctx.running = false
						}
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
					logStop(&jctx)
					return
				case true:
					jctx.running = true
				}
			}
		}
	}()
	return w, nil
}

func work(jctx *JCtx, statusch chan bool) {
	var retry bool
	var opts []grpc.DialOption

connect:
	// Read the host-name and vendor from the config as they might be changed
	vendor, err := getVendor(jctx)
	if opts, err = getGPRCDialOptions(jctx, vendor); err != nil {
		jLog(jctx, fmt.Sprintf("%v", err))
		statusch <- false
		return
	}

	hostname := jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
	if hostname == ":0" {
		statusch <- false
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
			goto connect
		}
	}

	if vendor.subscribe == nil {
		panic(fmt.Sprintf("could not found subscribe implementation for vendor %s", vendor.name))
	}
	code := vendor.subscribe(conn, jctx, statusch)

	// close the current connection and retry
	conn.Close()

	switch code {
	case SubRcSighupRestart:
		jLog(jctx, fmt.Sprintf("sighup detected, reconnect with new config for worker %s", jctx.file))
		retry = true
		goto connect
	case SubRcConnRetry:
		jLog(jctx, fmt.Sprintf("subscribe returns, reconnecting after 10s for worker %s", jctx.file))
		time.Sleep(10 * time.Second)
		retry = true
		goto connect
	case SubRcSighupNoRestart:
		jLog(jctx, fmt.Sprintf("not reconnecting for worker %s", jctx.file))
		return
	}
}
