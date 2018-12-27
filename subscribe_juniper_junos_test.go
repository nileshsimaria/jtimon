package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	flag "github.com/spf13/pflag"
)

func TestJTISIM(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		totalIn uint64
		totalKV uint64
	}{
		{
			name:    "multi-file-list-1",
			config:  "tests/data/juniper-junos/config/jtisim-interfaces-file-list.json",
			totalIn: 120,
			totalKV: 5940,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.Parse()
			*noppgoroutines = true
			*maxRun = 25
			*stateHandler = true
			*print = false

			err := GetConfigFiles(configFiles, &test.config)
			if err != nil {
				t.Errorf("config parsing error: %s", err)
			}

			var wg sync.WaitGroup
			wMap := make(map[string]*workerCtx)

			for _, file := range *configFiles {
				wg.Add(1)
				jctx, signalch, err := worker(file, &wg)
				if err != nil {
					wg.Done()
				} else {
					wMap[file] = &workerCtx{
						jctx:     jctx,
						signalch: signalch,
						err:      err,
					}
				}
			}

			// tell the workers (go routines) to actually start the work by Dialing
			// GRPC connection and send subscribe RPC
			for _, wCtx := range wMap {
				if wCtx.err == nil {
					wCtx.signalch <- syscall.SIGCONT
				}
			}

			go signalHandler(*configFileList, wMap, &wg)
			go maxRunHandler(*maxRun, wMap)

			wg.Wait()
			for _, wCtx := range wMap {
				jctx := wCtx.jctx
				if test.totalIn != jctx.stats.totalIn {
					t.Errorf("totalIn failed for config : %s wanted %v got %v", jctx.file, test.totalIn, jctx.stats.totalIn)
				}
				if test.totalKV != jctx.stats.totalKV {
					t.Errorf("totalKV failed for config : %s wanted %v got %v", jctx.file, test.totalKV, jctx.stats.totalKV)
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
			err := ConfigRead(jctx, true)
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
