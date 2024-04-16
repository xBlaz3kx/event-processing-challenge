package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xBlaz3kx/event-processing-challenge/internal/generator"
	"github.com/xBlaz3kx/event-processing-challenge/internal/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func main() {
	logger := observability.NewLogger("debug")
	logger.Info("Started the generator")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// todo fetch from cfg file/env
	kafkaCfg := "kafka:9092"

	// todo revisit and refactor the Kafka package to make it more testable
	conn, err := kafka.DialLeader(ctx, "tcp", kafkaCfg, "casino-event", 0)
	if err != nil {
		logger.Fatal("failed to dial leader", zap.Error(err))
	}

	gen := generator.New()
	defer gen.Cleanup()

	for event := range gen.Generate(ctx) {

		marshal, err := json.Marshal(event)
		if err != nil {
			logger.Error("failed to marshal event:", zap.Error(err))
			continue
		}

		err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			logger.Error("failed to write messages", zap.Error(err))
			continue
		}

		_, err = conn.WriteMessages(kafka.Message{Value: marshal})
		if err != nil {
			logger.Error("failed to write messages", zap.Error(err))
			continue
		}

		logger.Info("Published event", zap.Any("event", event))
	}

	logger.Info("Finished generating events")
}
