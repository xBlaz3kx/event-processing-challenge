package kafka

import (
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/net/context"
)

type Configuration struct {
	Connection string `yaml:"connection"`
}

type Producer struct {
	writer *kafka.Writer
	dialer *kafka.Dialer
}

func NewProducer() {

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "topic", 1)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("one!")},
		kafka.Message{Value: []byte("two!")},
		kafka.Message{Value: []byte("three!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
}
