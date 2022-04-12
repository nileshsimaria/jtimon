package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	flag "github.com/spf13/pflag"
)

var (
	configFiles     = flag.StringSlice("config", make([]string, 0), "Config file name(s)")
	configFileList  = flag.String("config-file-list", "", "List of Config files")
	expConfig       = flag.Bool("explore-config", false, "Explore full config of JTIMON and exit")
	print           = flag.Bool("print", false, "Print Telemetry data")
	outJSON         = flag.Bool("json", false, "Convert telemetry packet into JSON")
	logMux          = flag.Bool("log-mux-stdout", false, "All logs to stdout")
	maxRun          = flag.Int64("max-run", 0, "Max run time in seconds")
	stateHandler    = flag.Bool("stats-handler", false, "Use GRPC statshandler")
	versionOnly     = flag.Bool("version", false, "Print version and build-time of the binary and exit")
	compression     = flag.String("compression", "", "Enable HTTP/2 compression (gzip)")
	prom            = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	promHost        = flag.String("prometheus-host", "127.0.0.1", "IP to bind Prometheus service to")
	promPort        = flag.Int32("prometheus-port", 8090, "Prometheus port")
	prefixCheck     = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	pProf           = flag.Bool("pprof", false, "Profile JTIMON")
	pProfPort       = flag.Int32("pprof-port", 6060, "Profile port")
	noppgoroutines  = flag.Bool("no-per-packet-goroutines", false, "Spawn per packet go routines")
	genTestData     = flag.Bool("generate-test-data", false, "Generate test data")
	conTestData     = flag.Bool("consume-test-data", false, "Consume test data")
	dialOut         = flag.Bool("dial-out", false, "Act as server supporting dialOut connections")
	terminator      = flag.Bool("terminator", false, "Act as RPC terminator")
	myListeningIP   = flag.String("host", "0.0.0.0", "Server IP")
	myListeningPort = flag.Int("port", 32767, "The server port")
	skipVerify      = flag.Bool("skip-verify", false, "Skip verification of ssl authentication")
	myCACert        = flag.String("ca-cert", "./certs/self_signed/ca-cert.pem", "Path of CA cert")
	myCert          = flag.String("cert", "./certs/self_signed/server-cert.pem", "Path of server cert")
	myKey           = flag.String("pem", "./certs/self_signed/server-key.pem", "Path of server key")
	kafkaBroker     = flag.String("kafka-broker", "kafka:9092", "Comma seperated list of Kafka brokers each in the form ip:port")

	jtimonVersion = "version-not-available"
	buildTime     = "build-time-not-available"

	exporter *jtimonPExporter
)

const (
	gDeviceTs = "__device_timestamp__"
)

func main() {
	flag.Parse()
	if *pProf {
		go func() {
			addr := fmt.Sprintf("localhost:%d", *pProfPort)
			log.Println(http.ListenAndServe(addr, nil))
		}()
	}
	if *prom {
		exporter = promInit()
	}

	log.Printf("Version: %s BuildTime %s\n", jtimonVersion, buildTime)
	if *versionOnly {
		return
	}

	if *expConfig {
		config, err := ExploreConfig()
		if err == nil {
			log.Printf("\n%s\n", config)
		} else {
			log.Printf("can not generate config")
		}
		return
	}

	if *dialOut {
		// Server
		if *terminator { // Act as dialout rpc terminator
			// Keep listening for RPCs, this will never return back
			startDialOutServer(myListeningIP, myListeningPort)
		} else { // Act as just a consumer of data from kafka (data published by rpc terminator)
			createKafkaConsumerGroup("jtimon", []string{"gnmi-data"})
		}
	}

	if !*terminator { // If not a terminator, provide the config files to terminator
		err := GetConfigFiles(configFiles, *configFileList)
		if err != nil {
			log.Printf("config parsing error: %s", err)
			return
		}

		workers := NewJWorkers(*configFiles, *configFileList, *maxRun)
		workers.StartWorkers()
		workers.Wait()
	}

	log.Printf("all done ... exiting!")
}
