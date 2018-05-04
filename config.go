package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config struct
type Config struct {
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	User     string    `json:"user"`
	Password string    `json:"password"`
	Meta     bool      `json:"meta"`
	Eos      bool      `json:"eos"`
	Cid      string    `json:"cid"`
	API      api       `json:"api"`
	Grpc     grpccfg   `json:"grpc"`
	TLS      tlscfg    `json:"tls"`
	Influx   influxCfg `json:"influx"`
	Paths    []spath   `json:"paths"`
	Log      logT      `json:"log"`
}

type logT struct {
	File          string `json:"file"`
	Verbose       bool   `json:"verbose"`
	PeriodicStats int    `json:"periodic-stats"`
	DropCheck     bool   `json:"drop-check"`
	LatencyCheck  bool   `json:"latency-check"`
	CSVStats      bool   `json:"csv-stats"`
	handle        *os.File
	loger         *log.Logger
}

type api struct {
	Port int `json:"port"`
}

type grpccfg struct {
	Ws int32 `json:"ws"`
}

type tlscfg struct {
	ClientCrt  string `json:"clientcrt"`
	ClientKey  string `json:"clientkey"`
	CA         string `json:"ca"`
	ServerName string `json:"servername"`
}

type spath struct {
	Path string `json:"path"`
	Freq uint64 `json:"freq"`
	Mode string `json:"mode"`
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
