package stock

import (
	"context"
	"fmt"
)

type Service interface {
	Reserve(ctx context.Context, order Order) error
	CancelReservation(ctx context.Context, orderID uint64) error
	Collect(ctx context.Context, orderID uint64) error
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

func (s *service) Reserve(ctx context.Context, order Order) error {
	if err := s.repo.Reserve(ctx, order.OrderID, order.Items); err != nil {
		err = fmt.Errorf("reserve: %w", err)

		go s.kafkaClient.SendReset(ResetMsg{
			OrderID: order.OrderID,
			ErrMsg:  err.Error(),
		})

		return err
	}

	if err := s.kafkaClient.SendReservedOrder(order); err != nil {
		return fmt.Errorf("send msg to order reservations: %w", err)
	}

	return nil
}

func (s *service) CancelReservation(ctx context.Context, orderID uint64) error {
	if err := s.repo.CancelReservation(ctx, orderID); err != nil {
		return fmt.Errorf("cancel reservations: %w", err)
	}

	return nil
}

func (s *service) Collect(ctx context.Context, orderID uint64) error {
	if err := s.repo.Collect(ctx, orderID); err != nil {
		err = fmt.Errorf("collect: %w", err)

		go s.kafkaClient.SendReset(ResetMsg{
			OrderID: orderID,
			ErrMsg:  err.Error(),
		})

		return err
	}

	return nil
}
