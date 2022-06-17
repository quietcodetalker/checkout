package notification

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

var (
	notificationsTable = "notifications"
)

type Repository interface {
	CreateNotification(ctx context.Context, orderID uint64, userID uint64, ts time.Time) (uint64, error)
	GetTodayNotifications(ctx context.Context) ([]*Notification, error)
}

type pgRepo struct {
	db *pgxpool.Pool
}

func NewPgRepo(db *pgxpool.Pool) *pgRepo {
	return &pgRepo{
		db: db,
	}
}

var createNotification = fmt.Sprintf(`
INSERT INTO %s
(order_id, user_id, ts)
VALUES ($1, $2, $3)
RETURNING id
`, notificationsTable)

func (r *pgRepo) CreateNotification(ctx context.Context, orderID uint64, userID uint64, ts time.Time) (uint64, error) {
	var id uint64
	if err := r.db.QueryRow(ctx, createNotification, orderID, userID, ts).Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: db query row: %v", ErrInternal, err)
	}

	return id, nil
}

var getTodayNotifications = fmt.Sprintf(`
SELECT id, order_id, user_id, ts
FROM %s
WHERE DATE(ts) = current_date
`, notificationsTable)

func (r *pgRepo) GetTodayNotifications(ctx context.Context) ([]*Notification, error) {
	var notifications []*Notification
	rows, err := r.db.Query(ctx, getTodayNotifications)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: db query: %v", ErrInternal, err)
	}
	defer rows.Close()

	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.OrderID, &n.UserID, &n.Timestamp); err != nil {
			return nil, fmt.Errorf("%w: rows scan: %v", ErrInternal, err)
		}

		notifications = append(notifications, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: rows err: %v", ErrInternal, err)
	}

	return notifications, nil
}
