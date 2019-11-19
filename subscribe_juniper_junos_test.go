package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	flag "github.com/spf13/pflag"
)

func TestJTISIMSigHup(t *testing.T) {
	flag.Parse()
	*noppgoroutines = true
	*stateHandler = true
	*prefixCheck = true

	config := "tests/data/juniper-junos/config/jtisim-interfaces-file-list-sig.json"
	total := 3 // this file has three workers config

	err := GetConfigFiles(configFiles, config)
	if err != nil {
		t.Errorf("config parsing error: %s", err)
	}

	// create and start workers group with three workers
	workers := NewJWorkers(*configFiles, config, 0)
	workers.StartWorkers()

	// run them for 4 seconds
	time.Sleep(time.Duration(4) * time.Second)

	// we should have three workers
	if len(workers.m) != total {
		t.Errorf("workers does not match: want %d got %d", total, len(workers.m))
	}

	// change fileList
	workers.fileList = "tests/data/juniper-junos/config/jtisim-interfaces-file-list-sig-delete.json"
	total = 1               // two deleted so we end up with only one worker config
	workers.SIGHUPWorkers() // send sighup

	// run updated workers for six seconds
	time.Sleep(time.Duration(6) * time.Second)
	// we should have only one worker
	if len(workers.m) != total {
		t.Errorf("after sighup, workers does not match: want %d got %d", total, len(workers.m))
	}

	// change file list again, bring back all three workers again
	workers.fileList = "tests/data/juniper-junos/config/jtisim-interfaces-file-list-sig.json"
	total = 3 // another sighup, we should have three workers now
	workers.SIGHUPWorkers()

	// run the stest for 4 seconds. this includes two new workers and one existing worker
	time.Sleep(time.Duration(4) * time.Second)
	if len(workers.m) != total {
		t.Errorf("after sighup, workers does not match: want %d got %d", total, len(workers.m))
	}

	workers.EndWorkers() // end all workers, this will retain workers map for us to peek data for test
	workers.Wait()       // this should not block as we have stopped all workers

	w := workers.m
	tests := []struct {
		config  string
		totalIn uint64
	}{
		{
			config:  "tests/data/juniper-junos/config/jtisim-interfaces-1.json",
			totalIn: 40, // after 2nd sighup its new worker
		},
		{
			config:  "tests/data/juniper-junos/config/jtisim-interfaces-2.json",
			totalIn: 40, // after 2nd sighup its new worker
		},
		{
			config:  "tests/data/juniper-junos/config/jtisim-interfaces-3.json",
			totalIn: 80, // this worker was running from beginning (it got two sighup though)
		},
	}

	for _, test := range tests {
		if v, ok := w[test.config]; !ok {
			t.Errorf("workers map is missing entry for worker %s", test.config)
		} else {
			if v.jctx.stats.totalIn != test.totalIn {
				t.Errorf("totalIn mismatch for %s, want %d got %d", test.config, test.totalIn, v.jctx.stats.totalIn)
			}
		}
	}
}

func TestJTISIMSigInt(t *testing.T) {
	flag.Parse()
	*noppgoroutines = true
	*stateHandler = true
	*prefixCheck = true

	config := "tests/data/juniper-junos/config/jtisim-interfaces-file-list-sig.json"
	total := 3
	runTime := 12           // if you change runTime, you will have to change totalIn and totalKV as well
	totalIn := uint64(80)   // for 12 seconds
	totalKV := uint64(3960) // for 12 seconds

	err := GetConfigFiles(configFiles, config)
	if err != nil {
		t.Errorf("config parsing error: %s", err)
	}

	workers := NewJWorkers(*configFiles, config, 0)
	workers.StartWorkers()

	if len(workers.m) != total {
		t.Errorf("workers does not match: want %d got %d", total, len(workers.m))
	}

	time.Sleep(time.Duration(runTime) * time.Second)
	workers.EndWorkers()
	workers.Wait() // this should not block as we have stopped all workers

	for _, w := range workers.m {
		jctx := w.jctx
		if totalIn != jctx.stats.totalIn {
			t.Errorf("totalIn failed for config : %s wanted %v got %v", jctx.file, totalIn, jctx.stats.totalIn)
		}
		if totalKV != jctx.stats.totalKV {
			t.Errorf("totalKV failed for config : %s wanted %v got %v", jctx.file, totalKV, jctx.stats.totalKV)
		}
	}
}

