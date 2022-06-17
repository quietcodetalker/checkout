package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/go-playground/validator/v10"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka/router"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka/router/middleware"
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

	h.router.Handle("new_orders", h.create)
	h.router.Handle("paid_payments", h.handlePaidPayment)
	h.router.Handle("reset", h.reset)
	h.router.Handle("cancel", h.cancel)
}

func (h *KafkaHandler) create(ctx context.Context, _ string, raw []byte) error {
	var req CreateOrderReq
	if err := json.Unmarshal(raw, &req); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if _, err := h.svc.Create(ctx, req); err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (h *KafkaHandler) handlePaidPayment(ctx context.Context, _ string, raw []byte) error {
	var req Payment
	if err := json.Unmarshal(raw, &req); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.SendPaidOrder(ctx, req.OrderID); err != nil {
		return fmt.Errorf("create: %w", err)
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

	if err := h.svc.Delete(ctx, msg.OrderID); err != nil {
		return fmt.Errorf("delete: %w", err)
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

	if err := h.svc.Delete(ctx, msg.OrderID); err != nil {
		return fmt.Errorf("cancel: %w", err)
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
