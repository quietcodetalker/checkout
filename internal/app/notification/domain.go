package notification

import "time"

type Notification struct {
	ID        uint64    `json:"id"`
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

type EmailNotification struct {
	OrderID uint64
	UserID  uint64
}
