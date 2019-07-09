package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
)

// ConfigFileList to get the list of config file names
type ConfigFileList struct {
	Filenames []string `json:"config_file_list"`
}

// Config struct
type Config struct {
	Port            int           `json:"port"`
	Host            string        `json:"host"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	CID             string        `json:"cid"`
	Meta            bool          `json:"meta"`
	EOS             bool          `json:"eos"`
	GRPC            GRPCConfig    `json:"grpc"`
	TLS             TLSConfig     `json:"tls"`
	Influx          InfluxConfig  `json:"influx"`
	Paths           []PathsConfig `json:"paths"`
	Log             LogConfig     `json:"log"`
	Vendor          VendorConfig  `json:"vendor"`
	Alias           string        `json:"alias"`
	PasswordDecoder string        `json:"password-decoder"`
	DropTagPrefix	bool          `json:"dropTagPrefix"`
}

// VendorConfig definition
type VendorConfig struct {
	Name     string         `json:"name"`
	RemoveNS bool           `json:"remove-namespace"`
	Schema   []VendorSchema `json:"schema"`
}

// VendorSchema definition
type VendorSchema struct {
	Path string `json:"path"`
}

//LogConfig is config struct for logging
type LogConfig struct {
	File          string `json:"file"`
	PeriodicStats int    `json:"periodic-stats"`
	Verbose       bool   `json:"verbose"`
	out           *os.File
	logger        *log.Logger
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
	if config.Influx.HTTPTimeout == 0 {
		config.Influx.HTTPTimeout = DefaultIDBTimeout
	}
	if config.Influx.AccumulatorFrequency == 0 {
		config.Influx.AccumulatorFrequency = DefaultIDBAccumulatorFreq
	}
}

// ParseJSONConfigFileList parses file list config
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
	return "", errors.New("something is wrong, this should have not happened")
}

// IsVerboseLogging returns true if verbose logging is enabled, false otherwise
func IsVerboseLogging(jctx *JCtx) bool {
	return jctx.config.Log.Verbose
}

// GetConfigFiles to get the list of config files
func GetConfigFiles(cfgFile *[]string, cfgFileList string) error {
	if len(cfgFileList) != 0 {
		configfilelist, err := NewJTIMONConfigFilelist(cfgFileList)
		if err != nil {
			return fmt.Errorf("%v: [%v]", err, cfgFileList)
		}
		n := len(configfilelist.Filenames)
		if n == 0 {
			return fmt.Errorf("%s doesn't have any files", cfgFileList)
		}
		*cfgFile = configfilelist.Filenames
	} else {
		n := len(*cfgFile)
		if n == 0 {
			return fmt.Errorf("can not run without any config file")
		}
	}
	return nil
}

// ValidateConfigChange to check which config changes are allowed
func ValidateConfigChange(jctx *JCtx, config Config) error {
	runningCfg := jctx.config
	if !reflect.DeepEqual(runningCfg, config) {
		// config change is now only for path, it can be extended.
		if !reflect.DeepEqual(runningCfg.Paths, config.Paths) {
			return nil
		}
	}
	return fmt.Errorf("ValidateConfigChange: only paths are allowed to be changed")
}

// ConfigRead will read the config and init the services.
// In case of config changes, it will update the existing config
func ConfigRead(jctx *JCtx, init bool) error {
	var err error

	config, err := NewJTIMONConfig(jctx.file)
	if err != nil {
		log.Printf("config parsing error for %s: %v", jctx.file, err)
		return fmt.Errorf("config parsing (json unmarshal) error for %s: %v", jctx.file, err)
	}

	if init {
		jctx.config = config
		logInit(jctx)
		b, err := json.MarshalIndent(jctx.config, "", "    ")
		if err != nil {
			return fmt.Errorf("config parsing error (json marshal) for %s: %v", jctx.file, err)
		}

		jLog(jctx, fmt.Sprintf("Running config of JTIMON:\n %s", string(b)))
		// Decode the password if the config has provided the decode util
		if len(config.PasswordDecoder) > 0 {
			// Run the decode util with the input file as argument
			cmd := exec.Command(config.PasswordDecoder, jctx.file)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s:%s\n", err, errStr)
				return err
			}
			jctx.config.Password = outStr
		}
		// subscription channel (subch) is used to let go routine receiving telemetry
		// data know about certain events like sighup.
		jctx.control = make(chan os.Signal)

		go periodicStats(jctx)
		influxInit(jctx)
	} else {
		err := ValidateConfigChange(jctx, config)
		if err == nil {
			jctx.config.Paths = config.Paths
			jLog(jctx, fmt.Sprintf("config has been updated"))
		} else {
			return fmt.Errorf("no change in subscription path, ignoring config changes")
		}
	}
	return nil
}

// StringInSlice to check whether a string in in the slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
