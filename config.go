package main

import (
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
	Meta     bool
	Eos      bool
	Cid      string
	API      api
	Grpc     grpccfg
	TLS      tlscfg
	Influx   *influxCfg
	Paths    []spath
	CStats   statsT
	Log      logT
}

type logT struct {
	LogFileName string
	FileHandle  *os.File
}

type statsT struct {
	pStats   int64
	csvStats bool
}

type api struct {
	Port int
}

type grpccfg struct {
	Ws int32
}

type tlscfg struct {
	ClientCrt  string
	ClientKey  string
	CA         string
	ServerName string
}

type spath struct {
	Path string
	Freq uint64
	Mode string
}

func configInit(cfgFile string) config {
	// parse config file
	cfg := parseJSON(cfgFile)

	//logJSON(cfg)

	return cfg
}

func configValidation(jctx *jcontext) {
	if jctx.cfg.CStats.csvStats {
		if *dcheck {
			if jctx.cfg.Log.LogFileName == "" {
				log.Fatalf("Can't use --drop-check when cvs data log file is not used")
			}
		}
	}
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
	emitLog(fmt.Sprintf("API-Port: %v\n", cfg.API.Port))
	emitLog(fmt.Sprintf("gRPC window-size: %v\n", cfg.Grpc.Ws))

	emitLog(fmt.Sprintf("TLS Client-CRT: %v\n", cfg.TLS.ClientCrt))
	emitLog(fmt.Sprintf("TLS Client-KEY: %v\n", cfg.TLS.ClientKey))
	emitLog(fmt.Sprintf("TLS CA: %v\n", cfg.TLS.CA))
	emitLog(fmt.Sprintf("TLS Server-Name: %v\n", cfg.TLS.ServerName))

	for i := range cfg.Paths {
		emitLog(fmt.Sprintf("Path: %v Freq: %v Subscription-Mode: %v\n", cfg.Paths[i].Path, cfg.Paths[i].Freq, cfg.Paths[i].Mode))
	}
	if cfg.Influx != nil {
		emitLog(fmt.Sprintf("Server : %v\n", cfg.Influx.Server))
		emitLog(fmt.Sprintf("Port: %v\n", cfg.Influx.Port))
		emitLog(fmt.Sprintf("DBName: %v\n", cfg.Influx.Dbname))
		emitLog(fmt.Sprintf("Measurement: %v\n", cfg.Influx.Measurement))
		emitLog(fmt.Sprintf("Recreate DB: %v\n", cfg.Influx.Recreate))
		emitLog(fmt.Sprintf("User: %v\n", cfg.Influx.User))
		emitLog(fmt.Sprintf("Password: %v\n", cfg.Influx.Password))
		emitLog(fmt.Sprintf("Flat Schema: %v\n", cfg.Influx.Flat))
		emitLog(fmt.Sprintf("Diet Influx: %v\n", cfg.Influx.Diet))
	}
}
