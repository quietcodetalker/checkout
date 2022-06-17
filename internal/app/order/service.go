package order

import (
	"context"
	"fmt"
)

type Service interface {
	Create(ctx context.Context, req CreateOrderReq) (uint64, error)
	Delete(ctx context.Context, orderID uint64) error
	SendPaidOrder(ctx context.Context, orderID uint64) error
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

func (s *service) Create(ctx context.Context, req CreateOrderReq) (uint64, error) {
	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("create: %w", err)
	}

	if err := s.kafkaClient.SendSavedOrder(Order{
		OrderID:      id,
		UserID:       req.UserID,
		DeliveryDate: req.DeliveryDate,
		Email:        req.Email,
		Total:        req.Total,
		Items:        req.Items,
	}); err != nil {
		return 0, fmt.Errorf("send saved order: %w", err)
	}

	return id, nil
}

func (s *service) SendPaidOrder(ctx context.Context, orderID uint64) error {
	order, err := s.repo.Get(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	if err := s.kafkaClient.SendPaidOrder(*order); err != nil {
		return fmt.Errorf("send paid order: %w", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, orderID uint64) error {
	if err := s.repo.Delete(ctx, orderID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
