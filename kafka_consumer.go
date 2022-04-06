package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/nileshsimaria/jtimon/gnmi/gnmi"
)

func createKafkaConsumerGroup(name string, topics []string) error {
	// Process the response
	jctx := JCtx{
		config: Config{
			Host: os.Getenv("MY_NAME"),
			EOS:  true,
			Log: LogConfig{
				Verbose: true,
			},
		},
	}
	if jctx.config.Host == "" {
		jctx.config.Host = "localhost"
	}
	logInit(&jctx)

	// Get the device config to get to know the paths to be subscribed
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Version = sarama.MaxVersion
	kafkaClient, err := sarama.NewClient([]string{*kafkaBroker}, kafkaCfg)
	if err != nil {
		jLog(&jctx, fmt.Sprintf("Not able to connect to Kafka broker at %v: %v", *kafkaBroker, err))
		return err
	}

	kafkaConsumerGroup, err := sarama.NewConsumerGroupFromClient(name, kafkaClient)
	if err != nil {
		jLog(&jctx, fmt.Sprintf("Not able to create consumer: %v", err))
		return err
	}

	ctx := context.Background()
	//defer cancel()

	var gnmiConsumer = gnmiConsumerGroupHandler{jctx: &jctx}
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			log.Printf("Inside loop")
			if err := kafkaConsumerGroup.Consume(ctx, topics, &gnmiConsumer); err != nil {
				jLog(&jctx, fmt.Sprintf("Error from consumer: %v", err))
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
	jctx *JCtx
}

func (*gnmiConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	var err error
	log.Printf("Setup..")
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
		var rspFromDevice gnmi.SubscribeResponse
		proto.Unmarshal(msg.Value, &rspFromDevice)
		err = gnmiHandleResponse(handler.jctx, &rspFromDevice)
		if err != nil && strings.Contains(err.Error(), gGnmiJtimonIgnoreErrorSubstr) {
			log.Printf("Host: %v, gnmiHandleResponse failed: %v", cn, err)
			continue
		}
	}

	return err
}
