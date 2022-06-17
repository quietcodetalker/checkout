package notification

import (
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka"
)

// KafkaClient sends predefined messages to kafka.
type KafkaClient interface {
	SendEmailNotification(notification EmailNotification) error
}

type kafkaClient struct {
	emailNotificationsProducer kafka.Producer
}

// NewKafkaClient creates and instance of kafkaClient.
func NewKafkaClient(
	emailNotificationsProducer kafka.Producer,
) *kafkaClient {
	return &kafkaClient{
		emailNotificationsProducer: emailNotificationsProducer,
	}
}

func (c *kafkaClient) SendEmailNotification(notification EmailNotification) error {
	if err := c.emailNotificationsProducer.SendMessage(fmt.Sprint(notification.OrderID), notification); err != nil {
		return fmt.Errorf("%w: send message: %v", ErrInternal, err)
	}
	return nil
}
