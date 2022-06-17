package billing

type Order struct {
	OrderID uint64  `json:"order_id" validate:"required"`
	UserID  uint64  `json:"user_id" validate:"required"`
	Total   float64 `json:"total" validate:"required"`
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

type Receipt struct {
	OrderID uint64 `json:"order_id" validate:"required"`
}
