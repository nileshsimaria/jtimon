package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"bytes"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb/client/v2"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	lp "github.com/influxdata/line-protocol"
)

// InfluxCtx is run time info of InfluxDB data structures
type InfluxCtx struct {
	sync.Mutex
	influxClient      *client.Client
	batchWCh          chan []*client.Point
	batchWMCh         chan *batchWMData
	influx2Client     *influxdb2.Client
	influx2WriteAPI   *api.WriteAPIBlocking
	influx2BatchWCh   chan []*write.Point
	influx2BatchWMCh  chan *influx2BatchWMData
	influx2RecordWCh  chan string
	influx2RecordWMCh chan *influx2RecordWMData
	accumulatorCh     chan (*metricIDB)
	reXpath, reKey    *regexp.Regexp
}

type batchWMData struct {
	measurement string
	points      []*client.Point
}

type influx2BatchWMData struct {
	measurement string
	points      []*write.Point
}

type influx2RecordWMData struct {
	measurement string
	records     string
}

// InfluxConfig is the config of InfluxDB
type InfluxConfig struct {
	Server               string        `json:"server"`
	Port                 int           `json:"port"`
	Dbname               string        `json:"dbname"`
	User                 string        `json:"user"`
	Password             string        `json:"password"`
	Influx2              Influx2Config `json:"influx2"`
	Recreate             bool          `json:"recreate"`
	Measurement          string        `json:"measurement"`
	BatchSize            int           `json:"batchsize"`
	BatchFrequency       int           `json:"batchfrequency"`
	HTTPTimeout          int           `json:"http-timeout"`
	RetentionPolicy      string        `json:"retention-policy"`
	AccumulatorFrequency int           `json:"accumulator-frequency"`
	WritePerMeasurement  bool          `json:"write-per-measurement"`
}

type Influx2Config struct {
	Token          string `json:"token"`
	OrgName        string `json:"orgname"`
	WriteInRecords bool   `json:"writeinrecords"`
}

type metricIDB struct {
	tags   map[string]string
	fields map[string]interface{}
	ts     uint64
}

func newMetricIDB(tags map[string]string, fields map[string]interface{}, ts uint64) *metricIDB {
	return &metricIDB{
		tags:   tags,
		fields: fields,
		ts:     ts,
	}
}

func (m *metricIDB) accumulate(jctx *JCtx) {
	switch {
	case jctx.config.Influx.Influx2 == Influx2Config{}:
		if jctx.influxCtx.influxClient != nil {
			jctx.influxCtx.accumulatorCh <- m
		}
	default:
		if jctx.influxCtx.influx2Client != nil {
			jctx.influxCtx.accumulatorCh <- m
		}
	}
}

