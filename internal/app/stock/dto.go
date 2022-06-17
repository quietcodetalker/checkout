package stock

import "time"

// ResetMsg represents reset message.
type ResetMsg struct {
	OrderID uint64 `json:"order_id" validate:"required"`
	ErrMsg  string `json:"err_msg" validate:"required"`
}

// Item represents product with its quantity.
type Item struct {
	ProductID uint64 `json:"product_id"`
	Quantity  uint64 `json:"quantity"`
}

// Order is a order request message.
type Order struct {
	OrderID      uint64    `json:"order_id" validate:"required"`
	UserID       uint64    `json:"user_id" validate:"required"`
	Items        []*Item   `json:"items" validate:"required"`
	DeliveryDate time.Time `json:"delivery_date" validate:"required"`
	Email        string    `json:"email" validate:"required"`
	Total        float64   `json:"total" validate:"required"`
}

// CancelMsg represents cancellation message.
type CancelMsg struct {
	OrderID uint64 `json:"order_id" validate:"required"`
	Reason  string `json:"reason" validate:"required"`
}

type Payment struct {
	OrderID uint64 `json:"order_id" validate:"required"`
}
