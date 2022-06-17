package stock

import (
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka"
)

// KafkaClient sends predefined messages to kafka.
type KafkaClient interface {
	SendReservedOrder(order Order) error
	SendReset(msg ResetMsg) error
}

type kafkaClient struct {
	reservedOrdersProducer kafka.Producer
	resetProducer          kafka.Producer
}

// NewKafkaClient creates and instance of kafkaClient.
func NewKafkaClient(
	reservedOrderProducer kafka.Producer,
	resetProducer kafka.Producer,
) *kafkaClient {
	return &kafkaClient{
		reservedOrdersProducer: reservedOrderProducer,
		resetProducer:          resetProducer,
	}
}

func (c *kafkaClient) SendReservedOrder(order Order) error {
	if err := c.reservedOrdersProducer.SendMessage(fmt.Sprint(order.OrderID), order); err != nil {
		return fmt.Errorf("%w: send message: %v", ErrInternal, err)
	}
	return nil
}

func (c *kafkaClient) SendReset(msg ResetMsg) error {
	if err := c.resetProducer.SendMessage(fmt.Sprint(msg.OrderID), msg); err != nil {
		return fmt.Errorf("%w: send message: %v", ErrInternal, err)
	}
	return nil
}