func pointAcculumator(jctx *JCtx) {
	deviceTs := gDeviceTs
	freq := jctx.config.Influx.AccumulatorFrequency
	accumulatorCh := make(chan *metricIDB, 1024*10)
	jctx.influxCtx.accumulatorCh = accumulatorCh
	jLog(jctx, fmt.Sprintln("Accumulator frequency:", freq))

	ticker := time.NewTicker(time.Duration(freq) * time.Millisecond)

	go func() {
		for range ticker.C {
			n := len(accumulatorCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintf("Accumulated points : %d\n", n))
				var lastPoint *client.Point
				var points []*client.Point
				for i := 0; i < n; i++ {
					m := <-accumulatorCh

					// validate the point
					if pt, err := client.NewPoint("tmpmeasure", m.tags, m.fields, time.Now()); err != nil {
						jLog(jctx, fmt.Sprintf("pointAcculumator: Could not get TmpPoint: %v\n", err))
						continue
					} else {
						_, err := pt.Fields()
						if err != nil {
							jLog(jctx, fmt.Sprintf("addIDB: Could not get fields of the TmpPoint: %v\n", err))
							continue
						}
					}

					if lastPoint == nil {
						mName := ""
						if jctx.config.Influx.Measurement != "" {
							mName = jctx.config.Influx.Measurement
						} else {
							mName = m.tags["sensor"]
						}

						m.fields[deviceTs] = int64(m.ts)
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
							m.fields[deviceTs] = int64(m.ts)
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

				if len(points) > 0 {
					// See if we need to add lastPoint we are processing
					if eq := reflect.DeepEqual(points[len(points)-1], lastPoint); !eq {
						points = append(points, lastPoint)
					}
				}

				if len(points) > 0 {
					bp, err := client.NewBatchPoints(client.BatchPointsConfig{
						Database:        jctx.config.Influx.Dbname,
						Precision:       "ns",
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

func dbBatchWriteM(jctx *JCtx) {
	if jctx.influxCtx.influxClient == nil {
		return
	}

	batchSize := jctx.config.Influx.BatchSize
	batchMCh := make(chan *batchWMData, batchSize/4)
	jctx.influxCtx.batchWMCh = batchMCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			m := map[string][]*batchWMData{}
			n := len(batchMCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintln("#elements in the batchMCh channel : ", n))
				for i := 0; i < n; i++ {
					d := <-batchMCh
					v := m[d.measurement]
					m[d.measurement] = append(v, d)
				}
				jLog(jctx, fmt.Sprintln("#elements in the measurement map : ", len(m)))

			}

			for measurement, data := range m {
				jLog(jctx, fmt.Sprintf("measurement: %s, data len: %d", measurement, len(data)))

				bp, err := client.NewBatchPoints(client.BatchPointsConfig{
					Database:        jctx.config.Influx.Dbname,
					Precision:       "ns",
					RetentionPolicy: jctx.config.Influx.RetentionPolicy,
				})

				if err != nil {
					jLog(jctx, fmt.Sprintf("NewBatchPoints failed, error: %v", err))
					continue
				}

				for j := 0; j < len(data); j++ {
					packet := data[j].points
					k := 0
					for k = 0; k < len(packet); k++ {
						bp.AddPoint(packet[k])
						if len(bp.Points()) >= batchSize {
							jLog(jctx, fmt.Sprintf("Attempt to write %d points in %s", len(bp.Points()), measurement))
							if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
								jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s: %v", measurement, err))
							} else {
								jLog(jctx, fmt.Sprintln("Batch write successful for measurement: ", measurement))
							}

							bp, err = client.NewBatchPoints(client.BatchPointsConfig{
								Database:        jctx.config.Influx.Dbname,
								Precision:       "ns",
								RetentionPolicy: jctx.config.Influx.RetentionPolicy,
							})
						}
					}
				}
				if len(bp.Points()) > 0 {
					jLog(jctx, fmt.Sprintf("Attempt to write %d points in %s", len(bp.Points()), measurement))
					if err := (*jctx.influxCtx.influxClient).Write(bp); err != nil {
						jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s: %v", measurement, err))
					} else {
						jLog(jctx, fmt.Sprintln("Batch write successful for measurement: ", measurement))
					}

					bp, err = client.NewBatchPoints(client.BatchPointsConfig{
						Database:        jctx.config.Influx.Dbname,
						Precision:       "ns",
						RetentionPolicy: jctx.config.Influx.RetentionPolicy,
					})
				}
			}
		}
	}()
}

func dbBatchWrite(jctx *JCtx) {
	if jctx.influxCtx.influxClient == nil {
		return
	}

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
					Precision:       "ns",
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

func influx2PointAcculumator(jctx *JCtx) {
	deviceTs := gDeviceTs
	freq := jctx.config.Influx.AccumulatorFrequency
	accumulatorCh := make(chan *metricIDB, 1024*10)
	jctx.influxCtx.accumulatorCh = accumulatorCh
	jLog(jctx, fmt.Sprintln("Accumulator frequency:", freq))

	ticker := time.NewTicker(time.Duration(freq) * time.Millisecond)

	go func() {
		for range ticker.C {
			n := len(accumulatorCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintf("Accumulated points : %d\n", n))
				var lastPoint *write.Point
				var bp []*write.Point
				for i := 0; i < n; i++ {
					m := <-accumulatorCh

					// validate the point
					pt := influxdb2.NewPoint("tmpmeasure", m.tags, m.fields, time.Now())
					if pt == nil {
						jLog(jctx, fmt.Sprintf("influx2PointAcculumator: Could not get TmpPoint\n"))
						continue
					} else if len(pt.FieldList()) == 0 {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get fields of the TmpPoint\n"))
						continue
					}

					if lastPoint == nil {
						mName := ""
						if jctx.config.Influx.Measurement != "" {
							mName = jctx.config.Influx.Measurement
						} else {
							mName = m.tags["sensor"]
						}

						m.fields[deviceTs] = int64(m.ts)
						pt := influxdb2.NewPoint(mName, m.tags, m.fields, time.Now())
						if pt == nil {
							jLog(jctx, fmt.Sprintf("influx2PointAcculumator: Could not get NewPoint (first point)\n"))
							continue
						}
						lastPoint = pt
					} else {
						// let's see if we can merge
						var fieldFound = false
						Tags := make(map[string]string, len(lastPoint.TagList()))
						for _, tag := range lastPoint.TagList() {
							Tags[tag.Key] = tag.Value
						}
						eq := reflect.DeepEqual(m.tags, Tags)
						lastKV := make(map[string]interface{}, len(lastPoint.FieldList()))
						for _, field := range lastPoint.FieldList() {
							lastKV[field.Key] = field.Value
						}
						if eq {
							// tags are equal so most likely we will be able to merge.
							// we would also need to see if the field is not already part of the point,
							// if it is then we can merge because in 'config false' world of yang, keys
							// are optional inside list so instead of losing the point we'd  not merge.
							for mk := range m.fields {
								if _, ok := lastKV[mk]; ok {
									fieldFound = true
									break
								}
							}
						}

						if eq && !fieldFound {
							// We can merge
							name := lastPoint.Name()
							if lastKV == nil {
								jLog(jctx, fmt.Sprintf("addIDB: Could not get fields of the last point\n"))
								continue
							}
							// get the fields from last point for merging
							for k, v := range lastKV {
								m.fields[k] = v
							}
							pt := influxdb2.NewPoint(name, m.tags, m.fields, time.Now())
							if pt == nil {
								jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint (merging)\n"))
								continue
							}
							lastPoint = pt
						} else {
							// lastPoint tags and current point tags differes so we can not merge.
							// toss current point into the slice (points) and handle current point
							// by creating new *influxdb2.Point
							mName := ""
							if jctx.config.Influx.Measurement != "" {
								mName = jctx.config.Influx.Measurement
							} else {
								mName = m.tags["sensor"]
							}
							m.fields[deviceTs] = int64(m.ts)
							pt := influxdb2.NewPoint(mName, m.tags, m.fields, time.Now())
							if pt == nil {
								jLog(jctx, fmt.Sprintf("influx2PointAcculumator: Could not get NewPoint (first point)\n"))
								continue
							}
							bp = append(bp, lastPoint)
							lastPoint = pt
						}

						if len(bp) > 0 {
							// See if we need to add lastPoint we are processing
							if eq := reflect.DeepEqual(bp[len(bp)-1], lastPoint); !eq {
								bp = append(bp, lastPoint)
							}
						}

						if len(bp) > 0 {
							if !jctx.config.Influx.Influx2.WriteInRecords {
								for _, p := range bp {
									if jctx.config.Log.Verbose {
										jLog(jctx, fmt.Sprintf("\n\nPoint Name = %s\n", p.Name()))
										jLog(jctx, fmt.Sprintf("tags are following ...."))
										for _, tag := range p.TagList() {
											jLog(jctx, fmt.Sprintf("%s = %s", tag.Key, tag.Value))
										}
										jLog(jctx, fmt.Sprintf("fields are following ...."))
										for _, field := range p.FieldList() {
											jLog(jctx, fmt.Sprintf("%s = %v", field.Key, field.Value))
										}
									}
								}
								
								if err := (*jctx.influxCtx.influx2WriteAPI).WritePoint(context.Background(), bp...); err != nil {
									jLog(jctx, fmt.Sprintf("Batch DB write failed: %v", err))
								} else {
									jLog(jctx, fmt.Sprintln("Batch write successful! Number of points written post merge logic: ", len(bp)))
								}
							} else {
								lines, err := EncodePoints(time.Nanosecond, bp...)
								if err != nil {
									jLog(jctx, fmt.Sprintf("addIDB: Could not get batch records, %v)", err))
									continue
								}
								if jctx.config.Log.Verbose {
									jLog(jctx, fmt.Sprintf("Records are following ....\n", lines))
								}

								if err = (*jctx.influxCtx.influx2WriteAPI).WriteRecord(context.Background(), lines); err != nil {
									jLog(jctx, fmt.Sprintf("Batch records DB write failed: %v", err))
								} else {
									jLog(jctx, fmt.Sprintln("Batch records write successful! Number of records written post merge logic: ", len(bp)))
								}
							}
						}
					}
				}
			}
		}
	}()
}

func influx2DbBatchWriteM(jctx *JCtx) {
	if jctx.influxCtx.influx2Client == nil {
		return
	}

	batchSize := jctx.config.Influx.BatchSize
	batchMCh := make(chan *influx2BatchWMData, batchSize/4)
	jctx.influxCtx.influx2BatchWMCh = batchMCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(batchMCh)
			m := map[string][]*influx2BatchWMData{}
			if n != 0 {
				jLog(jctx, fmt.Sprintln("#elements in the batchMCh channel : ", n))
				for i := 0; i < n; i++ {
					d := <-batchMCh
					v := m[d.measurement]
					m[d.measurement] = append(v, d)
				}
				jLog(jctx, fmt.Sprintln("#elements in the measurement map : ", len(m)))
			}

			for measurement, data := range m {
				jLog(jctx, fmt.Sprintf("measurement: %s, data len: %d", measurement, len(data)))

				var bp []*write.Point
				for j := 0; j < len(data); j++ {
					packet := data[j].points
					k := 0
					for k = 0; k < len(packet); k++ {
						bp = append(bp, packet[k])
						if len(bp) >= batchSize {
							jLog(jctx, fmt.Sprintf("Attempt to write %d points in %s", len(bp), measurement))
							if err := (*jctx.influxCtx.influx2WriteAPI).WritePoint(context.Background(), bp...); err != nil {
								jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s: %v", measurement, err))
							} else {
								jLog(jctx, fmt.Sprintln("Batch write successful for measurement: ", measurement))
							}
							bp = bp[:0]
						}
					}
				}
				if len(bp) > 0 {
					jLog(jctx, fmt.Sprintf("Attempt to write %d points in %s", len(bp), measurement))
					if err := (*jctx.influxCtx.influx2WriteAPI).WritePoint(context.Background(), bp...); err != nil {
						jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s: %v", measurement, err))
					} else {
						jLog(jctx, fmt.Sprintln("Batch write successful for measurement: ", measurement))
					}
				}
			}
		}
	}()
}

func influx2DbBatchWrite(jctx *JCtx) {
	if jctx.influxCtx.influx2Client == nil {
		return
	}

	batchSize := jctx.config.Influx.BatchSize
	batchCh := make(chan []*write.Point, batchSize)
	jctx.influxCtx.influx2BatchWCh = batchCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(batchCh)
			if n != 0 {
				var bp []*write.Point
				for i := 0; i < n; i++ {
					packet := <-batchCh
					for j := 0; j < len(packet); j++ {
						bp = append(bp, packet[j])
					}
				}

				jLog(jctx, fmt.Sprintf("Batch processing: #packets:%d #points:%d\n", n, len(bp)))

				if err := (*jctx.influxCtx.influx2WriteAPI).WritePoint(context.Background(), bp...); err != nil {
					jLog(jctx, fmt.Sprintf("Batch DB write failed: %v", err))
				} else {
					jLog(jctx, fmt.Sprintln("Batch write successful! Post batch write available points: ", len(batchCh)))
				}
			}
		}
	}()
}

