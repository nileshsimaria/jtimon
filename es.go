package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esutil"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

// ElasticCtx is run time info of ESDB data structures
type ElasticCtx struct {
	sync.Mutex
	esClient *elasticsearch.Client
	batchWCh chan []*Article
	accumulatorCh  chan (*metricIDB)
	reXpath, reKey *regexp.Regexp
}

// EsConfig is the config of ES
type EsConfig struct {
	Server               string `json:"server"`
	Port                 int    `json:"port"`
	Idxname              string `json:"idxname"`
	User                 string `json:"user"`
	Password             string `json:"password"`
	Recreate             bool   `json:"recreate"`
	Measurement          string `json:"measurement"`
	BatchSize            int    `json:"batchsize"`
	BatchFrequency       int    `json:"batchfrequency"`
	WriteTimeout         int    `json:"write-timeout"`
	FlushBytes           int    `json:"flush-bytes"`
	FlushInterval        int    `json:"flush-interval"`
	AccumulatorFrequency int  `json:"accumulator-frequency"`
	WritePerMeasurement  bool `json:"write-per-measurement"`
}

type Article struct {
	Measurement string                 `json:"measurement"`
	Fields      map[string]interface{} `json:"fields"`
	Tags        map[string]string      `json:"tags"`
}

func articleAcculumator(jctx *JCtx) {
	deviceTs := gDeviceTs
	freq := jctx.config.Es.AccumulatorFrequency
	accumulatorCh := make(chan *metricIDB, 1024*10)
	jctx.esCtx.accumulatorCh = accumulatorCh
	jLog(jctx, fmt.Sprintln("Accumulator frequency:", freq))

	ticker := time.NewTicker(time.Duration(freq) * time.Millisecond)

	go func() {
		for range ticker.C {
			n := len(accumulatorCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintf("Accumulated articles : %d\n", n))
				var lastArticle Article
				var articles []Article
				for i := 0; i < n; i++ {
					m := <-accumulatorCh

					if lastArticle.Measurement == "" {
						mName := ""
						if jctx.config.Es.Measurement != "" {
							mName = jctx.config.Es.Measurement
						} else {
							mName = m.tags["sensor"]
						}

						m.fields[deviceTs] = int64(m.ts)
						lastArticle = Article{Measurement: mName, Fields: m.fields, Tags: m.tags}
					} else {
						// let's see if we can merge
						var fieldFound = false
						eq := reflect.DeepEqual(m.tags, lastArticle.Tags)
						if eq {
							// tags are equal so most likely we will be able to merge.
							// we would also need to see if the field is not already part of the article,
							// if it is then we can merge because in 'config false' world of yang, keys
							// are optional inside list so instead of losing the article we'd  not merge.
							for mk := range m.fields {
								lastKV := lastArticle.Fields
								if _, ok := lastKV[mk]; ok {
									fieldFound = true
									break
								}
							}
						}
						if eq && !fieldFound {
							// We can merge
							lastKV := lastArticle.Fields
							name := lastArticle.Measurement
							// get the fields from last article for merging
							for k, v := range lastKV {
								m.fields[k] = v
							}
							lastArticle = Article{Measurement: name, Fields: m.fields, Tags: m.tags}
						} else {
							// lastArticle tags and current article tags differes so we can not merge.
							// toss current article into the slice (articles) and handle current article
							// by creating new Article
							mName := ""
							if jctx.config.Es.Measurement != "" {
								mName = jctx.config.Es.Measurement
							} else {
								mName = m.tags["sensor"]
							}
							m.fields[deviceTs] = int64(m.ts)
							a := Article{Measurement: mName, Fields: m.fields, Tags: m.tags}
							articles = append(articles, lastArticle)
							lastArticle = a
						}
					}

					if len(articles) > 0 {
						// See if we need to add lastArticle we are processing
						if eq := reflect.DeepEqual(articles[len(articles)-1], lastArticle); !eq {
							articles = append(articles, lastArticle)
						}
					}

					if len(articles) > 0 {
						var countSuccessful uint64
						bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
							Client:        jctx.esCtx.esClient,
							Index:         jctx.config.Es.Idxname,
							NumWorkers:    runtime.NumCPU(),
							FlushBytes:    jctx.config.Es.FlushBytes,
							FlushInterval: time.Duration(jctx.config.Es.FlushInterval) * time.Second,
							Timeout:       time.Duration(jctx.config.Es.WriteTimeout) * time.Second,
						})
						if err != nil {
							jLog(jctx, fmt.Sprintf("Error creating the indexer: %s", err))
						}

						for _, article := range articles {
							if jctx.config.Log.Verbose {
								jLog(jctx, fmt.Sprintf("\n\nArticle Name = %s\n", article.Measurement))
								jLog(jctx, fmt.Sprintf("tags are following ...."))
								for k, v := range article.Tags {
									jLog(jctx, fmt.Sprintf("%s = %s", k, v))
								}
								jLog(jctx, fmt.Sprintf("fields are following ...."))
								for k, v := range article.Fields {
									jLog(jctx, fmt.Sprintf("%s = %s", k, v))
								}
							}
							content, err := json.Marshal(article)
							if err != nil {
								jLog(jctx, fmt.Sprintf("articleAcculumator: Could not encode article: %s", err))
							}
							err = bi.Add(
								context.Background(),
								esutil.BulkIndexerItem{
									Action: "index",
									Body:   bytes.NewReader(content),
									OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
										atomic.AddUint64(&countSuccessful, 1)
									},
									OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
										if err != nil {
											jLog(jctx, fmt.Sprintf("articleAcculumator: Batch write failed: %s", err))
										} else {
											jLog(jctx, fmt.Sprintf("articleAcculumator: Batch write failed: %s: %s", res.Error.Type, res.Error.Reason))
										}
									},
								},
							)

							if err != nil {
								jLog(jctx, fmt.Sprintf("articleAcculumator failed to add an item to the indexer: %s", err))
							}
						}
						jLog(jctx, fmt.Sprintf("Write %d articles successfully", uint64(countSuccessful)))
						if bi != nil {
							if err := bi.Close(context.Background()); err != nil {
								log.Printf("esInit failed to close bulk indexer: %v\n", err)
							}
						}
					}
				}
			}
		}
	}()
}

