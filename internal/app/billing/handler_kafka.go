package billing

import (
	"github.com/go-playground/validator/v10"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka/router"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka/router/middleware"
)

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
)

type KafkaHandler struct {
	svc      Service
	router   *router.SaramaRouter
	validate *validator.Validate
}

func NewKafkaHandler(
	svc Service,
) *KafkaHandler {
	h := &KafkaHandler{
		svc:      svc,
		router:   router.NewSaramaRouter(),
		validate: validator.New(),
	}

	h.setupRoutes()

	return h
}

func (h *KafkaHandler) setupRoutes() {
	h.router.Use(middleware.Logger)

	h.router.Handle("reserved_orders", h.createPayment)
	h.router.Handle("receipts", h.approvePayment)
	h.router.Handle("cancel", h.cancel)
	h.router.Handle("reset", h.reset)
}

func (h *KafkaHandler) createPayment(ctx context.Context, _ string, raw []byte) error {
	var msg Order
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.AddPayment(ctx, msg.OrderID, msg.UserID, msg.Total); err != nil {
		return fmt.Errorf("createPayment: %w", err)
	}

	return nil
}

func (h *KafkaHandler) approvePayment(ctx context.Context, _ string, raw []byte) error {
	var msg Receipt
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.ApprovePayment(ctx, msg.OrderID); err != nil {
		return fmt.Errorf("approve payment: %w", err)
	}

	return nil
}

func (h *KafkaHandler) reset(ctx context.Context, _ string, raw []byte) error {
	var msg ResetMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.CancelPayment(ctx, msg.OrderID); err != nil {
		return fmt.Errorf("cancel reservations: %w", err)
	}

	return nil
}

func (h *KafkaHandler) cancel(ctx context.Context, _ string, raw []byte) error {
	var msg CancelMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.CancelPayment(ctx, msg.OrderID); err != nil {
		return fmt.Errorf("cancel reservations: %w", err)
	}

	return nil
}

func (h *KafkaHandler) Setup(session sarama.ConsumerGroupSession) error {
	return h.router.Setup(session)
}

func (h *KafkaHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return h.router.Cleanup(session)
}

func (h *KafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return h.router.ConsumeClaim(session, claim)
}
