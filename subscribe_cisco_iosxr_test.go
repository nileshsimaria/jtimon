package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nileshsimaria/jtimon/multi-vendor/cisco/iosxr/telemetry-proto"
	flag "github.com/spf13/pflag"
)

func TestTransformPath(t *testing.T) {
	tests := []struct {
		name   string
		setEnv bool
		input  string
		output string
	}{
		{
			name:   "noenv1",
			setEnv: false,
			input:  "/interfaces",
			output: "/interfaces",
		},
		{
			name:   "env1",
			setEnv: true,
			input:  "/interfaces",
			output: "hbot_interfaces",
		},
		{
			name:   "noenv2",
			setEnv: false,
			input:  "/interfaces/interfaces",
			output: "/interfaces/interfaces",
		},
		{
			name:   "env2",
			setEnv: true,
			input:  "/interfaces/interface",
			output: "hbot_interfaces_interface",
		},
		{
			name:   "env3",
			setEnv: true,
			input:  "/interfaces/interface[name=\"xe-0/0/0\"]/state/mtu",
			output: "hbot_interfaces_interface_name__xe_0_0_0___state_mtu",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setEnv {
				os.Setenv("MV_CISCO_IOSXR_XFORM_PATH", "yes")
				defer os.Unsetenv("MV_CISCO_IOSXR_XFORM_PATH")
			}
			op := transformPath(test.input)
			if op != test.output {
				t.Errorf("got: %s, want: %s", op, test.output)
			}
		})
	}
}

func TestXRInflux(t *testing.T) {
	host := "127.0.0.1"
	port := 50052

	tt := []struct {
		name   string
		config string
		jctx   *JCtx
	}{
		{
			name:   "xr-all",
			config: "tests/data/cisco-ios-xr/config/xr-all-influx.json",
			jctx: &JCtx{
				file: "tests/data/cisco-ios-xr/config/xr-all-influx.json",
			},
		},
		{
			name:   "xr-wdsysmon",
			config: "tests/data/cisco-ios-xr/config/xr-wdsysmon-influx.json",
			jctx: &JCtx{
				file: "tests/data/cisco-ios-xr/config/xr-wdsysmon-influx.json",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			if err := influxStore(host, port, STOREOPEN, test.config+".testres"); err != nil {
				t.Errorf("influxStore(open) failed: %v", err)
			}

			jctx := test.jctx
			err := ConfigRead(jctx, true, nil)
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}

			schema, err := getXRSchema(jctx)
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
					message := new(telemetry.Telemetry)
					err = proto.Unmarshal(d, message)
					if err != nil {
						t.Errorf("error %v for test config %s", err, test.config)
					}
					path := message.GetEncodingPath()
					if path == "" {
						continue
					}

					ePath := strings.Split(path, "/")
					if len(ePath) == 1 {
						for _, nodes := range schema.nodes {
							for _, node := range nodes {
								if strings.Compare(ePath[0], node.Name) == 0 {
									for _, fields := range message.GetDataGpbkv() {
										parentPath := []string{node.Name}
										processTopLevelMsg(jctx, node, fields, parentPath)
									}
								}
							}
						}
					} else if len(ePath) >= 2 {
						for _, nodes := range schema.nodes {
							for _, node := range nodes {
								if strings.Compare(ePath[0], node.Name) == 0 {
									processMultiLevelMsg(jctx, node, ePath, message)
								}
							}
						}

					}
				}
			}

			// we will need to give LPServer some time to process all the points
			time.Sleep(time.Duration(8) * time.Second)
			if err := influxStore(host, port, STORECLOSE, test.config+".testres"); err != nil {
				t.Errorf("influxStore(close) failed: %v", err)
			}

			if err := compareResults(jctx); err != nil {
				t.Log(err)
			}
		})
	}
}

