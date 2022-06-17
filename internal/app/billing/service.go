package billing

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/cache"
	"log"
	"time"
)

type Service interface {
	AddPayment(ctx context.Context, orderID uint64, userID uint64, total float64) error
	GetPayment(ctx context.Context, orderID uint64) (*Payment, error)
	ApprovePayment(ctx context.Context, orderID uint64) error
	CancelPayment(ctx context.Context, orderID uint64) error
}

type service struct {
	repo        Repository
	kafkaClient KafkaClient
	cache       cache.Cache
}

func NewService(repo Repository, kafkaClient KafkaClient, cache cache.Cache) *service {
	return &service{
		repo:        repo,
		kafkaClient: kafkaClient,
		cache:       cache,
	}
}

func (s *service) AddPayment(ctx context.Context, orderID uint64, userID uint64, total float64) error {
	if err := s.repo.AddPayment(ctx, orderID, userID, total); err != nil {
		err = fmt.Errorf("add payment: %w", err)

		go s.kafkaClient.SendReset(ResetMsg{
			OrderID: orderID,
			ErrMsg:  err.Error(),
		})

		return err
	}

	if err := s.cache.Set(ctx, fmt.Sprint(orderID), Payment{
		OrderID: orderID,
		UserID:  userID,
		Total:   total,
	}, time.Hour*24); err != nil {
		log.Printf("[ERROR] set cache value: %v", err)
	}

	if err := s.kafkaClient.SendPendingPayment(Payment{
		OrderID: orderID,
		Total:   total,
	}); err != nil {
		return fmt.Errorf("send pending payment: %w", err)
	}

	return nil
}

func (s *service) GetPayment(ctx context.Context, orderID uint64) (*Payment, error) {
	v, err := s.cache.Get(ctx, fmt.Sprint(orderID))
	if err == nil {
		if payment, ok := v.(Payment); ok {
			return &payment, nil
		}
		return nil, fmt.Errorf("%w: payment value of invalid type", ErrInternal)
	}
	log.Printf("[ERROR] get cache value: %v", err)

	payment, err := s.repo.GetPayment(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}

	if err := s.cache.Set(ctx, fmt.Sprint(orderID), *payment, time.Hour*24); err != nil {
		log.Printf("[ERROR] set cache value: %v", err)
	}

	return payment, nil
}

func (s *service) ApprovePayment(ctx context.Context, orderID uint64) error {
	p, err := s.repo.ApprovePayment(ctx, orderID)
	if err != nil {
		err = fmt.Errorf("approve payment: %w", err)

		go s.kafkaClient.SendReset(ResetMsg{
			OrderID: orderID,
			ErrMsg:  err.Error(),
		})

		return err
	}

	if err := s.kafkaClient.SendPaidPayment(Payment{
		OrderID: orderID,
		Total:   p.Total,
	}); err != nil {
		return fmt.Errorf("send pending payment: %w", err)
	}

	return nil
}

func (s *service) CancelPayment(ctx context.Context, orderID uint64) error {
	if err := s.repo.CancelPayment(ctx, orderID); err != nil {
		return fmt.Errorf("cancel payment: %w", err)
	}

	return nil
}
