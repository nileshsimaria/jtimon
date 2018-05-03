package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

var (
	// DefaultBatchSize to use if user has not provided in the config
	DefaultBatchSize = 1 * 1024 * 1024
	//DefaultBatchFreq is 2 seconds
	DefaultBatchFreq = 2
)

type iFluxCtx struct {
	sync.Mutex
	influxc *client.Client
	batchCh chan *client.Point
}

type influxCfg struct {
	Server         string
	Port           int
	Dbname         string
	User           string
	Password       string
	Recreate       bool
	Measurement    string
	Diet           bool
	BatchSize      int
	BatchFrequency int
}

type timeDiff struct {
	field string
	tags  map[string]string
}

func setupBatchWriteIDB(jctx *JCtx) {
	batchSize := jctx.cfg.Influx.BatchSize
	if batchSize == 0 {
		batchSize = DefaultBatchSize
	}

	batchCh := make(chan *client.Point, batchSize)
	jctx.iFlux.batchCh = batchCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.cfg.Influx.BatchFrequency
	if bFreq == 0 {
		bFreq = DefaultBatchFreq
	}

	l(true, jctx, fmt.Sprintln("batch-size :", batchSize, "batch-freq", bFreq))
	ticker := time.NewTicker(time.Duration(bFreq) * time.Second)
	go func() {
		for range ticker.C {
			n := len(batchCh)
			l(true, jctx, fmt.Sprintln("Total available points for batching: ", n))
			if n != 0 {
				bp, err := client.NewBatchPoints(client.BatchPointsConfig{
					Database:  jctx.cfg.Influx.Dbname,
					Precision: "us",
				})
				if err != nil {
					log.Fatal(err)
				}

				for i := 0; i < n; i++ {
					bp.AddPoint(<-batchCh)
				}

				if err := (*jctx.iFlux.influxc).Write(bp); err != nil {
					log.Fatal(err)
				}
				l(true, jctx, fmt.Sprintln("Batch write sucessful! Post batch write available points: ", len(batchCh)))
			}
		}
	}()
}

// Takes in XML path with predicates and returns list of tags+values
// along with a final XML path without predicates
func spitTagsNPath(xmlpath string) (string, map[string]string) {
	re := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*)\\=([^\\[]*)\\]")
	subs := re.FindAllStringSubmatch(xmlpath, -1)
	tags := make(map[string]string)

	// Given XML path, this will spit out final path without predicates
	if len(subs) > 0 {
		for _, sub := range subs {
			tagKey := strings.Split(xmlpath, sub[0])[0]
			tagKey += "/" + strings.TrimSpace(sub[1]) + "/@" + strings.TrimSpace(sub[2])
			tagValue := strings.Replace(sub[3], "'", "", -1)

			tags[tagKey] = tagValue
			xmlpath = strings.Replace(xmlpath, sub[0], "/"+strings.TrimSpace(sub[1]), 1)
		}
	}

	return xmlpath, tags
}

func mName(ocData *na_pb.OpenConfigData, cfg Config) string {
	if cfg.Influx.Measurement != "" {
		return cfg.Influx.Measurement
	}
	if ocData != nil {
		return ocData.SystemId
	}
	return ""
}