func TestXRTagsPoints(t *testing.T) {
	flag.Parse()
	*conTestData = true

	tt := []struct {
		name   string
		config string
		jctx   *JCtx
	}{
		{
			name:   "xr-all",
			config: "tests/data/cisco-ios-xr/config/xr-all.json",
			jctx: &JCtx{
				file: "tests/data/cisco-ios-xr/config/xr-all.json",
			},
		},
		{
			name:   "xr-wdsysmon",
			config: "tests/data/cisco-ios-xr/config/xr-wdsysmon.json",
			jctx: &JCtx{
				file: "tests/data/cisco-ios-xr/config/xr-wdsysmon.json",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			jctx := test.jctx
			err := ConfigRead(jctx, true, nil)
			if err != nil {
				t.Errorf("error %v for test config %s", err, test.config)
			}

			schema, err := getXRSchema(jctx)
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
					message := new(telemetry.Telemetry)
					err = proto.Unmarshal(d, message)
					if err != nil {
						t.Errorf("error %v for test config %s", err, test.config)
					}
					path := message.GetEncodingPath()
					if path == "" {
						continue
					}

					ePath := strings.Split(path, "/")
					if len(ePath) == 1 {
						for _, nodes := range schema.nodes {
							for _, node := range nodes {
								if strings.Compare(ePath[0], node.Name) == 0 {
									for _, fields := range message.GetDataGpbkv() {
										parentPath := []string{node.Name}
										processTopLevelMsg(jctx, node, fields, parentPath)
									}
								}
							}
						}
					} else if len(ePath) >= 2 {
						for _, nodes := range schema.nodes {
							for _, node := range nodes {
								if strings.Compare(ePath[0], node.Name) == 0 {
									processMultiLevelMsg(jctx, node, ePath, message)
								}
							}
						}

					}
				}

			}
			if err := compareResults(jctx); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}
func TestXRSchema(t *testing.T) {

	tt := []struct {
		name       string
		schemaPath string
		schemaStr  string
	}{
		{
			name:       "directory",
			schemaPath: "tests/data/cisco-ios-xr/schema",
			schemaStr: `openconfig-bgp:bgp
				neighbors
					neighbor
						neighbor-address[key]
						afi-safis
							afi-safi
								afi-safi-name[key]
			openconfig-rib-bgp:bgp-rib
				afi-safis
					afi-safi-name[key]
					afi-safi
						afi-safi-name[key]
						ipv4-unicast
							neighbors
								neighbor
									neighbor-address[key]
			openconfig-interfaces:interfaces
				interface
					name[key]
					subinterfaces
						subinterface
							index[key]
			Cisco-IOS-XR-infra-statsd-oper:infra-statistics
				interfaces
					interface
						interface-name[key]
						protocols
							protocol
								protocol-name[key]
						cache
							protocols
								protocol
									protocol-name[key]
						total
							protocols
								protocol
									protocol-name[key]
						latest
							protocols
								protocol
									protocol-name[key]		
			Cisco-IOS-XR-wdsysmon-fd-oper:system-monitoring
				cpu-utilization
					node-name[key]
					process-cpu
						process-name[key]`,
		},
		{
			name:       "file",
			schemaPath: "tests/data/cisco-ios-xr/schema/interfaces.json",
			schemaStr: `openconfig-interfaces:interfaces
				interface
					name[key]
					subinterfaces
						subinterface
							index[key]`,
		},
		{
			name:       "env",
			schemaPath: "",
			schemaStr: `openconfig-interfaces:interfaces
				interface
					name[key]
					subinterfaces
						subinterface
							index[key]`,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			var jctx *JCtx
			if test.schemaPath == "" {
				os.Setenv("MV_CISCO_IOSXR_SCHEMA", "tests/data/cisco-ios-xr/schema/interfaces.json")
				defer os.Unsetenv("MV_CISCO_IOSXR_SCHEMA")
				jctx = &JCtx{
					config: Config{
						Vendor: VendorConfig{
							Name:     "cisco-iosxr",
							RemoveNS: true,
						},
					},
				}
			} else {
				jctx = &JCtx{
					config: Config{
						Vendor: VendorConfig{
							Name:     "cisco-iosxr",
							RemoveNS: true,
							Schema: []VendorSchema{
								{test.schemaPath},
							},
						},
					},
				}
			}

			if schema, err := getXRSchema(jctx); err != nil {
				t.Errorf("error %v for %s", err, test.schemaPath)
			} else {
				got := fmt.Sprintf("%s", schema)
				if compareString(got, test.schemaStr) == false {
					t.Errorf("want: \n%s\n, got: \n%s\n", test.schemaStr, got)
				}
			}
		})
	}
}
