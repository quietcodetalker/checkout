package order

import "time"

type CreateOrderReq struct {
	UserID       uint64    `json:"user_id" validate:"required"`
	Items        []*Item   `json:"items" validate:"required"`
	DeliveryDate time.Time `json:"delivery_date" validate:"required"`
	Email        string    `json:"email" validate:"required"`
	Total        float64   `json:"total" validate:"required"`
}

type Payment struct {
	OrderID uint64 `json:"order_id" validate:"required"`
}
