package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	Host     string
	Port     int
	User     string
	Password string
	Cid      string
	Api      api
	Grpc     grpccfg
	Influx   *influxCfg
	Paths    []spath
}

type api struct {
	Port int
}

type grpccfg struct {
	Ws int32
}

type spath struct {
	Path string
	Freq uint32
}

func configInit(cfgFile string) config {
	if cfgFile == "" {
		fmt.Printf("Enter config file name: ")
		r := bufio.NewScanner(os.Stdin)
		r.Scan()
		cfgFile = r.Text()
	}

	// parse config file
	cfg := parseJSON(cfgFile)

	logJSON(cfg)

	return cfg
}

func parseJSON(cfgFile string) config {
	var cfg config

	file, e := ioutil.ReadFile(cfgFile)
	if e != nil {
		log.Fatalf("File error: %v\n", e)
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &cfg); err != nil {
		panic(err)
	}
	return cfg
}

func logJSON(cfg config) {
	emitLog(fmt.Sprintf("Processing json config\n"))
	emitLog(fmt.Sprintf("Host: %v\n", cfg.Host))
	emitLog(fmt.Sprintf("Port: %v\n", cfg.Port))
	emitLog(fmt.Sprintf("CID:  %v\n", cfg.Cid))
	emitLog(fmt.Sprintf("API-Port: %v\n", cfg.Api.Port))
	emitLog(fmt.Sprintf("gRPC window-size: %v\n", cfg.Grpc.Ws))
	for i := range cfg.Paths {
		emitLog(fmt.Sprintf("Path: %v Freq: %v\n", cfg.Paths[i].Path, cfg.Paths[i].Freq))
	}
	if cfg.Influx != nil {
		emitLog(fmt.Sprintf("Server : %v\n", cfg.Influx.Server))
		emitLog(fmt.Sprintf("Port: %v\n", cfg.Influx.Port))
		emitLog(fmt.Sprintf("DBName: %v\n", cfg.Influx.Dbname))
		emitLog(fmt.Sprintf("Measurement: %v\n", cfg.Influx.Measurement))
		emitLog(fmt.Sprintf("Recreate DB: %v\n", cfg.Influx.Recreate))
		emitLog(fmt.Sprintf("User: %v\n", cfg.Influx.User))
		emitLog(fmt.Sprintf("Password: %v\n", cfg.Influx.Password))
	}
}