func influx2DbRecordWriteM(jctx *JCtx) {
	if jctx.influxCtx.influx2Client == nil {
		return
	}

	batchSize := jctx.config.Influx.BatchSize
	batchMCh := make(chan *influx2RecordWMData, batchSize/4)
	jctx.influxCtx.influx2RecordWMCh = batchMCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			m := map[string][]*influx2RecordWMData{}
			n := len(batchMCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintln("#elements in the batchMCh channel : ", n))
				for i := 0; i < n; i++ {
					d := <-batchMCh
					v := m[d.measurement]
					m[d.measurement] = append(v, d)
				}
				jLog(jctx, fmt.Sprintln("#elements in the measurement map : ", len(m)))

			}

			for measurement, data := range m {
				jLog(jctx, fmt.Sprintf("measurement: %s, data len: %d", measurement, len(data)))

				var sb strings.Builder
				var cnt int
				for j := 0; j < len(data); j++ {
					packet := data[j].records
					sb.WriteString(packet)
					cnt += 1
					if cnt >= batchSize {
						jLog(jctx, fmt.Sprintf("Attempt to write %d records in %s", cnt, measurement))
						if err := (*jctx.influxCtx.influx2WriteAPI).WriteRecord(context.Background(), fmt.Sprintf(sb.String())); err != nil {
							jLog(jctx, fmt.Sprintf("Batch records DB write failed for measurement %s: %v", measurement, err))
						} else {
							jLog(jctx, fmt.Sprintln("Batch records write successful for measurement: ", measurement))
						}
						sb.Reset()
						cnt = 0
					}
				}
				if cnt > 0 {
					jLog(jctx, fmt.Sprintf("Attempt to write %d records in %s", cnt, measurement))
					if err := (*jctx.influxCtx.influx2WriteAPI).WriteRecord(context.Background(), fmt.Sprintf(sb.String())); err != nil {
						jLog(jctx, fmt.Sprintf("Batch records DB write failed for measurement %s: %v", measurement, err))
					} else {
						jLog(jctx, fmt.Sprintln("Batch records write successful for measurement: ", measurement))
					}
				}
			}
		}
	}()
}

