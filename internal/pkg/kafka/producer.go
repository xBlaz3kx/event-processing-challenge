package kafka

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Configuration struct {
	Brokers []string `yaml:"brokers"`
	GroupId string   `yaml:"groupId"`
}

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewProducer creates a new Kafka producer for the given topic.
func NewProducer(logger *zap.Logger, configuration Configuration, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(configuration.Brokers...),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireAll,
		Async:                  true, // make the writer asynchronous
		WriteTimeout:           time.Second * 2,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				logger.Error("failed to write messages", zap.Error(err))
			}
		},
	}
	return &Producer{
		writer: w,
		logger: logger,
	}
}

// Publish publishes a message to the Kafka topic. Message must be json serializable.
func (p *Producer) Publish(ctx context.Context, message interface{}) error {
	// Marshal the message, if possible
	marshal, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}

	p.logger.Info("Publishing message", zap.Any("message", message))
	// Write the message to the Kafka topic
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: marshal,
	})
	if err != nil {
		return errors.Wrap(err, "failed to write message")
	}

	return nil
}
