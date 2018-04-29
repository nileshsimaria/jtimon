package main

import (
	"encoding/json"
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
	File          string
	Verbose       bool
	PeriodicStats int  `json:"periodic-stats"`
	DropCheck     bool `json:"drop-check"`
	LatencyCheck  bool `json:"latency-check"`
	handle        *os.File
	loger         *log.Logger
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

func configInit(cfgFile string) (config, error) {
	// parse config file
	cfg, err := parseJSON(cfgFile)
	cfg.CStats.csvStats = *csvStats

	return cfg, err
}

func parseJSON(cfgFile string) (config, error) {
	var cfg config

	file, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(file, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