func esdbBatchWriteM(jctx *JCtx) {
	if jctx.esCtx.esClient == nil {
		return
	}

	batchSize := jctx.config.Es.BatchSize
	batchMCh := make(chan []*Article, batchSize/4)
	jctx.esCtx.batchWCh = batchMCh

	// wake up periodically and perform batch write into EsDB
	bFreq := jctx.config.Es.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			m := make(map[string][]*Article)
			n := len(batchMCh)
			if n != 0 {
				jLog(jctx, fmt.Sprintln("#elements in the batchMCh channel : ", n))
				for i := 0; i < n; i++ {
					packet := <-batchMCh
					for j := 0; j < len(packet); j++ {
						v := m[packet[j].Measurement]
						m[packet[j].Measurement] = append(v, packet[j])
					}
				}
				jLog(jctx, fmt.Sprintln("#elements in the measurement map : ", len(m)))
			}

			for measurement, articles := range m {
				jLog(jctx, fmt.Sprintf("measurement: %s, data len: %d", measurement, len(articles)))

				var countSuccessful uint64
				bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
					Client:        jctx.esCtx.esClient,
					Index:         jctx.config.Es.Idxname,
					NumWorkers:    runtime.NumCPU(),
					FlushBytes:    jctx.config.Es.FlushBytes,
					FlushInterval: time.Duration(jctx.config.Es.FlushInterval) * time.Second,
					Timeout:       time.Duration(jctx.config.Es.WriteTimeout) * time.Second,
				})
				if err != nil {
					jLog(jctx, fmt.Sprintf("Error creating the indexer: %s", err))
				}
				for _, article := range articles {
					content, err := json.Marshal(article)
					if err != nil {
						jLog(jctx, fmt.Sprintf("Could not encode article: %s", err))
					}
					err = bi.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							Action: "index",
							Body: bytes.NewReader(content),
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								atomic.AddUint64(&countSuccessful, 1)
							},
							OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
								if err != nil {
									jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s: %v", measurement, err))
								} else {
									jLog(jctx, fmt.Sprintf("Batch DB write failed for measurement %s %s: %s", measurement, res.Error.Type, res.Error.Reason))
								}
							},
						},
					)

					if err != nil {
						jLog(jctx, fmt.Sprintf("Failed to add an item to the indexer for measurement %s: %s", measurement, err))
					}
				}
				jLog(jctx, fmt.Sprintf("Write %d documents successfully for measurement %s, post batch write available documents: %d", uint64(countSuccessful), measurement, len(batchMCh)))
				if bi != nil {
					if err := bi.Close(context.Background()); err != nil {
						log.Printf("esInit failed to close bulk indexer: %v\n", err)
					}
				}
			}
		}
	}()
}

