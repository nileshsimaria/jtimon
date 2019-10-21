package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
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

	return nil
}

func addKafka(ocData *na_pb.OpenConfigData, jctx *JCtx, rtime time.Time) {
	if jctx.config.Kafka == nil {
		return
	}

	b, err := proto.Marshal(ocData)
	if err != nil {
		jLog(jctx, fmt.Sprintf("Kafka proto marsha error: %v", err))
	}

	m := &sarama.ProducerMessage{
		Topic: "healthbot",
		Value: sarama.ByteEncoder(b),
	}

	if _, _, err := jctx.config.Kafka.producer.SendMessage(m); err != nil {
		jLog(jctx, fmt.Sprintf("Kafka SendMessage failed, error: %v", err))
	}
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
