package main

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

var (
	// DefaultIDBBatchSize to use if user has not provided in the config
	DefaultIDBBatchSize = 1024 * 1024
	//DefaultIDBBatchFreq is 2 seconds
	DefaultIDBBatchFreq = 2000
	//DefaultAccumulatorFreq is 2 seconds
	DefaultIDBAccumulatorFreq = 2000
)

// InfluxCtx is run time info of InfluxDB data structures
type InfluxCtx struct {
	sync.Mutex
	influxClient   *client.Client
	batchWCh       chan []*client.Point
	accumulatorCh  chan (*metricIDB)
	reXpath, reKey *regexp.Regexp
}

// InfluxConfig is the config of InfluxDB
type InfluxConfig struct {
	Server               string `json:"server"`
	Port                 int    `json:"port"`
	Dbname               string `json:"dbname"`
	User                 string `json:"user"`
	Password             string `json:"password"`
	Recreate             bool   `json:"recreate"`
	Measurement          string `json:"measurement"`
	Diet                 bool   `json:"diet"`
	BatchSize            int    `json:"batchsize"`
	BatchFrequency       int    `json:"batchfrequency"`
	RetentionPolicy      string `json:"retention-policy"`
	AccumulatorFrequency int    `json:"accumulator-frequency"`
}

type metricIDB struct {
	tags   map[string]string
	fields map[string]interface{}
}

func newMetricIDB(tags map[string]string, fields map[string]interface{}) *metricIDB {
	return &metricIDB{
		tags:   tags,
		fields: fields,
	}
}

func (m *metricIDB) accumulate(jctx *JCtx) {
	if jctx.influxCtx.influxClient != nil {
		jctx.influxCtx.accumulatorCh <- m
	}
}

func pointAcculumator(jctx *JCtx) {
	freq := jctx.config.Influx.AccumulatorFrequency
	accumulatorCh := make(chan *metricIDB, 1024*10)
	jctx.influxCtx.accumulatorCh = accumulatorCh
	jLog(jctx, fmt.Sprintln("Accumulator frequency:", freq))

	ticker := time.NewTicker(time.Duration(freq) * time.Millisecond)

	go func() {
		for range ticker.C {
			n := len(accumulatorCh)
			jLog(jctx, fmt.Sprintf("Accumulated points : %d\n", n))
			if n != 0 {
				var lastPoint *client.Point
				var points []*client.Point
				for i := 0; i < n; i++ {
					m := <-accumulatorCh
					if lastPoint == nil {
						mName := ""
						if jctx.config.Influx.Measurement != "" {
							mName = jctx.config.Influx.Measurement
						} else {
							mName = m.tags["sensor"]
						}

						pt, err := client.NewPoint(mName, m.tags, m.fields, time.Now())
						if err != nil {
							jLog(jctx, fmt.Sprintf("pointAcculumator: Could not get NewPoint (first point): %v\n", err))
							continue
						}
						lastPoint = pt
					} else {
						// let's see if we can merge
						var fieldFound = false
						eq := reflect.DeepEqual(m.tags, lastPoint.Tags())
						if eq {
							// tags are equal so most likely we will be able to merge.
							// we would also need to see if the field is not already part of the point,
							// if it is then we can merge because in 'config false' world of yang, keys
							// are optional inside list so instead of losing the point we'd  not merge.
							for mk := range m.fields {
								lastKV, _ := lastPoint.Fields()
								if _, ok := lastKV[mk]; ok {
									fieldFound = true
									break
								}
							}
						}
						if eq && !fieldFound {
							// We can merge
							lastKV, err := lastPoint.Fields()
							name := lastPoint.Name()
							if err != nil {
								jLog(jctx, fmt.Sprintf("addIDB: Could not get fields of the last point: %v\n", err))
								continue
							}
							// get the fields from last point for merging
							for k, v := range lastKV {
								m.fields[k] = v
							}
							pt, err := client.NewPoint(name, m.tags, m.fields, time.Now())
							if err != nil {
								jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint (merging): %v\n", err))
								continue
							}
							lastPoint = pt
						} else {
							// lastPoint tags and current point tags differes so we can not merge.
							// toss current point into the slice (points) and handle current point
							// by creating new *client.Point
							mName := ""
							if jctx.config.Influx.Measurement != "" {
								mName = jctx.config.Influx.Measurement
							} else {
								mName = m.tags["sensor"]
							}
							pt, err := client.NewPoint(mName, m.tags, m.fields, time.Now())
							if err != nil {
								jLog(jctx, fmt.Sprintf("pointAcculumator: Could not get NewPoint (first point): %v\n", err))
								continue
							}
							points = append(points, lastPoint)
							lastPoint = pt
						}
					}
				}
				// See if we need to add lastPoint we are processing
				if eq := reflect.DeepEqual(points[len(points)-1], lastPoint); !eq {
					points = append(points, lastPoint)
				}

				if len(points) > 0 {
					bp, err := client.NewBatchPoints(client.BatchPointsConfig{
						Database:        jctx.config.Influx.Dbname,
						Precision:       "us",
						RetentionPolicy: jctx.config.Influx.RetentionPolicy,
					})

					if err != nil {
						jLog(jctx, fmt.Sprintf("NewBatchPoints failed, error: %v\n", err))
						return
					}

					for _, p := range points {
						bp.AddPoint(p)
						if jctx.config.Log.Verbose {
							jLog(jctx, fmt.Sprintf("\n\nPoint Name = %s\n", p.Name()))
							jLog(jctx, fmt.Sprintf("tags are following ...."))
							for k, v := range p.Tags() {
								jLog(jctx, fmt.Sprintf("%s = %s", k, v))
							}
							fields, err := p.Fields()
							if err != nil {
								jLog(jctx, fmt.Sprintf("%v", err))
							} else {
								jLog(jctx, fmt.Sprintf("fields are following ...."))
								for k, v := range fields {
									jLog(jctx, fmt.Sprintf("%s = %s", k, v))
								}
							}
						}
					}
					if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
						jLog(jctx, fmt.Sprintf("Batch DB write failed: %v", err))
					} else {
						jLog(jctx, fmt.Sprintln("Batch write successful! Number of points written post merge logic: ", len(points)))
					}

				}
			}
		}
	}()
}