func esdbBatchWrite(jctx *JCtx) {
	if jctx.esCtx.esClient == nil {
		return
	}

	batchSize := jctx.config.Es.BatchSize
	batchCh := make(chan []*Article, batchSize)
	jctx.esCtx.batchWCh = batchCh

	// wake up periodically and perform batch write into EsDB
	bFreq := jctx.config.Es.BatchFrequency
	jLog(jctx, fmt.Sprintln("batch size:", batchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(batchCh)
			if n != 0 {
				articles := make([]*Article, 0)
				var countSuccessful uint64
				bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
					Client:        jctx.esCtx.esClient,
					Index:         jctx.config.Es.Idxname,
					NumWorkers:    runtime.NumCPU(),
					FlushBytes:    jctx.config.Es.FlushBytes,
					FlushInterval: time.Duration(jctx.config.Es.FlushInterval) * time.Second,
					Timeout:       time.Duration(jctx.config.Es.WriteTimeout) * time.Second,
				})
				if err != nil {
					jLog(jctx, fmt.Sprintf("Error creating the indexer: %s", err))
				}

				for i := 0; i < n; i++ {
					packet := <-batchCh
					for j := 0; j < len(packet); j++ {
						articles = append(articles, packet[j])
					}
				}
				jLog(jctx, fmt.Sprintf("Batch processing: #packets:%d #articles:%d\n", n, len(articles)))

				for _, article := range articles {
					content, err := json.Marshal(article)
					if err != nil {
						jLog(jctx, fmt.Sprintf("Could not encode article: %s", err))
					}
					err = bi.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							Action: "index",
							Body:   bytes.NewReader(content),
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								atomic.AddUint64(&countSuccessful, 1)
							},
							OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
								if err != nil {
									jLog(jctx, fmt.Sprintf("Batch DB write failed: %s", err))
								} else {
									jLog(jctx, fmt.Sprintf("Batch DB write failed: %s: %s", res.Error.Type, res.Error.Reason))
								}
							},
						},
					)

					if err != nil {
						jLog(jctx, fmt.Sprintf("Failed to add an item to the indexer: %s", err))
					}
				}
				jLog(jctx, fmt.Sprintf("Write %d documents successfully, post batch write available documents: %d ", uint64(countSuccessful), len(batchCh)))
				if bi != nil {
					if err := bi.Close(context.Background()); err != nil {
						log.Printf("esInit failed to close bulk indexer: %v\n", err)
					}
				}
			}
		}
	}()
}

