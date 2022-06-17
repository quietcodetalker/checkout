package notification

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Service interface {
	CreateNotification(ctx context.Context, orderID uint64, userID uint64, ts time.Time) (uint64, error)
	Check(ctx context.Context) error
}

type service struct {
	repo        Repository
	kafkaClient KafkaClient
}

func NewService(repo Repository, kafkaClient KafkaClient) *service {
	return &service{
		repo:        repo,
		kafkaClient: kafkaClient,
	}
}

func (s service) CreateNotification(ctx context.Context, orderID uint64, userID uint64, ts time.Time) (uint64, error) {
	id, err := s.repo.CreateNotification(ctx, orderID, userID, ts)
	if err != nil {
		return 0, fmt.Errorf("create notification: %w", err)
	}

	return id, nil
}

func (s service) Check(ctx context.Context) error {
	notifications, err := s.repo.GetTodayNotifications(ctx)
	if err != nil {
		return fmt.Errorf("get today notifications: %w", err)
	}

	for _, ntf := range notifications {
		ntf := ntf
		go func() {
			if err := s.kafkaClient.SendEmailNotification(EmailNotification{
				OrderID: ntf.OrderID,
				UserID:  ntf.UserID,
			}); err != nil {
				log.Printf("send email notification: %v", err)
			}
		}()
	}

	return nil
}