func TestJTISIMRetrySigInt(t *testing.T) {
	flag.Parse()
	*noppgoroutines = true
	*stateHandler = true
	*prefixCheck = true

	config := "tests/data/juniper-junos/config/jtisim-interfaces-file-list-sig-int.json"
	total := 1
	runTime := 10
	retryTime := 10

	err := GetConfigFiles(configFiles, config)
	if err != nil {
		t.Errorf("config parsing error: %s", err)
	}

	workers := NewJWorkers(*configFiles, config, 0)
	workers.StartWorkers()

	if len(workers.m) != total {
		t.Errorf("workers does not match: want %d got %d", total, len(workers.m))
	}

	time.Sleep(time.Duration(runTime) * time.Second)
	workers.EndWorkers()
	// Wait for the signal to be consumed as the worker may be already waiting
	// in retry Timer
	time.Sleep(time.Duration(retryTime+5) * time.Second)

	for _, w := range workers.m {
		jctx := w.jctx
		// Check the the interrupt in the control channel is received.
		select {
		case s := <-jctx.control:
			switch s {
			case os.Interrupt:
				t.Errorf("Interrupt signal is not handled at connection for %s", jctx.file)
			}
		default:
		}
	}
}

func TestPrometheus(t *testing.T) {
	flag.Parse()
	tests := []struct {
		name   string
		config string
		total  int
		maxRun int64
	}{
		{
			name:   "influx-1",
			config: "tests/data/juniper-junos/config/jtisim-prometheus.json",
			maxRun: 6,
			total:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host := "127.0.0.1"
			port := 8090

			*noppgoroutines = true
			*stateHandler = true
			*prom = true
			exporter = promInit()

			defer func() {
				*prom = false
				exporter = nil
			}()

			configFiles = &[]string{test.config}

			err := GetConfigFiles(configFiles, "")
			if err != nil {
				t.Errorf("config parsing error: %s", err)
			}

			workers := NewJWorkers(*configFiles, test.config, test.maxRun)
			workers.StartWorkers()
			workers.Wait()

			if len(workers.m) != test.total {
				t.Errorf("workers does not match: want %d got %d", test.total, len(workers.m))
			}

			if len(workers.m) != 1 {
				t.Errorf("cant't run prometheus test for more than one worker")
			}

			if w, ok := workers.m[test.config]; !ok {
				t.Errorf("could not found worker for config %s", test.config)
			} else {
				jctx := w.jctx
				if err := prometheusCollect(host, port, jctx); err != nil {
					t.Errorf("%v", err)
				} else {
					if jctx.testRes, err = os.Open(test.config + ".testres"); err != nil {
						t.Errorf("could not open %s", test.config+".testres")
					} else {
						if err := compareResults(jctx); err != nil {
							t.Errorf("%v", err)
						}
						jctx.testRes.Close()
					}
				}
			}
		})
	}
}
func TestInflux(t *testing.T) {
	flag.Parse()
	host := "127.0.0.1"
	port := 50052
	tests := []struct {
		name   string
		config string
		total  int
		maxRun int64
	}{
		{
			name:   "influx-1",
			config: "tests/data/juniper-junos/config/jtisim-influx.json",
			maxRun: 25,
			total:  1,
		},
		{
			name:   "influx-alias",
			config: "tests/data/juniper-junos/config/jtisim-influx-alias.json",
			maxRun: 8,
			total:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			*noppgoroutines = true
			configFiles = &[]string{test.config}
			err := GetConfigFiles(configFiles, "")
			if err != nil {
				t.Errorf("config parsing error: %s", err)
			}
			if err := influxStore(host, port, STOREOPEN, test.config+".testres"); err != nil {
				t.Errorf("influxStore(open) failed: %v", err)
			}

			workers := NewJWorkers(*configFiles, test.config, test.maxRun)
			workers.StartWorkers()
			workers.Wait()
			if err := influxStore(host, port, STORECLOSE, test.config+".testres"); err != nil {
				t.Errorf("influxStore(close) failed: %v", err)
			}

			if len(workers.m) != test.total {
				t.Errorf("workers does not match: want %d got %d", test.total, len(workers.m))
			}

			for _, w := range workers.m {
				jctx := w.jctx
				if jctx.testRes, err = os.Open(test.config + ".testres"); err != nil {
					t.Errorf("could not open %s", test.config+".testres")
				} else {
					if err := compareResults(jctx); err != nil {
						t.Errorf("%v", err)
					}
					jctx.testRes.Close()
				}
			}
		})
	}
}
func TestJTISIMMaxRun(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		total       int
		maxRun      int64
		totalIn     uint64
		totalKV     uint64
		compression string
	}{
		{
			name:        "multi-file-list-1",
			config:      "tests/data/juniper-junos/config/jtisim-interfaces-file-list.json",
			total:       2,
			maxRun:      25,   // if you change maxRun, please change totalIn and totalKV as well
			totalIn:     120,  // for 25 seconds
			totalKV:     5940, // for 25 seconds
			compression: "gzip",
		},
		{
			name:        "multi-file-list-1",
			config:      "tests/data/juniper-junos/config/jtisim-interfaces-file-list.json",
			total:       2,
			maxRun:      25,   // if you change maxRun, please change totalIn and totalKV as well
			totalIn:     120,  // for 25 seconds
			totalKV:     5940, // for 25 seconds
			compression: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name+test.compression, func(t *testing.T) {
			flag.Parse()
			*noppgoroutines = true
			*stateHandler = true
			*prefixCheck = true
			*compression = test.compression

			err := GetConfigFiles(configFiles, test.config)
			if err != nil {
				t.Errorf("config parsing error: %s", err)
			}

			workers := NewJWorkers(*configFiles, test.config, test.maxRun)
			workers.StartWorkers()
			workers.Wait()

			if len(workers.m) != test.total {
				t.Errorf("workers does not match: want %d got %d", test.total, len(workers.m))
			}

			for _, w := range workers.m {
				jctx := w.jctx
				if test.totalIn != jctx.stats.totalIn {
					t.Errorf("totalIn failed for config : %s wanted %v got %v", jctx.file, test.totalIn, jctx.stats.totalIn)
				}
				if test.totalKV != jctx.stats.totalKV {
					t.Errorf("totalKV failed for config : %s wanted %v got %v", jctx.file, test.totalKV, jctx.stats.totalKV)
				}
				switch test.compression {
				case "gzip":
					if jctx.stats.totalInPayloadWireLength >= jctx.stats.totalInPayloadLength {
						t.Errorf("gzip compression failed: totalInPayloadWireLength = %v totalInPayloadLength = %v",
							jctx.stats.totalInPayloadWireLength, jctx.stats.totalInPayloadLength)
					}
				case "":
					if jctx.stats.totalInPayloadWireLength != jctx.stats.totalInPayloadLength {
						t.Errorf("no compression failed: totalInPayloadWireLength = %v totalInPayloadLength = %v",
							jctx.stats.totalInPayloadWireLength, jctx.stats.totalInPayloadLength)
					}

				}
			}
		})
	}
}

