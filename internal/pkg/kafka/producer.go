package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
)

// Producer sends messages to kafka.
type Producer interface {
	SendMessage(key string, msg interface{}) error
	Close() error
}

type saramaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewSaramaProducer creates an instance of saramaProducer.
func NewSaramaProducer(brokers []string, topic string) (*saramaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: init sync producer err: %v", ErrInternal, err)
	}

	p := &saramaProducer{
		producer: producer,
		topic:    topic,
	}

	return p, nil
}

// SendMessage sends a message to kafka.
func (p *saramaProducer) SendMessage(key string, msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%w: marshal err: %v", ErrInvalidArgument, err)
	}

	_, _, err = p.producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: p.topic,
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(b),
		})

	return err
}

// Close closes a connection to kafka.
func (p *saramaProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("%w: close producer err: %v", ErrInternal, err)
	}

	return nil
}
