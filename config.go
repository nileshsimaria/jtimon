package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config struct
type Config struct {
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
	Influx   influxCfg
	Paths    []spath
	Log      logT
}

type logT struct {
	File          string
	Verbose       bool
	PeriodicStats int  `json:"periodic-stats"`
	DropCheck     bool `json:"drop-check"`
	LatencyCheck  bool `json:"latency-check"`
	CSVStats      bool `json:"csv-stats"`
	handle        *os.File
	loger         *log.Logger
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

// NewJTIMONConfig to return config object
func NewJTIMONConfig(file string) (Config, error) {
	// parse config file
	config, err := ParseJSON(file)
	return config, err
}

// ParseJSON parses JSON encoded config of JTIMON
func ParseJSON(file string) (Config, error) {
	var config Config

	f, err := ioutil.ReadFile(file)
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(f, &config); err != nil {
		return config, err
	}
	return config, nil
}