func influx2DbRecordWrite(jctx *JCtx) {
	if jctx.influxCtx.influx2Client == nil {
		return
	}

	batchSize := jctx.config.Influx.BatchSize
	batchCh := make(chan string, batchSize)
	jctx.influxCtx.influx2RecordWCh = batchCh

	// wake up periodically and perform batch write into InfluxDB
	bFreq := jctx.config.Influx.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(batchCh)
			if n != 0 {
				var sb strings.Builder
				for i := 0; i < n; i++ {
					packet := <-batchCh
					sb.WriteString(packet)
				}

				jLog(jctx, fmt.Sprintf("Batch processing: #packets:%d\n", n))
				
				if err := (*jctx.influxCtx.influx2WriteAPI).WriteRecord(context.Background(), fmt.Sprintf(sb.String())); err != nil {
					jLog(jctx, fmt.Sprintf("Batch records DB write failed: %v", err))
				} else {
					jLog(jctx, fmt.Sprintln("Batch records write successful! Post batch write available points: ", len(batchCh)))
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
					tags[getAlias(jctx.alias, tagKey)] = tagValue
				}
			}
			// Remove the key value pairs from the given xpath
			xmlpath = strings.Replace(xmlpath, sub[0], "/"+strings.TrimSpace(sub[1]), 1)
			xmlpath = getAlias(jctx.alias, xmlpath)
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

type row struct {
	tags   map[string]string
	fields map[string]interface{}
}

func newRow(tags map[string]string, fields map[string]interface{}) (*row, error) {
	return &row{
		tags:   tags,
		fields: fields,
	}, nil
}

// A go routine to add one telemetry packet in to InfluxDB
func addIDB(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	deviceTs := gDeviceTs
	cfg := jctx.config

	prefix := ""
	prefixXmlpath := ""
	var prefixTags map[string]string
	var tags map[string]string
	var xmlpath string
	prefixTags = nil

	rows := make([]*row, 0)

	for _, v := range ocData.Kv {
		kv := make(map[string]interface{})

		switch {
		case v.Key == "__prefix__":
			prefix = v.GetStrValue()
			prefixXmlpath, prefixTags = spitTagsNPath(jctx, prefix)
			continue
		case strings.HasPrefix(v.Key, "__"):
			continue
		}

		key := v.Key
		if key[0] != '/' {
			if strings.Contains(key, "[") {
				key = prefix + v.Key
				xmlpath, tags = spitTagsNPath(jctx, key)
			} else {
				xmlpath = prefixXmlpath + key
				tags = prefixTags
				xmlpath = getAlias(jctx.alias, xmlpath)
			}
		} else {
			xmlpath, tags = spitTagsNPath(jctx, key)
		}

		if tags == nil {
			continue
		}

		tags["device"] = cfg.Host
		tags["sensor"] = ocData.Path

		switch v.Value.(type) {
		case *na_pb.KeyValue_StrValue:
			kv[xmlpath] = v.GetStrValue()
		case *na_pb.KeyValue_DoubleValue:
			var floatVal float64
			val := v.GetDoubleValue()
			checkAndCeilFloatValues(nil, &val, &floatVal)
			kv[xmlpath] = floatVal
		case *na_pb.KeyValue_IntValue:
			kv[xmlpath] = float64(v.GetIntValue())
		case *na_pb.KeyValue_UintValue:
			if jctx.config.EnableUintSupport {
				kv[xmlpath] = v.GetUintValue()
			} else {
				kv[xmlpath] = float64(v.GetUintValue())
			}
		case *na_pb.KeyValue_SintValue:
			kv[xmlpath] = float64(v.GetSintValue())
		case *na_pb.KeyValue_BoolValue:
			kv[xmlpath] = v.GetBoolValue()
		case *na_pb.KeyValue_BytesValue:
			kv[xmlpath] = v.GetBytesValue()
		case *na_pb.KeyValue_FloatValue:
			var floatVal float64
			value32 := v.GetFloatValue()
			checkAndCeilFloatValues(&value32, nil, &floatVal)
			kv[xmlpath] = floatVal
		default:
		}

		if *genTestData {
			testDataPoints(jctx, GENTESTEXPDATA, tags, kv)
		}
		if *conTestData {
			testDataPoints(jctx, GENTESTRESDATA, tags, kv)
		}

		switch {
		case jctx.config.Influx.Influx2 == Influx2Config{}:
			if jctx.influxCtx.influxClient == nil {
				continue
			}
		default:
			if jctx.influxCtx.influx2Client == nil {
				continue
			}
		}

		if len(kv) != 0 {
			if len(rows) != 0 {
				lastRow := rows[len(rows)-1]
				eq := reflect.DeepEqual(tags, lastRow.tags)
				if eq {
					// We can merge
					for k, v := range kv {
						lastRow.fields[k] = v
					}
				} else {
					// Could not merge as tags are different
					kv[deviceTs] = int64(ocData.Timestamp)
					rw, err := newRow(tags, kv)
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get NewRow (no merge): %v", err))
						continue
					}
					rows = append(rows, rw)
				}
			} else {
				// First row for this sensor
				kv[deviceTs] = int64(ocData.Timestamp)
				rw, err := newRow(tags, kv)
				if err != nil {
					jLog(jctx, fmt.Sprintf("addIDB: Could not get NewRow (first row): %v", err))
					continue
				}
				rows = append(rows, rw)
			}
		}
	}
	switch {
	case cfg.Influx.Influx2 == Influx2Config{}:
		points := make([]*client.Point, 0)
		if len(rows) > 0 {
			for _, row := range rows {
				pt, err := client.NewPoint(mName(ocData, jctx.config), row.tags, row.fields, rtime)
				if err != nil {
					jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint : %v", err))
					continue
				}
				points = append(points, pt)
			}
		}

		if len(points) > 0 {
			if jctx.config.Influx.WritePerMeasurement {
				jctx.influxCtx.batchWMCh <- &batchWMData{
					measurement: mName(ocData, jctx.config),
					points:      points,
				}
			} else {
				jctx.influxCtx.batchWCh <- points
			}

			if IsVerboseLogging(jctx) {
				jLog(jctx, fmt.Sprintf("Sending %d points to batch channel for path: %s\n", len(points), ocData.Path))
				for i := 0; i < len(points); i++ {
					jLog(jctx, fmt.Sprintf("Tags: %+v\n", points[i].Tags()))
					if f, err := points[i].Fields(); err == nil {
						jLog(jctx, fmt.Sprintf("KVs : %+v\n", f))
					}
				}
			}
		}
	default:
		if !cfg.Influx.Influx2.WriteInRecords {
			points := make([]*write.Point, 0)
			if len(rows) > 0 {
				for _, row := range rows {
					pt := influxdb2.NewPoint(mName(ocData, jctx.config), row.tags, row.fields, rtime)
					if pt == nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get NewPoint"))
						continue
					}
					points = append(points, pt)
				}
			}

			if len(points) > 0 {
				if jctx.config.Influx.WritePerMeasurement {
					jctx.influxCtx.influx2BatchWMCh <- &influx2BatchWMData{
						measurement: mName(ocData, jctx.config),
						points:      points,
					}
				} else {
					jctx.influxCtx.influx2BatchWCh <- points
				}

				if IsVerboseLogging(jctx) {
					jLog(jctx, fmt.Sprintf("Sending %d points to batch channel for path: %s\n", len(points), ocData.Path))
					for i := 0; i < len(points); i++ {
						Tags := make(map[string]string, len(points[i].TagList()))
						for _, tag := range points[i].TagList() {
							Tags[tag.Key] = tag.Value
						}
						jLog(jctx, fmt.Sprintf("Tags: %+v\n", Tags))
						KVs := make(map[string]interface{}, len(points[i].FieldList()))
						for _, field := range points[i].FieldList() {
							KVs[field.Key] = field.Value
						}
						jLog(jctx, fmt.Sprintf("KVs : %+v\n", KVs))
					}
				}
			}
		} else {
			var sb strings.Builder
			var cnt int
			if len(rows) > 0 {
				for _, row := range rows {
					pt := influxdb2.NewPoint(mName(ocData, jctx.config), row.tags, row.fields, rtime)
					line, err := EncodePoints(time.Nanosecond, pt)
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get new record, %v)", err))
						continue
					}
					sb.WriteString(line)
					cnt += 1
				}
			}

			lines := fmt.Sprintf(sb.String())
			if cnt > 0 {
				if jctx.config.Influx.WritePerMeasurement {
					jctx.influxCtx.influx2RecordWMCh <- &influx2RecordWMData{
						measurement: mName(ocData, jctx.config),
						records:     lines,
					}
				} else {
					jctx.influxCtx.influx2RecordWCh <- lines
				}

				if IsVerboseLogging(jctx) {
					jLog(jctx, fmt.Sprintf("Sending %d records to batch channel for path: %s\n", cnt, ocData.Path))
					jLog(jctx, fmt.Sprintf("Record: %+v\n", lines))
				}
			}
		}
	}
}