// A go routine to add one telemetry packet in to ESDB
func addESDB(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	deviceTs := gDeviceTs
	cfg := jctx.config

	prefix := ""
	prefixXmlpath := ""
	var prefixTags map[string]string
	var tags map[string]string
	var xmlpath string
	prefixTags = nil

	articles := make([]*Article, 0)
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

		if jctx.esCtx.esClient == nil {
			continue
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
						jLog(jctx, fmt.Sprintf("addESDB: Could not get NewRow (no merge): %v", err))
						continue
					}
					rows = append(rows, rw)
				}
			} else {
				// First row for this sensor
				kv[deviceTs] = int64(ocData.Timestamp)
				rw, err := newRow(tags, kv)
				if err != nil {
					jLog(jctx, fmt.Sprintf("addESDB: Could not get NewRow (first row): %v", err))
					continue
				}
				rows = append(rows, rw)
			}
		}
	}
	if len(rows) > 0 {
		for _, row := range rows {
		        articles = append(articles, &Article{
                        Measurement: mName(ocData, jctx.config),
                        Fields:      row.fields,
                        Tags:        row.tags,
		        })
		}
	}

	if len(rows) > 0 {
		jctx.esCtx.batchWCh <- articles

		if IsVerboseLogging(jctx) {
			jLog(jctx, fmt.Sprintf("Sending %d articles to batch channel for path: %s\n", len(articles), ocData.Path))
			for i := 0; i < len(articles); i++ {
				jLog(jctx, fmt.Sprintf("Tags: %+v\n", articles[i].Tags))
				jLog(jctx, fmt.Sprintf("KVs : %+v\n", articles[i].Fields))
			}
		}
	}
}

func getEsClient(cfg Config) *elasticsearch.Client {
	if cfg.Es.Server == "" {
		return nil
	}

	// TODO: Vivek Resolve it only once and reuse the endpoint
	resolvedArr, err := net.ResolveTCPAddr("tcp", cfg.Es.Server+":"+strconv.Itoa(cfg.Es.Port))
	if err != nil {
		log.Printf("ResolveTCPAddr failed for %s, err: %v\n", cfg.Es.Server+":"+strconv.Itoa(cfg.Es.Port), err)
		return nil
	}
	addr := fmt.Sprintf("http://%v:%v", resolvedArr.IP, resolvedArr.Port)
	c, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			addr,
		},
		Username: cfg.Es.User,
		Password: cfg.Es.Password,
	})

	if err != nil {
		log.Printf("Failed to get esdb client: %v\n", err)
	}
	return c
}

func esInit(jctx *JCtx) {
	cfg := jctx.config

	jLog(jctx, "invoking getEsClient")
	jctx.esCtx.esClient = getEsClient(cfg)
	jctx.esCtx.reXpath = regexp.MustCompile(MatchExpressionXpath)
	jctx.esCtx.reKey = regexp.MustCompile(MatchExpressionKey)
	if cfg.Es.Server != "" && jctx.esCtx.esClient != nil {
		res, err := jctx.esCtx.esClient.Indices.Exists([]string{cfg.Es.Idxname})
		if err != nil {
			log.Printf("esInit failed to check existing indices: %v\n", err)
		}
		if res.StatusCode == 200 && cfg.Es.Recreate {
			jctx.esCtx.esClient.Indices.Delete(
				[]string{cfg.Es.Idxname},
				jctx.esCtx.esClient.Indices.Delete.WithIgnoreUnavailable(true))
			result, err := jctx.esCtx.esClient.Indices.Create(cfg.Es.Idxname)
			if err != nil {
			        log.Printf("esInit failed to create bulk indexer: %v\n", err)
			}
			result.Body.Close()
			if err != nil {
				log.Printf("esInit failed to bulk indexer creation err: %s", err)
			}
			if result.IsError() {
				log.Printf("esInit failed to bulk indexer creation err: %s", result)
			}
		} else if res.StatusCode == 404 {
			result, err := jctx.esCtx.esClient.Indices.Create(cfg.Es.Idxname)
			if err != nil {
			        log.Printf("esInit failed to create bulk indexer: %v\n", err)
			}
			result.Body.Close()
			if err != nil {
				log.Printf("esInit failed to bulk indexer creation err: %s", err)
			}
			if result.IsError() {
				log.Printf("esInit failed to bulk indexer creation err: %s", result)
			}
		}
	}

	if cfg.Es.Server != "" && jctx.esCtx.esClient != nil {
		if cfg.Es.WritePerMeasurement {
			esdbBatchWriteM(jctx)
		} else {
			esdbBatchWrite(jctx)
		}
		articleAcculumator(jctx)
		jLog(jctx, "Successfully initialized ESDBClient")
	}
}