// A go routine to add header of gRPC in to influxDB
func addGRPCHeader(jctx *JCtx, hmap map[string]interface{}) {
	cfg := jctx.cfg
	jctx.iFlux.Lock()
	defer jctx.iFlux.Unlock()

	if jctx.iFlux.influxc == nil {
		return
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  cfg.Influx.Dbname,
		Precision: "us",
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(hmap) != 0 {
		m := mName(nil, jctx.cfg)
		m = fmt.Sprintf("%s-%d-HDR", m, jctx.idx)
		tags := make(map[string]string)
		pt, err := client.NewPoint(m, tags, hmap, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)
		if err := (*jctx.iFlux.influxc).Write(bp); err != nil {
			log.Fatal(err)
		}
	}
}

// A go routine to add summary of stats collection in to influxDB
func addIDBSummary(jctx *JCtx, stmap map[string]interface{}) {
	cfg := jctx.cfg
	jctx.iFlux.Lock()
	defer jctx.iFlux.Unlock()

	if jctx.iFlux.influxc == nil {
		return
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  cfg.Influx.Dbname,
		Precision: "us",
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(stmap) != 0 {
		m := mName(nil, jctx.cfg)
		m = fmt.Sprintf("%s-%d-LOG", m, jctx.idx)
		tags := make(map[string]string)
		pt, err := client.NewPoint(m, tags, stmap, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)
		if err := (*jctx.iFlux.influxc).Write(bp); err != nil {
			log.Fatal(err)
		}
	}
}

// A go routine to add one telemetry packet in to InfluxDB
func addIDB(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	cfg := jctx.cfg
	if jctx.iFlux.influxc == nil {
		return
	}

	prefix := ""
	for _, v := range ocData.Kv {
		kv := make(map[string]interface{})
		if *stateHandler {
			kv["platency"] = rtime.UnixNano()/1000000 - int64(ocData.Timestamp)
			if v.Key == "__timestamp__" {
				if rtime.UnixNano()/1000000 < int64(v.GetUintValue()) {
					kv["elatency"] = 0
				} else {
					kv["elatency"] = rtime.UnixNano()/1000000 - int64(v.GetUintValue())
				}
				kv["ilatency"] = int64(v.GetUintValue()) - int64(ocData.Timestamp)
			}
			if v.Key == "__agentd_rx_timestamp__" {
				kv["arxlatency"] = int64(v.GetUintValue()) - int64(ocData.Timestamp)
			}
			if v.Key == "__agentd_tx_timestamp__" {
				kv["atxlatency"] = int64(v.GetUintValue()) - int64(ocData.Timestamp)
			}
		}

		if v.Key == "__prefix__" {
			prefix = v.GetStrValue()
		}

		key := v.Key
		if strings.HasPrefix(key, "/") == false {
			key = prefix + v.Key
		}

		xmlpath, tags := spitTagsNPath(key)
		tags["device"] = cfg.Host
		tags["sensor"] = ocData.Path
		kv["sequence_number"] = float64(ocData.SequenceNumber)
		kv["component_id"] = ocData.ComponentId

		if cfg.Influx.Diet == false {
			switch v.Value.(type) {
			case *na_pb.KeyValue_StrValue:
				if val, err := strconv.ParseInt(v.GetStrValue(), 10, 64); err == nil {
					kv[xmlpath+"-int"] = val
				} else {
					kv[xmlpath] = v.GetStrValue()
				}
				break
			case *na_pb.KeyValue_DoubleValue:
				kv[xmlpath+"-float"] = float64(v.GetDoubleValue())
				break
			case *na_pb.KeyValue_IntValue:
				kv[xmlpath+"-float"] = float64(v.GetIntValue())
				break
			case *na_pb.KeyValue_UintValue:
				kv[xmlpath+"-float"] = float64(v.GetUintValue())
				break
			case *na_pb.KeyValue_SintValue:
				kv[xmlpath+"-float"] = float64(v.GetSintValue())
				break
			case *na_pb.KeyValue_BoolValue:
				kv[xmlpath+"-bool"] = v.GetBoolValue()
				break
			case *na_pb.KeyValue_BytesValue:
				kv[xmlpath+"-bytes"] = v.GetBytesValue()
				break
			default:
			}
		}

		if len(kv) != 0 {
			pt, err := client.NewPoint(mName(ocData, jctx.cfg), tags, kv, rtime)
			if err != nil {
				log.Fatal(err)
			}
			if jctx.cfg.Log.Verbose {
				l(true, jctx, fmt.Sprintln(pt.String()))
			}
			jctx.iFlux.batchCh <- pt
		}
	}
}

func getInfluxClient(cfg Config) *client.Client {
	if cfg.Influx.Server == "" {
		return nil
	}
	addr := fmt.Sprintf("http://%v:%v", cfg.Influx.Server, cfg.Influx.Port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: cfg.Influx.User,
		Password: cfg.Influx.Password,
	})

	if err != nil {
		log.Fatal(err)
	}
	return &c
}

func queryIDB(clnt client.Client, cmd string, db string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func influxInit(jctx *JCtx) {
	cfg := jctx.cfg
	c := getInfluxClient(cfg)

	if cfg.Influx.Server != "" && cfg.Influx.Recreate == true && c != nil {
		_, err := queryIDB(*c, fmt.Sprintf("DROP DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
		if err != nil {
			log.Fatal(err)
		}
		_, err = queryIDB(*c, fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
		if err != nil {
			log.Fatal(err)
		}
	}
	jctx.iFlux.influxc = c
	if cfg.Influx.Server != "" && c != nil {
		setupBatchWriteIDB(jctx)
	}

	l(true, jctx, "Successfully initialized InfluxDB Client")
}