func TestVMXTagsPoints(t *testing.T) {
	flag.Parse()
	*conTestData = true
	*noppgoroutines = true

	tests := []struct {
		name   string
		config string
		jctx   *JCtx
	}{
		{
			name:   "interfaces",
			config: "tests/data/juniper-junos/config/interfaces.json",
			jctx: &JCtx{
				file: "tests/data/juniper-junos/config/interfaces.json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jctx := test.jctx
			err := ConfigRead(jctx, true, nil)
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}

			sizeFileContent, err := ioutil.ReadFile(jctx.file + ".testmeta")
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}

			data, err := os.Open(jctx.file + ".testbytes")
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}
			defer data.Close()

			testRes, err := os.Create(jctx.file + ".testres")
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}
			defer testRes.Close()
			jctx.testRes = testRes

			sizes := strings.Split(string(sizeFileContent), ":")
			for _, size := range sizes {
				if size != "" {
					n, err := strconv.ParseInt(size, 10, 64)
					if err != nil {
						t.Errorf("error %v for test config %s", err, test.config)
					}
					d := make([]byte, n)
					bytesRead, err := data.Read(d)
					if err != nil {
						t.Errorf("error %v for test config %s", err, test.config)
					}
					if int64(bytesRead) != n {
						t.Errorf("want %d got %d from testbytes", n, bytesRead)
					}
					ocData := new(na_pb.OpenConfigData)
					err = proto.Unmarshal(d, ocData)

					if err != nil {
						t.Errorf("error %v for test config %s", err, test.config)
					}
					addIDB(ocData, jctx, time.Now())
				}
			}
			if err := compareResults(jctx); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}