func dbBatchWrite(jctx *JCtx) {
	batchSize := jctx.config.Influx.BatchSize
	batchCh := make(chan []*client.Point, batchSize)
	jctx.influxCtx.batchWCh = batchCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(batchCh)
			if n != 0 {
				bp, err := client.NewBatchPoints(client.BatchPointsConfig{
					Database:        jctx.config.Influx.Dbname,
					Precision:       "us",
					RetentionPolicy: jctx.config.Influx.RetentionPolicy,
				})

				if err != nil {
					jLog(jctx, fmt.Sprintf("NewBatchPoints failed, error: %v\n", err))
					return
				}

				for i := 0; i < n; i++ {
					packet := <-batchCh
					for j := 0; j < len(packet); j++ {
						bp.AddPoint(packet[j])
					}
				}

				jLog(jctx, fmt.Sprintf("Batch processing: #packets:%d #points:%d\n", n, len(bp.Points())))

				if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
					jLog(jctx, fmt.Sprintf("Batch DB write failed: %v", err))
				} else {
					jLog(jctx, fmt.Sprintln("Batch write successful! Post batch write available points: ", len(batchCh)))
				}
			}
		}
	}()
}

// Takes in XML path with predicates and returns list of tags+values
// along with a final XML path without predicates
func spitTagsNPath(jctx *JCtx, xmlpath string) (string, map[string]string) {
	// reXpath regex splits the given xmlpath string into element-name and its
	// keyvalue pairs
	// Example :
	// 		foo/bar/interfaces[name = ge-/0/0/0]
	//		Regex will split the string into the following groups
	//			group 0  /interfaces[name = 'ge-/0/0/0' and unit = ' 0']
	//			group 1  interfaces
	//			group 2  name = 'ge-/0/0/0' and unit = ' 0'
	subs := jctx.influxCtx.reXpath.FindAllStringSubmatch(xmlpath, -1)
	tags := make(map[string]string)

	// Given XML path, this will spit out final path without predicates
	if len(subs) > 0 {
		for _, sub := range subs {
			tagKeyPrefix := strings.Split(xmlpath, sub[0])[0]
			// From the key value pairs extract the key and value
			// the first and second group will contain the key and value
			// respectively.
			keyValues := jctx.influxCtx.reKey.FindAllStringSubmatch(sub[2], -1)

			if len(keyValues) > 0 {
				for _, keyValue := range keyValues {
					tagKey := tagKeyPrefix + "/" + strings.TrimSpace(sub[1]) +
						"/@" + strings.TrimSpace(keyValue[1])
					tagValue := strings.Replace(strings.TrimSpace(keyValue[2]), "'", "", -1)
					// Store as key value pairs
					tags[getAlias(tagKey)] = tagValue
				}
			}
			// Remove the key value pairs from the given xpath
			xmlpath = strings.Replace(xmlpath, sub[0], "/"+strings.TrimSpace(sub[1]), 1)
			xmlpath = getAlias(xmlpath)
		}
	}

	return xmlpath, tags
}

// SubscriptionPathFromPath to extract subscription path from path
func SubscriptionPathFromPath(path string) string {
	tokens := strings.Split(path, ":")
	if len(tokens) == 4 {
		return tokens[2]
	}
	return ""
}

func mName(ocData *na_pb.OpenConfigData, cfg Config) string {
	if cfg.Influx.Measurement != "" {
		return cfg.Influx.Measurement
	}

	if ocData != nil {
		path := ocData.Path
		return SubscriptionPathFromPath(path)
	}
	return ""
}

// A go routine to add header of gRPC in to influxDB
func addGRPCHeader(jctx *JCtx, hmap map[string]interface{}) {
	cfg := jctx.config

	if jctx.influxCtx.influxClient == nil {
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
		m := mName(nil, jctx.config)
		m = fmt.Sprintf("%s-%s-HDR", m, jctx.config.Host)
		tags := make(map[string]string)
		pt, err := client.NewPoint(m, tags, hmap, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)
		if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
			log.Fatal(err)
		}
	}
}

