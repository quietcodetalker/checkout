package order

import (
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka"
)

// KafkaClient sends predefined messages to kafka.
type KafkaClient interface {
	SendSavedOrder(order Order) error
	SendReset(msg ResetMsg) error
	SendPaidOrder(order Order) error
}

type kafkaClient struct {
	savedOrdersProducer kafka.Producer
	paidOrdersProducer  kafka.Producer
	resetProducer       kafka.Producer
}

// NewKafkaClient creates and instance of kafkaClient.
func NewKafkaClient(
	savedOrdersProducer kafka.Producer,
	paidOrdersProducer kafka.Producer,
	resetProducer kafka.Producer,
) *kafkaClient {
	return &kafkaClient{
		savedOrdersProducer: savedOrdersProducer,
		paidOrdersProducer:  paidOrdersProducer,
		resetProducer:       resetProducer,
	}
}

func (c *kafkaClient) SendSavedOrder(order Order) error {
	if err := c.savedOrdersProducer.SendMessage(fmt.Sprint(order.OrderID), order); err != nil {
		return fmt.Errorf("%w: send message: %v", ErrInternal, err)
	}
	return nil
}

func (c *kafkaClient) SendPaidOrder(order Order) error {
	if err := c.paidOrdersProducer.SendMessage(fmt.Sprint(order.OrderID), order); err != nil {
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
