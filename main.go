package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	flag "github.com/spf13/pflag"
)

const (
	// DefaultGRPCWindowSize is the default GRPC Window Size
	DefaultGRPCWindowSize = 1048576
	// MatchExpressionXpath is for the pattern matching the xpath and key-value pairs
	MatchExpressionXpath = "\\/([^\\/]*)\\[(.*?)+?(?:\\])"
	// MatchExpressionKey is for pattern matching the single and multiple key value pairs
	MatchExpressionKey = "([A-Za-z0-9-/]*)=(.*?)?(?:and|$)+"
)

var (
	configFiles    = flag.StringSlice("config", make([]string, 0), "Config file name(s)")
	configFileList = flag.String("config-file-list", "", "List of Config files")
	aliasFile      = flag.String("alias-file", "", "File containing aliasing information")
	expConfig      = flag.Bool("explore-config", false, "Explore full config of JTIMON and exit")
	print          = flag.Bool("print", false, "Print Telemetry data")
	outJSON        = flag.Bool("json", false, "Convert telemetry packet into JSON")
	logMux         = flag.Bool("log-mux-stdout", false, "All logs to stdout")
	maxRun         = flag.Int64("max-run", 0, "Max run time in seconds")
	stateHandler   = flag.Bool("stats-handler", false, "Use GRPC statshandler")
	versionOnly    = flag.Bool("version", false, "Print version and build-time of the binary and exit")
	compression    = flag.String("compression", "", "Enable HTTP/2 compression (gzip, deflate)")
	prom           = flag.Bool("prometheus", false, "Stats for prometheus monitoring system")
	promPort       = flag.Int32("prometheus-port", 8090, "Prometheus port")
	prefixCheck    = flag.Bool("prefix-check", false, "Report missing __prefix__ in telemetry packet")
	pProf          = flag.Bool("pprof", false, "Profile JTIMON")
	pProfPort      = flag.Int32("pprof-port", 6060, "Profile port")
	noppgoroutines = flag.Bool("no-per-packet-goroutines", false, "Spawn per packet go routines")
	genTestData    = flag.Bool("generate-test-data", false, "Generate test data")
	conTestData    = flag.Bool("consume-test-data", false, "Consume test data")

	jtimonVersion = "version-not-available"
	buildTime     = "build-time-not-available"

	exporter *jtimonPExporter
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

	err := GetConfigFiles(configFiles, *configFileList)
	if err != nil {
		log.Printf("config parsing error: %s", err)
		return
	}

	if *aliasFile != "" {
		aliasInit()
	}

	workers := NewJWorkers(*configFiles, *configFileList, *maxRun)
	workers.StartWorkers()
	workers.Wait()

	log.Printf("All done ... exiting!\n")
}
