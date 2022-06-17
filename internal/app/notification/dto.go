package notification

import "time"

type CreateNotificationReq struct {
	OrderID   string    `json:"order_id" validate:"required"`
	UserID    uint64    `json:"user_id" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

type CheckReq struct{}

type Payment struct {
	OrderID uint64  `json:"order_id" validate:"required"`
	UserID  uint64  `json:"user_id" validate:"required"`
	Total   float64 `json:"total" validate:"required"`
}

type Item struct {
	ProductID uint64 `json:"product_id"`
	Quantity  uint64 `json:"quantity"`
}

type Order struct {
	OrderID      uint64    `json:"order_id"`
	UserID       uint64    `json:"user_id"`
	Total        float64   `json:"total"`
	DeliveryDate time.Time `json:"delivery_date"`
	Email        string    `json:"email"`
	Items        []*Item   `json:"items"`
}
