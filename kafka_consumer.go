package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/nileshsimaria/jtimon/dialout"
)

func createKafkaConsumerGroup(name string, topics []string) error {
	// Get the device config to get to know the paths to be subscribed
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Version = sarama.MaxVersion
	kafkaClient, err := sarama.NewClient(strings.Split(*kafkaBroker, ","), kafkaCfg)
	if err != nil {
		log.Printf(fmt.Sprintf("Not able to connect to Kafka broker at %v: %v", *kafkaBroker, err))
		return err
	}

	kafkaConsumerGroup, err := sarama.NewConsumerGroupFromClient(name, kafkaClient)
	if err != nil {
		log.Printf(fmt.Sprintf("Not able to create consumer: %v", err))
		return err
	}

	ctx := context.Background()
	//defer cancel()

	var gnmiConsumer gnmiConsumerGroupHandler
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			log.Printf("Inside loop")
			if err := kafkaConsumerGroup.Consume(ctx, topics, &gnmiConsumer); err != nil {
				log.Printf(fmt.Sprintf("Error from consumer: %v", err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Printf("Err: %v", ctx.Err())
				return
			}
		}
	}()

	return err
}

type gnmiConsumerGroupHandler struct {
	m         sync.RWMutex
	dbHandles map[string]*JCtx
}

func (handler *gnmiConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	var err error
	log.Printf("Setup..")
	handler.dbHandles = map[string]*JCtx{}
	return err
}

func (*gnmiConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	var err error
	log.Printf("Cleanup..")
	return err
}

func (handler *gnmiConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var err error

	// Process the response

	// TODO: Vivek Use metadata to pass the complete context? what kafka version it is in jcloud?
	// TODO: Vivek change cn
	cn := *myListeningIP

	log.Printf("Consuming..")
	for msg := range claim.Messages() {
		var dialOutRsp dialout.DialOutResponse
		proto.Unmarshal(msg.Value, &dialOutRsp)

		var cfg Config
		json.Unmarshal(dialOutRsp.DialOutContext, &cfg)

		var jctx *JCtx
		handler.m.Lock()
		jctx, ok := handler.dbHandles[cfg.Influx.Dbname]
		if !ok {
			jctx = &JCtx{config: cfg}
			if cfg.Influx.Server != "" {
				influxInit(jctx)
				handler.dbHandles[cfg.Influx.Dbname] = jctx
				log.Printf("Influx Host: %v, dbHandles: %v", cn, cfg.Influx.Dbname)
			}
			if cfg.Es.Server != "" {
				esInit(jctx)
				handler.dbHandles[cfg.Es.Idxname] = jctx
				log.Printf("Es Host: %v, dbHandles: %v", cn, cfg.Es.Idxname)
			}
			handler.m.Unlock()
		} else {
			handler.m.Unlock()
			jctx.config = cfg
		}

		for _, rsp := range dialOutRsp.Response {
			err = gnmiHandleResponse(jctx, rsp)
			if err != nil && strings.Contains(err.Error(), gGnmiJtimonIgnoreErrorSubstr) {
				log.Printf("Host: %v, gnmiHandleResponse failed: %v", cn, err)
				continue
			}
		}
	}

	return err
}