func getInfluxClient(cfg Config, timeout int) interface{} {
	if cfg.Influx.Server == "" {
		return nil
	}

	// TODO: Vivek Resolve it only once and reuse the endpoint
	resolvedArr, err := net.ResolveTCPAddr("tcp", cfg.Influx.Server+":"+strconv.Itoa(cfg.Influx.Port))
	if err != nil {
		log.Printf("ResolveTCPAddr failed for %s, err: %v\n", cfg.Influx.Server+":"+strconv.Itoa(cfg.Influx.Port), err)
		return nil
	}
	addr := fmt.Sprintf("http://%v:%v", resolvedArr.IP, resolvedArr.Port)

	switch {
	case cfg.Influx.Influx2 == Influx2Config{}:
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     addr,
			Username: cfg.Influx.User,
			Password: cfg.Influx.Password,
			Timeout:  time.Duration(timeout) * time.Second,
		})

		if err != nil {
			log.Printf("Failed to get influxdb client: %v\n", err)
		}
		return &c
	default:
		c := influxdb2.NewClientWithOptions(addr, cfg.Influx.Influx2.Token,
			influxdb2.DefaultOptions().SetHTTPRequestTimeout(uint(timeout)))

		_, err := c.Health(context.Background())
		if err != nil {
			log.Printf("Failed to get healthy influxdb2 client: %v\n", err)
		}
		return &c
	}
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

