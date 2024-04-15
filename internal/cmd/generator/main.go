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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	gen := generator.New()
	defer gen.Cleanup()

	for event := range gen.Generate(ctx) {

		// todo revisit and refactor the Kafka package to make it more testable
		conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "topic", 1)
		if err != nil {
			logger.Fatal("failed to dial leader", zap.Error(err))
		}

		marshal, err := json.Marshal(event)
		if err != nil {
			logger.Fatal("failed to marshal event:", zap.Error(err))
			continue
		}

		err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.WriteMessages(kafka.Message{Value: marshal})
		if err != nil {
			logger.Fatal("failed to write messages", zap.Error(err))
			continue
		}

		logger.Info("Published event", zap.Any("event", event))
	}

	logger.Info("Finished generating events")
}
