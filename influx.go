package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type iFluxCtx struct {
	sync.Mutex
	influxc *client.Client
	tdm     map[string]timeDiff
}

type influxCfg struct {
	Server      string
	Port        int
	Dbname      string
	User        string
	Password    string
	Recreate    bool
	Measurement string
}

type timeDiff struct {
	field string
	tags  map[string]string
}

func addTimeDiff(jctx *jcontext, sensor string, tags map[string]string, field string) {
	jctx.iFlux.Lock()
	defer jctx.iFlux.Unlock()

	if jctx.iFlux.tdm == nil {
		jctx.iFlux.tdm = make(map[string]timeDiff)
	}

	_, ok := jctx.iFlux.tdm[sensor]
	if ok == false {
		jctx.iFlux.tdm[sensor] = timeDiff{field, tags}
		fmt.Printf("tdd-sensor: %s\n", sensor)
		fmt.Printf("tdd-field : %s\n", field)
		for tn, tv := range tags {
			fmt.Printf("tdd-tag: name: %s value: %s\n", tn, tv)
		}
	}
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

func getMeasurementName(ocData *na_pb.OpenConfigData, cfg config) string {
	if cfg.Influx.Measurement != "" {
		return cfg.Influx.Measurement
	}
	return ocData.SystemId
}

// A go routine to add one telemetry packet in to InfluxDB
func addIDB(ocData *na_pb.OpenConfigData, jctx *jcontext, rtime time.Time) {
	cfg := jctx.cfg

	jctx.iFlux.Lock()
	defer jctx.iFlux.Unlock()
	prefix := ""

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  cfg.Influx.Dbname,
		Precision: "us",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range ocData.Kv {
		kv := make(map[string]interface{})
		kv["platency"] = rtime.UnixNano()/1000000 - int64(ocData.Timestamp)
		if v.Key == "__timestamp__" {
			if rtime.UnixNano()/1000000 < int64(v.GetUintValue()) {
				kv["elatency"] = 0
			} else {
				kv["elatency"] = rtime.UnixNano()/1000000 - int64(v.GetUintValue())
			}
			kv["ilatency"] = int64(v.GetUintValue()) - int64(ocData.Timestamp)
			//fmt.Printf("ilatency: %v\n", kv["ilatency"])
		}

		if v.Key == "__prefix__" {
			prefix = v.GetStrValue()
		}

		key := v.Key
		if strings.HasPrefix(key, "/") == false {
			key = prefix + v.Key
		}

		xmlpath, tags := spitTagsNPath(key)
		if *td == true {
			if strings.HasPrefix(v.Key, "__") == false {
				addTimeDiff(jctx, ocData.Path, tags, xmlpath)
			}
		}
		tags["device"] = cfg.Host
		tags["sensor"] = ocData.Path

		switch v.Value.(type) {
		case *na_pb.KeyValue_StrValue:
			if val, err := strconv.ParseInt(v.GetStrValue(), 10, 64); err == nil {
				kv[xmlpath+"-int"] = val
			} else {
				kv[xmlpath] = v.GetStrValue()
			}
			break
		case *na_pb.KeyValue_DoubleValue:
			kv[xmlpath] = v.GetDoubleValue()
			break
		case *na_pb.KeyValue_IntValue:
			kv[xmlpath] = float64(v.GetIntValue())
			break
		case *na_pb.KeyValue_UintValue:
			kv[xmlpath] = float64(v.GetUintValue())
			break
		case *na_pb.KeyValue_SintValue:
			kv[xmlpath] = float64(v.GetSintValue())
			break
		case *na_pb.KeyValue_BoolValue:
			kv[xmlpath] = v.GetBoolValue()
			break
		case *na_pb.KeyValue_BytesValue:
			kv[xmlpath] = v.GetBytesValue()
			break
		default:
		}

		if len(kv) != 0 {
			pt, err := client.NewPoint(getMeasurementName(ocData, jctx.cfg), tags, kv, rtime)
			if err != nil {
				log.Fatal(err)
			}
			bp.AddPoint(pt)
		}
	}

	if err := (*jctx.iFlux.influxc).Write(bp); err != nil {
		log.Fatal(err)
	}
}

func influxDBQueryString(jctx *jcontext) {
	jctx.iFlux.Lock()
	defer jctx.iFlux.Unlock()

	fmt.Println("influxDBQueryString()")

	for sensor, timeDiff := range jctx.iFlux.tdm {
		fmt.Printf("tdd-sensor: %s\n", sensor)
		fmt.Printf("tdd-field : %s\n", timeDiff.field)
		for tn, tv := range timeDiff.tags {
			fmt.Printf("tdd-tag: name: %s value: %s\n", tn, tv)
		}
	}

	//resp, err := queryIDB(*iFlux.influxc, fmt.Sprintf("DROP DATABASE %s", cfg.Influx.Dbname), cfg.Influx.Dbname)
	//fmt.Printf("%v\n", resp)

}
func getInfluxClient(cfg config) *client.Client {
	if cfg.Influx == nil {
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

func influxInit(cfg config) *client.Client {
	c := getInfluxClient(cfg)

	if cfg.Influx != nil && cfg.Influx.Recreate == true && c != nil {
		_, err := queryIDB(*c, fmt.Sprintf("DROP DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
		if err != nil {
			log.Fatal(err)
		}
		_, err = queryIDB(*c, fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
		if err != nil {
			log.Fatal(err)
		}
	}
	return c
}
