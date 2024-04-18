package main

import (
	"os"
	"os/signal"
	"strings"

	kafkaApi "github.com/xBlaz3kx/event-processing-challenge/internal/api/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/generator"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/kafka"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func main() {
	logger := observability.NewLogger("debug")
	logger.Info("Started the generator")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Get the Kafka configuration from the environment
	brokers := os.Getenv("KAFKA_BROKERS")
	kafkaCfg := kafka.Configuration{Brokers: strings.Split(brokers, ",")}

	producer := kafka.NewProducer(logger, kafkaCfg, kafkaApi.BaseEventTopic)
	gen := generator.New()
	defer gen.Cleanup()

	for event := range gen.Generate(ctx) {

		err := producer.Publish(ctx, event)
		if err != nil {
			logger.Error("failed to publish messages", zap.Error(err))
			continue
		}

		logger.Info("Published event", zap.Any("event", event))
	}

	logger.Info("Finished generating events")
}
