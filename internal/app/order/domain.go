package order

import "time"

// Item represents product with its quantity.
type Item struct {
	ProductID uint64 `json:"product_id"`
	Quantity  uint64 `json:"quantity"`
}

// Order is a order request message.
type Order struct {
	OrderID      uint64    `json:"order_id"`
	UserID       uint64    `json:"user_id"`
	Total        float64   `json:"total"`
	DeliveryDate time.Time `json:"delivery_date"`
	Email        string    `json:"email"`
	Items        []*Item   `json:"items"`
}

// ResetMsg represents reset message.
type ResetMsg struct {
	OrderID uint64 `json:"order_id" validate:"required"`
	ErrMsg  string `json:"err_msg" validate:"required"`
}

// CancelMsg represents cancellation message.
type CancelMsg struct {
	OrderID uint64 `json:"order_id" validate:"required"`
	Reason  string `json:"reason" validate:"required"`
}
