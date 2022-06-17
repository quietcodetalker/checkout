package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
)

// Consumer consumes messages from kafka and sends them to consumer group handler provided via the constructor.
type Consumer interface {
	Close() error
}

type saramaConsumer struct {
	router        sarama.ConsumerGroupHandler
	consumerGroup sarama.ConsumerGroup

	ctx       context.Context
	cancelCtx context.CancelFunc
}

// NewSaramaConsumer creates an instance of sarama consumer.
func NewSaramaConsumer(ctx context.Context, brokers []string, topics []string, groupID string, h sarama.ConsumerGroupHandler) (*saramaConsumer, error) {
	var err error

	internalCtx, cancel := context.WithCancel(ctx)

	c := &saramaConsumer{
		ctx:       internalCtx,
		cancelCtx: cancel,
		router:    h,
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	c.consumerGroup, err = sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: init consumer group err: %v", ErrInternal, err)
	}

	consumeMsg := func() <-chan error {
		errCh := make(chan error, 1)
		go func() {
			errCh <- c.consumerGroup.Consume(c.ctx, topics, c.router)
		}()
		return errCh
	}

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case err := <-consumeMsg():
				if err != nil {
					log.Printf("consume msg err: %v", err)
				}
			}
		}
	}()

	return c, nil
}

// Close closes a connection to kafka.
func (c *saramaConsumer) Close() error {
	c.cancelCtx()

	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("%w: close consumer group err: %v", ErrInternal, err)
	}

	return nil
}
