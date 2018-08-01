package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// ConfigFileList to get the list of config file names
type ConfigFileList struct {
	Filenames []string `json:"config_file_list"`
}

// Config struct
type Config struct {
	Port     int           `json:"port"`
	Host     string        `json:"host"`
	User     string        `json:"user"`
	Password string        `json:"password"`
	CID      string        `json:"cid"`
	Meta     bool          `json:"meta"`
	EOS      bool          `json:"eos"`
	API      APIConfig     `json:"api"`
	GRPC     GRPCConfig    `json:"grpc"`
	TLS      TLSConfig     `json:"tls"`
	Influx   InfluxConfig  `json:"influx"`
	Paths    []PathsConfig `json:"paths"`
	Log      LogConfig     `json:"log"`
	// mux      sync.RWMutex
}

//LogConfig is config struct for logging
type LogConfig struct {
	File          string `json:"file"`
	PeriodicStats int    `json:"periodic-stats"`
	Verbose       bool   `json:"verbose"`
	DropCheck     bool   `json:"drop-check"`
	LatencyCheck  bool   `json:"latency-check"`
	CSVStats      bool   `json:"csv-stats"`
	FileHandle    *os.File
	Logger        *log.Logger
}

// APIConfig is config struct for API Server
type APIConfig struct {
	Port int `json:"port"`
}

//GRPCConfig is to specify GRPC params
type GRPCConfig struct {
	WS int32 `json:"ws"`
}

// TLSConfig is to specify TLS params
type TLSConfig struct {
	ClientCrt  string `json:"clientcrt"`
	ClientKey  string `json:"clientkey"`
	CA         string `json:"ca"`
	ServerName string `json:"servername"`
}

// PathsConfig to specify subscription path, reporting-interval (freq), etc,.
type PathsConfig struct {
	Path string `json:"path"`
	Freq uint64 `json:"freq"`
	Mode string `json:"mode"`
}

// NewJTIMONConfigFilelist to return configfilelist object
func NewJTIMONConfigFilelist(file string) (ConfigFileList, error) {
	// Parse config file
	configfilelist, err := ParseJSONConfigFileList(file)
	return configfilelist, err
}

// NewJTIMONConfig to return config object
func NewJTIMONConfig(file string) (Config, error) {
	// parse config file
	config, err := ParseJSON(file)
	return config, err
}

func fillupDefaults(config *Config) {
	// fill up defaults
	if config.GRPC.WS == 0 {
		config.GRPC.WS = DefaultGRPCWindowSize
	}
	if config.Influx.BatchFrequency == 0 {
		config.Influx.BatchFrequency = DefaultIDBBatchFreq
	}
	if config.Influx.BatchSize == 0 {
		config.Influx.BatchSize = DefaultIDBBatchSize
	}
}

// ParseJSONConfigFileList parses JSON encoded string of JTIMON Config files
func ParseJSONConfigFileList(file string) (ConfigFileList, error) {
	var configfilelist ConfigFileList

	f, err := ioutil.ReadFile(file)
	if err != nil {
		return configfilelist, err
	}

	if err := json.Unmarshal(f, &configfilelist); err != nil {
		return configfilelist, err
	}

	return configfilelist, err
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

	fillupDefaults(&config)

	if _, err := ValidateConfig(config); err != nil {
		log.Fatalf("Invalid config %v\n", err)
	}

	return config, nil
}

// ValidateConfig for config validation
func ValidateConfig(config Config) (string, error) {
	b, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return "", err
	}
	return string(b), nil

}

// ExploreConfig of JTIMON
func ExploreConfig() (string, error) {
	var config Config
	c := "{\"paths\": [{}]}"

	if err := json.Unmarshal([]byte(c), &config); err == nil {
		if b, err := json.MarshalIndent(config, "", "    "); err == nil {
			return string(b), nil
		}
	}
	return "", errors.New("Something is very wrong - This should have not happened")
}

// IsVerboseLogging returns true if verbose logging is enabled, false otherwise
func IsVerboseLogging(jctx *JCtx) bool {
	return jctx.config.Log.Verbose
}

// GetConfigFiles to get the list of config files
func GetConfigFiles(cfgFile *[]string, cfgFileList *string) error {
	if len(*cfgFileList) != 0 {
		configfilelist, err := NewJTIMONConfigFilelist(*cfgFileList)
		if err != nil {
			return fmt.Errorf("Error %v in %v", err, cfgFileList)
		}
		n := len(configfilelist.Filenames)
		if n == 0 {

			return fmt.Errorf("File list doesn't have any files in %v", *cfgFileList)
		}
		*cfgFile = configfilelist.Filenames
	} else {
		n := len(*cfgFile)
		if n == 0 {
			return fmt.Errorf("Can not run without any config file")
		}
	}
	return nil
}
