package billing

import (
	"database/sql/driver"
	"fmt"
)

type Payment struct {
	OrderID uint64  `json:"order_id" validate:"required"`
	UserID  uint64  `json:"user_id"`
	Total   float64 `json:"total" validate:"required"`
}

type PaymentStatus int

func (t *PaymentStatus) Scan(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w: value of unexpected type <%#v>", ErrInternal, v)
	}

	// 'pending', 'paid', 'cancelled', 'failed'
	switch str {
	case "pending":
		*t = Pending
	case "paid":
		*t = Paid
	case "cancelled":
		*t = Cancelled
	case "failed":
		*t = Failed
	default:
		return fmt.Errorf("%w: unexpected value <%#v>", ErrInternal, str)
	}

	return nil
}

func (t PaymentStatus) Value() (driver.Value, error) {
	switch t {
	case Pending:
		return "pending", nil
	case Paid:
		return "paid", nil
	case Cancelled:
		return "cancelled", nil
	case Failed:
		return "failed", nil
	default:
		return nil, fmt.Errorf("%w: unexpected value <%#v>", ErrInternal, t)
	}
}

const (
	Pending PaymentStatus = iota
	Paid
	Cancelled
	Failed
)