// A go routine to add summary of stats collection in to influxDB
func addIDBSummary(jctx *JCtx, stmap map[string]interface{}) {
	cfg := jctx.config

	if jctx.influxCtx.influxClient == nil {
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
		m := mName(nil, jctx.config)
		m = fmt.Sprintf("%s-%s-LOG", m, jctx.config.Host)
		tags := make(map[string]string)
		pt, err := client.NewPoint(m, tags, stmap, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)
		if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
			log.Fatal(err)
		}
	}
}

// A go routine to add one telemetry packet in to InfluxDB
func addIDB(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	cfg := jctx.config
	if jctx.influxCtx.influxClient == nil {
		return
	}

	prefix := ""
	points := make([]*client.Point, 0)

	for _, v := range ocData.Kv {
		kv := make(map[string]interface{})
		if *stateHandler && *latencyProfile {
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

		switch {
		case v.Key == "__prefix__":
			prefix = v.GetStrValue()
			continue
		case strings.HasPrefix(v.Key, "__"):
			continue
		}

		key := v.Key
		if !strings.HasPrefix(key, "/") {
			key = prefix + v.Key
		}

		xmlpath, tags := spitTagsNPath(jctx, key)
		tags["device"] = cfg.Host
		tags["sensor"] = ocData.Path

		if !cfg.Influx.Diet {
			switch v.Value.(type) {
			case *na_pb.KeyValue_StrValue:
				kv[xmlpath] = v.GetStrValue()
			case *na_pb.KeyValue_DoubleValue:
				kv[xmlpath] = v.GetDoubleValue()
			case *na_pb.KeyValue_IntValue:
				kv[xmlpath] = float64(v.GetIntValue())
			case *na_pb.KeyValue_UintValue:
				kv[xmlpath] = float64(v.GetUintValue())
			case *na_pb.KeyValue_SintValue:
				kv[xmlpath] = float64(v.GetSintValue())
			case *na_pb.KeyValue_BoolValue:
				kv[xmlpath] = v.GetBoolValue()
			case *na_pb.KeyValue_BytesValue:
				kv[xmlpath] = v.GetBytesValue()
			default:
			}
		}

		if len(kv) != 0 {
			if len(points) != 0 {
				lastPoint := points[len(points)-1]
				eq := reflect.DeepEqual(tags, lastPoint.Tags())
				if eq {
					// We can merge
					lastKV, err := lastPoint.Fields()
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get fields of the last point: %v\n", err))
						continue
					}
					// get the fields from last point for merging
					for k, v := range lastKV {
						kv[k] = v
					}
					pt, err := client.NewPoint(mName(ocData, jctx.config), tags, kv, rtime)
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint (merging): %v\n", err))
						continue
					}
					// Replace last point with new point having merged fields
					points[len(points)-1] = pt
				} else {
					// Could not merge as tags are different
					pt, err := client.NewPoint(mName(ocData, jctx.config), tags, kv, rtime)
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint (no merge): %v\n", err))
						continue
					}
					points = append(points, pt)
				}
			} else {
				// First point for this sensor
				pt, err := client.NewPoint(mName(ocData, jctx.config), tags, kv, rtime)
				if err != nil {
					jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint (first point): %v\n", err))
					continue
				}
				points = append(points, pt)
			}
		}
	}

	if len(points) > 0 {
		jctx.influxCtx.batchWCh <- points

		if IsVerboseLogging(jctx) {
			jLog(jctx, fmt.Sprintf("Sending %d points to batch channel for path: %s\n", len(points), ocData.Path))
			for i := 0; i < len(points); i++ {
				jLog(jctx, fmt.Sprintf("Tags: %+v\n", points[i].Tags()))
				if f, err := points[i].Fields(); err != nil {
					jLog(jctx, fmt.Sprintf("KVs : %+v\n", f))
				}
			}
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
		log.Printf("Failed to get influxdb client: %v\n", err)
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
	cfg := jctx.config
	c := getInfluxClient(cfg)

	if cfg.Influx.Server != "" && c != nil {
		if cfg.Influx.Recreate {
			_, err := queryIDB(*c, fmt.Sprintf("DROP DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
			if err != nil {
				log.Printf("influxInit failed to drop table %v\n", err)
			}
		}
		_, err := queryIDB(*c, fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.Influx.Dbname), cfg.Influx.Dbname)
		if err != nil {
			log.Printf("influxInit failed to create database: %v\n", err)
		}
	}

	jctx.influxCtx.influxClient = c
	jctx.influxCtx.reXpath = regexp.MustCompile(MatchExpressionXpath)
	jctx.influxCtx.reKey = regexp.MustCompile(MatchExpressionKey)
	if cfg.Influx.Server != "" && c != nil {
		dbBatchWrite(jctx)
		pointAcculumator(jctx)
		jLog(jctx, "Successfully initialized InfluxDB Client")
	}
}
