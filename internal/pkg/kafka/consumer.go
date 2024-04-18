package kafka

import (
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Consumer[T comparable] struct {
	logger *zap.Logger
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer for the given topic.
func NewConsumer[T comparable](logger *zap.Logger, cfg Configuration, topic string) *Consumer[T] {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	return &Consumer[T]{
		logger: logger,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   cfg.Brokers,
			Topic:     topic,
			GroupID:   cfg.GroupId,
			Partition: 0,
			MinBytes:  1e3, // 1KB
			MaxBytes:  1e4, // 1MB
			MaxWait:   time.Millisecond * 40,
			Dialer:    dialer,
		}),
	}
}

// Read reads messages from the Kafka topic and calls the callback function with the model and error.
func (c *Consumer[T]) Read(ctx context.Context, model T, callback func(T, error)) {
	// Wrap in a goroutine to read messages in the background
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ctx, _ := context.WithTimeout(ctx, time.Millisecond*100)
				message, err := c.reader.ReadMessage(ctx)

				switch {
				case err == nil:
					// Unmarshal the message value into the model
					err = json.Unmarshal(message.Value, &model)
					if err != nil {
						callback(model, err)
						continue
					}

					callback(model, nil)
				case errors.Is(err, context.DeadlineExceeded):
					continue
				case errors.Is(err, io.EOF):
					// Reader is closed, exit the goroutine
					return
				default:
					callback(model, err)
					return
				}
			}
		}
	}()
}

func (c *Consumer[T]) Close() error {
	c.logger.Debug("Closing Kafka consumer")
	return c.reader.Close()
}