func closeInfluxClient(clnt interface{}) {
	switch clnt.(type) {
	case influxdb2.Client:
		if clnt, ok := clnt.(influxdb2.Client); ok {
			clnt.Close()
		}
	case client.Client:
		if clnt, ok := clnt.(client.Client); ok {
			_ = clnt.Close()
		}
	}
}

func influxInit(jctx *JCtx) {
	cfg := jctx.config
	jLog(jctx, "invoking getInfluxClient for init")

	if c, ok := getInfluxClient(cfg, 10*cfg.Influx.HTTPTimeout).(*client.Client); ok { // high timeout for init
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

		jLog(jctx, "invoking getInfluxClient")
		if jctx.influxCtx.influxClient, ok = getInfluxClient(cfg, cfg.Influx.HTTPTimeout).(*client.Client); ok {
			jctx.influxCtx.reXpath = regexp.MustCompile(MatchExpressionXpath)
			jctx.influxCtx.reKey = regexp.MustCompile(MatchExpressionKey)
			if cfg.Influx.Server != "" && c != nil {
				if cfg.Influx.WritePerMeasurement {
					dbBatchWriteM(jctx)
				} else {
					dbBatchWrite(jctx)
				}
				pointAcculumator(jctx)
				jLog(jctx, "Successfully initialized InfluxDB Client")
			}
		}

		if c != nil {
			closeInfluxClient(*c)
		}
	}
}

