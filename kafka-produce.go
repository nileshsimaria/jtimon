package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/influxdata/influxdb/client/v2"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

// KafkaConfig type
type KafkaConfig struct {
	Version            string   `json:"version"`
	Brokers            []string `json:"brokers"`
	ClientID           string   `json:"client-id"`
	Topic              string   `json:"topic"`
	CompressionCodec   int      `json:"compression-codec"`
	RequiredAcks       int      `json:"required-acks"`
	MaxRetry           int      `json:"max-retry"`
	MaxMessageBytes    int      `json:"max-message-bytes"`
	SASLUser           string   `json:"sasl-username"`
	SASLPass           string   `json:"sasl-password"`
	TLSCA              string   `json:"tls-ca"`
	TLSCert            string   `json:"tls-cert"`
	TLSKey             string   `json:"tls-key"`
	InsecureSkipVerify bool     `json:"insecure-skip-verify"`
	producer           sarama.SyncProducer
	kbatchCh           chan []*client.Point
}

// KafkaConnect to connect to kafka bus
func KafkaConnect(k *KafkaConfig) error {
	c := sarama.NewConfig()

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}
		c.Version = version
	}

	if k.ClientID != "" {
		c.ClientID = k.ClientID
	} else {
		c.ClientID = "JTIMON"
	}

	c.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	c.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	c.Producer.Retry.Max = k.MaxRetry
	c.Producer.Return.Successes = true

	if k.MaxMessageBytes > 0 {
		c.Producer.MaxMessageBytes = k.MaxMessageBytes
	}

	if k.TLSCert != "" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: k.InsecureSkipVerify,
			Renegotiation:      tls.RenegotiateNever,
		}

		if k.TLSCA != "" {
			pool, err := getCertPool([]string{k.TLSCA})
			if err != nil {
				return err
			}
			tlsConfig.RootCAs = pool
		}

		if k.TLSCert != "" && k.TLSKey != "" {
			err := loadCert(tlsConfig, k.TLSCert, k.TLSKey)
			if err != nil {
				return err
			}
		}

		if tlsConfig != nil {
			c.Net.TLS.Config = tlsConfig
			c.Net.TLS.Enable = true
		}
	}

	if k.SASLUser != "" && k.SASLPass != "" {
		c.Net.SASL.User = k.SASLUser
		c.Net.SASL.Password = k.SASLPass
		c.Net.SASL.Enable = true
	}

	producer, err := sarama.NewSyncProducer(k.Brokers, c)
	if err != nil {
		return err
	}

	k.producer = producer
	return nil
}

// KafkaInit to initialize Kafka
func KafkaInit(jctx *JCtx) error {
	cfg := jctx.config
	if cfg.Kafka == nil {
		return nil
	}

	if err := KafkaConnect(cfg.Kafka); err != nil {
		return err
	}

	kafkaBatchWrite(jctx)
	return nil
}

func addKafka(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	if jctx.config.Kafka == nil {
		return
	}

	cfg := jctx.config

	prefix := ""
	prefixXmlpath := ""
	var prefixTags map[string]string
	var tags map[string]string
	var xmlpath string
	prefixTags = nil

	points := make([]*client.Point, 0)
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

		tags["device"] = cfg.Host
		tags["sensor"] = ocData.Path

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
					rw, err := newRow(tags, kv)
					if err != nil {
						jLog(jctx, fmt.Sprintf("addIDB: Could not get NewRow (no merge): %v", err))
						continue
					}
					rows = append(rows, rw)
				}
			} else {
				// First row for this sensor
				rw, err := newRow(tags, kv)
				if err != nil {
					jLog(jctx, fmt.Sprintf("addIDB: Could not get NewRow (first row): %v", err))
					continue
				}
				rows = append(rows, rw)
			}
		}
	}
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
		jctx.config.Kafka.kbatchCh <- points

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
}

func kafkaBatchWrite(jctx *JCtx) {
	kbatchSize := 1024
	kbatchCh := make(chan []*client.Point, kbatchSize)
	jctx.config.Kafka.kbatchCh = kbatchCh
	bFreq := 2000

	jLog(jctx, fmt.Sprintln("kafka batch size:", kbatchSize, "batch frequency:", bFreq))

	ticker := time.NewTicker(time.Duration(bFreq) * time.Millisecond)
	go func() {
		for range ticker.C {
			n := len(kbatchCh)
			if n != 0 {
				bp, err := client.NewBatchPoints(client.BatchPointsConfig{
					Database:        jctx.config.Influx.Dbname,
					Precision:       "us",
					RetentionPolicy: jctx.config.Influx.RetentionPolicy,
				})

				if err != nil {
					jLog(jctx, fmt.Sprintf("NewBatchPoints failed, error: %v", err))
					return
				}

				for i := 0; i < n; i++ {
					packet := <-kbatchCh
					for j := 0; j < len(packet); j++ {
						bp.AddPoint(packet[j])
					}
				}

				jLog(jctx, fmt.Sprintf("Kafka batch processing: #packets:%d #points:%d", n, len(bp.Points())))

				var b bytes.Buffer
				for _, p := range bp.Points() {
					if _, err := b.WriteString(p.PrecisionString(bp.Precision())); err != nil {
						jLog(jctx, fmt.Sprintf("Kafka SendMessage prep-1 failed, error: %v", err))
					}

					if err := b.WriteByte('\n'); err != nil {
						jLog(jctx, fmt.Sprintf("Kafka SendMessage prep-2 failed, error: %v", err))
					}
				}

				m := &sarama.ProducerMessage{
					Topic: "test",
					Value: sarama.StringEncoder(b.String()),
				}
				if _, _, err := jctx.config.Kafka.producer.SendMessage(m); err != nil {
					jLog(jctx, fmt.Sprintf("Kafka SendMessage failed, error: %v", err))
				}
			}
		}
	}()
}

func getCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCert(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()
	return nil
}
