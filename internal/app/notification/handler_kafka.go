package notification

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

	h.router.Handle("paid_orders", h.createDelayedNotification)
	h.router.Handle("check", h.check)
}

func (h *KafkaHandler) createDelayedNotification(ctx context.Context, _ string, raw []byte) error {
	var msg Order
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	_, err := h.svc.CreateNotification(ctx, msg.OrderID, msg.UserID, msg.DeliveryDate)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (h *KafkaHandler) check(ctx context.Context, _ string, raw []byte) error {
	var msg CheckReq
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: unmarshal: %v", ErrInvalidMsg, err)
	}

	if err := h.validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: validate: %v", ErrInvalidMsg, err)
	}

	if err := h.svc.Check(ctx); err != nil {
		return fmt.Errorf("create: %w", err)
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
