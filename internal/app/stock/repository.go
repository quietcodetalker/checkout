package stock

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	quantitiesTable   = "quantities"
	reservationsTable = "reservations"
)

// DBTX is an interface that both *pgxpool.Pool and pgx.Tx implements.
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

// Repository represents stock database repository.
type Repository interface {
	Reserve(ctx context.Context, orderID uint64, items []*Item) error
	CancelReservation(ctx context.Context, orderID uint64) error
	Collect(ctx context.Context, orderID uint64) error
}

type pgRepo struct {
	db      *pgxpool.Pool
	queries *pgQueries
}

// NewPgRepo creates an instance of pgRepo.
func NewPgRepo(db *pgxpool.Pool) *pgRepo {
	return &pgRepo{
		db: db,
		queries: &pgQueries{
			db: db,
		},
	}
}

// Reserve ...
func (r *pgRepo) Reserve(ctx context.Context, orderID uint64, items []*Item) error {
	if err := r.execTx(ctx, func(q *pgQueries) error {
		enough, err := q.isEnough(ctx, items)
		if err != nil {
			return fmt.Errorf("isEnough: %w", err)
		}

		if !enough {
			return ErrNotEnough
		}

		err = q.createReservations(ctx, orderID, items)
		if err != nil {
			return fmt.Errorf("createReservations: %w", err)
		}

		if err = q.reduce(ctx, items); err != nil {
			return fmt.Errorf("reduce: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("execTx: %w", err)
	}

	return nil
}

func (r *pgRepo) CancelReservation(ctx context.Context, orderID uint64) error {
	if err := r.execTx(ctx, func(q *pgQueries) error {
		items, err := q.removeReservations(ctx, orderID)
		if err != nil {
			return fmt.Errorf("remove reservations: %w", err)
		}

		if err = q.increase(ctx, items); err != nil {
			return fmt.Errorf("increase: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("execTx: %w", err)
	}

	return nil
}

func (r *pgRepo) Collect(ctx context.Context, orderID uint64) error {
	if err := r.execTx(ctx, func(q *pgQueries) error {
		_, err := q.removeReservations(ctx, orderID)
		if err != nil {
			return fmt.Errorf("remove reservations: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("execTx: %w", err)
	}

	return nil
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

type pgQueries struct {
	db DBTX
}

var getQuantitiesQuery = fmt.Sprintf(`
SELECT product_id, quantity FROM %s WHERE product_id = ANY ($1)
`, quantitiesTable)

func (q *pgQueries) isEnough(ctx context.Context, items []*Item) (bool, error) {
	ids := make([]uint64, 0, len(items))
	m := make(map[uint64]uint64, len(items))
	for _, item := range items {
		ids = append(ids, item.ProductID)
		m[item.ProductID] = item.Quantity
	}

	rows, err := q.db.Query(ctx, getQuantitiesQuery, ids)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w: db query: %v", ErrInternal, err)
	}
	defer rows.Close()

	var pid uint64
	var qnt uint64
	var count uint64

	for rows.Next() {
		if err = rows.Scan(&pid, &qnt); err != nil {
			return false, fmt.Errorf("%w: rows scan: %v", ErrInternal, err)
		}
		if m[pid] > qnt {
			return false, nil
		}

		count++
	}

	if count < uint64(len(m)) {
		return false, nil
	}

	if err = rows.Err(); err != nil {
		return false, fmt.Errorf("%w: rows err: %v", ErrInternal, err)
	}

	return true, nil
}

var increaseQuery = fmt.Sprintf(`
INSERT INTO %s as q (product_id, quantity)
VALUES ($1, $2)
ON CONFLICT (product_id)
DO UPDATE SET quantity = q.quantity + EXCLUDED.quantity
`, quantitiesTable)

func (q *pgQueries) increase(ctx context.Context, items []*Item) error {
	for _, item := range items {
		if _, err := q.db.Exec(ctx, increaseQuery, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
		}
	}

	return nil
}

var reduceQuery = fmt.Sprintf(`
INSERT INTO %s as q (product_id, quantity)
VALUES ($1, $2)
ON CONFLICT (product_id)
DO UPDATE SET quantity = q.quantity - EXCLUDED.quantity
`, quantitiesTable)

func (q *pgQueries) reduce(ctx context.Context, items []*Item) error {
	for _, item := range items {
		if _, err := q.db.Exec(ctx, reduceQuery, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
		}
	}

	return nil
}

var createReservationQuery = fmt.Sprintf(
	"INSERT INTO %s (order_id, product_id, quantity) VALUES ($1, $2, $3)",
	reservationsTable,
)

func (q *pgQueries) createReservations(ctx context.Context, orderID uint64, items []*Item) error {
	for _, item := range items {
		if _, err := q.db.Exec(ctx, createReservationQuery, orderID, item.ProductID, item.Quantity); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.ConstraintName {
				case "reservations_pkey":
					return fmt.Errorf("%w: db exec: %v", ErrFailedPrecondition, err)
				}
			}
			return fmt.Errorf("%w: db exec: %v", ErrInternal, err)
		}
	}

	return nil
}

var removeReservationsQuery = fmt.Sprintf(
	"DELETE FROM %s WHERE order_id = $1 RETURNING product_id, quantity",
	reservationsTable,
)

func (q *pgQueries) removeReservations(ctx context.Context, orderID uint64) ([]*Item, error) {
	rows, err := q.db.Query(ctx, removeReservationsQuery, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: db exec: %v", ErrInternal, err)
	}
	defer rows.Close()

	var items []*Item
	var pid uint64
	var qnt uint64

	for rows.Next() {
		if err = rows.Scan(&pid, &qnt); err != nil {
			return nil, fmt.Errorf("%w: rows scan: %v", ErrInternal, err)
		}
		items = append(items, &Item{
			ProductID: pid,
			Quantity:  qnt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: rows err: %v", ErrInternal, err)
	}

	return items, nil
}