func influx2Init(jctx *JCtx) {
	cfg := jctx.config
	jLog(jctx, "invoking getInfluxClient for init")

	if c, ok := getInfluxClient(cfg, 10*cfg.Influx.HTTPTimeout).(*influxdb2.Client); ok { // high timeout for init
		if cfg.Influx.Server != "" && c != nil {
			bucketsAPI := (*c).BucketsAPI()
			existingBuckets, err := bucketsAPI.GetBuckets(context.Background())
			if err != nil {
				log.Printf("influx2Init failed to find existing buckets %v\n", err)
			}
			var bucketFound = false
			for _, b := range *existingBuckets {
				if strings.Compare(b.Name, cfg.Influx.Dbname) == 0 {
					bucketFound = true
					if cfg.Influx.Recreate {
						err = bucketsAPI.DeleteBucket(context.Background(), &b)
						if err != nil {
							log.Printf("influx2Init failed to drop table %v\n", err)
						}
					}
				}
			}
			if !bucketFound || bucketFound && cfg.Influx.Recreate {
				org, err := (*c).OrganizationsAPI().FindOrganizationByName(context.Background(), cfg.Influx.Influx2.OrgName)
				if err != nil {
					log.Printf("influx2Init failed to find organization: %v\n", err)
				}
				_, err = bucketsAPI.CreateBucketWithName(context.Background(), org, cfg.Influx.Dbname)
				if err != nil {
					log.Printf("influx2Init failed to create database: %v\n", err)
				}
			}
		}

		jLog(jctx, "invoking getInfluxClient")
		if jctx.influxCtx.influx2Client, ok = getInfluxClient(cfg, cfg.Influx.HTTPTimeout).(*influxdb2.Client); ok {
			jctx.influxCtx.reXpath = regexp.MustCompile(MatchExpressionXpath)
			jctx.influxCtx.reKey = regexp.MustCompile(MatchExpressionKey)
			if cfg.Influx.Server != "" && c != nil {
				writeAPI := (*jctx.influxCtx.influx2Client).WriteAPIBlocking(jctx.config.Influx.Influx2.OrgName, jctx.config.Influx.Dbname)
				jctx.influxCtx.influx2WriteAPI = &writeAPI
				if jctx.config.Influx.Influx2.WriteInRecords {
					if cfg.Influx.WritePerMeasurement {
						influx2DbRecordWriteM(jctx)
					} else {
						influx2DbRecordWrite(jctx)
					}
				} else {
					if cfg.Influx.WritePerMeasurement {
						influx2DbBatchWriteM(jctx)
					} else {
						influx2DbBatchWrite(jctx)
					}
				}
				influx2PointAcculumator(jctx)
				jLog(jctx, "Successfully initialized InfluxDB2 Client")
			}
		}

		if c != nil {
			closeInfluxClient(*c)
		}
	}
}

func checkAndCeilFloatValues(val32 *float32, val64 *float64, cieled *float64) {
	if val32 != nil {
		*cieled = float64(*val32)
		if math.IsInf(*cieled, 1) {
			*cieled = math.MaxFloat64
		} else if math.IsInf(*cieled, -1) {
			*cieled = -math.MaxFloat64
		}
		return
	}

	if val64 == nil {
		*cieled = 0
		return
	}

	if math.IsInf(*val64, 1) {
		*cieled = math.MaxFloat64
	} else if math.IsInf(*val64, -1) {
		*cieled = -math.MaxFloat64
	} else {
		*cieled = *val64
	}

	return
}

// Borrowed from https://github.com/influxdata/influxdb-client-go/blob/master/internal/write/service.go#L302
func EncodePoints(precision time.Duration, points ...*write.Point) (string, error) {
	var buffer bytes.Buffer
	e := lp.NewEncoder(&buffer)
	e.SetFieldTypeSupport(lp.UintSupport)
	e.FailOnFieldErr(true)
	e.SetPrecision(precision)
	for _, point := range points {
		_, err := e.Encode(point)
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}