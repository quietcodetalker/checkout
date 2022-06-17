package billing

import (
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka"
)

// KafkaClient sends predefined messages to kafka.
type KafkaClient interface {
	SendPendingPayment(payment Payment) error
	SendPaidPayment(payment Payment) error
	SendReset(msg ResetMsg) error
}

type kafkaClient struct {
	pendingPaymentsProducer kafka.Producer
	paidPaymentsProducer    kafka.Producer
	resetProducer           kafka.Producer
}

// NewKafkaClient creates and instance of kafkaClient.
func NewKafkaClient(
	pendingPaymentsProducer kafka.Producer,
	paidPaymentsProducer kafka.Producer,
	resetProducer kafka.Producer,
) *kafkaClient {
	return &kafkaClient{
		pendingPaymentsProducer: pendingPaymentsProducer,
		paidPaymentsProducer:    paidPaymentsProducer,
		resetProducer:           resetProducer,
	}
}

func (c *kafkaClient) SendPendingPayment(payment Payment) error {
	if err := c.pendingPaymentsProducer.SendMessage(fmt.Sprint(payment.OrderID), payment); err != nil {
		return fmt.Errorf("%w: send message: %v", ErrInternal, err)
	}
	return nil
}

func (c *kafkaClient) SendPaidPayment(payment Payment) error {
	if err := c.paidPaymentsProducer.SendMessage(fmt.Sprint(payment.OrderID), payment); err != nil {
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
