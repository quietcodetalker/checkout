package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

const (
	ordersTable      = "orders"
	ordersItemsTable = "orders_items"
)

type Repository interface {
	Create(ctx context.Context, req CreateOrderReq) (uint64, error)
	Delete(ctx context.Context, orderID uint64) error
	Get(ctx context.Context, orderID uint64) (*Order, error)
}

type pgRepo struct {
	db *pgxpool.Pool
}

func NewPgRepo(db *pgxpool.Pool) *pgRepo {
	return &pgRepo{
		db: db,
	}
}

func (r *pgRepo) Create(ctx context.Context, req CreateOrderReq) (uint64, error) {
	var id uint64

	if err := r.execTx(ctx, func(q *pgQueries) error {
		var err error

		id, err = q.createOrder(ctx, req.UserID, req.DeliveryDate, req.Email, req.Total)
		if err != nil {
			return fmt.Errorf("create req: %w", err)
		}

		if err := q.createOrderItems(ctx, id, req.Items); err != nil {
			return fmt.Errorf("create req items: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("exec tx: %w", err)
	}

	return id, nil
}

func (r *pgRepo) Delete(ctx context.Context, orderID uint64) error {
	if err := r.execTx(ctx, func(q *pgQueries) error {
		if err := q.deleteOrderItems(ctx, orderID); err != nil {
			return fmt.Errorf("delete order items: %w", err)
		}

		if err := q.deleteOrder(ctx, orderID); err != nil {
			return fmt.Errorf("delete order: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}

	return nil
}

func (r *pgRepo) Get(ctx context.Context, orderID uint64) (*Order, error) {
	var order *Order
	if err := r.execTx(ctx, func(q *pgQueries) error {
		var err error

		order, err = q.getOrder(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order: %w", err)
		}

		order.Items, err = q.getOrderItems(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order items: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("exec tx: %w", err)
	}

	return order, nil
}

// execTx creates a database transaction with ReadCommitted isolation level and
// execute provided function in the scope of the transaction.
func (r *pgRepo) execTx(ctx context.Context, fn func(queries *pgQueries) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("%w: begin transaction: %v", ErrInternal, err)
	}

	q := &pgQueries{db: tx}
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx: %w, rb: %v", err, rbErr)
		}
		return fmt.Errorf("transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%w: commit transaction: %v", ErrInternal, err)
	}

	return nil
}

// DBTX is an interface that both *pgxpool.Pool and pgx.Tx implements.
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type pgQueries struct {
	db DBTX
}

var createOrderQuery = fmt.Sprintf(`
INSERT INTO %s
(user_id, delivery_date, email, total)
VALUES ($1, $2, $3, $4)
RETURNING order_id
`, ordersTable)

func (q *pgQueries) createOrder(
	ctx context.Context,
	userID uint64,
	deliveryDate time.Time,
	email string,
	total float64,
) (uint64, error) {
	var id uint64
	if err := q.db.QueryRow(ctx, createOrderQuery, userID, deliveryDate, email, total).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "orders_pkey":
				return 0, fmt.Errorf("%w: db exec: %v", ErrFailedPrecondition, err)
			}
		}
		return 0, fmt.Errorf("%w: db exec: %v", ErrInternal, err)
	}

	return id, nil
}

var createOrderItemQuery = fmt.Sprintf("INSERT INTO %s (order_id, product_id, quantity) VALUES ($1, $2, $3)", ordersItemsTable)

func (q *pgQueries) createOrderItems(ctx context.Context, orderID uint64, items []*Item) error {
	for _, item := range items {
		if _, err := q.db.Exec(ctx, createOrderItemQuery, orderID, item.ProductID, item.Quantity); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.ConstraintName {
				case "orders_items_pkey":
					fallthrough
				case "orders_items_quantity_check":
					fallthrough
				case "orders_items_order_id_fkey":
					return fmt.Errorf("%w: db exec: %v", ErrFailedPrecondition, err)
				}
			}
			return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
		}
	}

	return nil
}

var deleteOrderQuery = fmt.Sprintf("DELETE FROM %s WHERE order_id = $1", ordersTable)

func (q *pgQueries) deleteOrder(ctx context.Context, orderID uint64) error {
	if _, err := q.db.Exec(ctx, deleteOrderQuery, orderID); err != nil {
		return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
	}

	return nil
}

var deleteOrderItemsQuery = fmt.Sprintf("DELETE FROM %s WHERE order_id = $1", ordersItemsTable)

func (q *pgQueries) deleteOrderItems(ctx context.Context, orderID uint64) error {
	if _, err := q.db.Exec(ctx, deleteOrderItemsQuery, orderID); err != nil {
		return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
	}

	return nil
}

var getOrderQuery = fmt.Sprintf(`
SELECT user_id, delivery_date, email, total 
FROM %s WHERE order_id = $1
`, ordersTable)

func (q *pgQueries) getOrder(ctx context.Context, orderID uint64) (*Order, error) {
	o := Order{
		OrderID: orderID,
	}

	if err := q.db.QueryRow(ctx, getOrderQuery, orderID).Scan(
		&o.UserID,
		&o.DeliveryDate,
		&o.Email,
		&o.Total,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: db query row: %v", ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: db query row: %v", ErrInternal, err)
	}

	return &o, nil
}

var getOrderItemsQuery = fmt.Sprintf("SELECT product_id, quantity FROM %s WHERE order_id = $1", ordersItemsTable)

func (q *pgQueries) getOrderItems(ctx context.Context, orderID uint64) ([]*Item, error) {
	var items []*Item
	rows, err := q.db.Query(ctx, getOrderItemsQuery, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: db query row: %v", ErrInternal, err)
	}
	defer rows.Close()

	var productID uint64
	var quantity uint64
	for rows.Next() {
		if err := rows.Scan(&productID, &quantity); err != nil {
			return nil, fmt.Errorf("%w: rows scan: %v", ErrInternal, err)
		}

		items = append(items, &Item{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: rows err: %v", ErrInternal, err)
	}

	return items, nil
}
