package billing

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	paymentsTable        = "payments"
	paymentStatusesTable = "payments_statuses"
)

type Repository interface {
	AddPayment(ctx context.Context, orderID uint64, userID uint64, total float64) error
	GetPayment(ctx context.Context, orderID uint64) (*Payment, error)
	ApprovePayment(ctx context.Context, orderID uint64) (*Payment, error)
	CancelPayment(ctx context.Context, orderID uint64) error
}

type pgRepo struct {
	dbMaster  *pgxpool.Pool
	dbReplica *pgxpool.Pool
}

func NewPgRepo(dbMaster *pgxpool.Pool, dbReplica *pgxpool.Pool) *pgRepo {
	return &pgRepo{
		dbMaster:  dbMaster,
		dbReplica: dbReplica,
	}
}

func (r *pgRepo) AddPayment(ctx context.Context, orderID uint64, userID uint64, total float64) error {
	if err := r.execTx(ctx, r.dbMaster, func(q *pgQueries) error {
		if err := q.createPayment(ctx, orderID, userID, total); err != nil {
			return fmt.Errorf("create payment: %w", err)
		}

		if err := q.createStatus(ctx, orderID); err != nil {
			return fmt.Errorf("create status: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("execTx: %w", err)
	}

	return nil
}

func (r *pgRepo) GetPayment(ctx context.Context, orderID uint64) (*Payment, error) {
	q := &pgQueries{db: r.dbReplica}
	return q.getPayment(ctx, orderID)
}

func (r *pgRepo) ApprovePayment(ctx context.Context, orderID uint64) (*Payment, error) {
	var p *Payment
	if err := r.execTx(ctx, r.dbMaster, func(q *pgQueries) error {
		var err error

		if err = q.updateStatus(ctx, orderID, Paid); err != nil {
			return fmt.Errorf("update status: %w", err)
		}

		p, err = q.getPayment(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get payment: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("execTx: %w", err)
	}
	return p, nil
}

func (r *pgRepo) CancelPayment(ctx context.Context, orderID uint64) error {
	q := &pgQueries{db: r.dbMaster}
	if err := q.updateStatus(ctx, orderID, Cancelled); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

// execTx creates a database transaction with ReadCommitted isolation level and
// execute provided function in the scope of the transaction.
func (r *pgRepo) execTx(ctx context.Context, db *pgxpool.Pool, fn func(queries *pgQueries) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
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

var createPaymentQuery = fmt.Sprintf(`
INSERT INTO %s
(order_id, user_id, total)
VALUES ($1, $2, $3)
`, paymentsTable)

func (q *pgQueries) createPayment(ctx context.Context, orderID uint64, userID uint64, total float64) error {
	if _, err := q.db.Exec(ctx, createPaymentQuery, orderID, userID, total); err != nil {
		return fmt.Errorf("%w: dbMaster exec: %v", ErrInternal, err)
	}

	return nil
}

var getPaymentQuery = fmt.Sprintf(`
SELECT total
FROM %s
WHERE order_id = $1
`, paymentsTable)

func (q *pgQueries) getPayment(ctx context.Context, orderID uint64) (*Payment, error) {
	p := Payment{OrderID: orderID}

	if err := q.db.QueryRow(ctx, getPaymentQuery, orderID).Scan(&p.Total); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: dbMaster query row: %v", ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: dbMaster query row: %v", ErrInternal, err)
	}

	return &p, nil
}

var createStatusQuery = fmt.Sprintf(`
INSERT INTO %s
(order_id)
VALUES ($1)
`, paymentStatusesTable)

func (q *pgQueries) createStatus(ctx context.Context, orderID uint64) error {
	if _, err := q.db.Exec(ctx, createStatusQuery, orderID); err != nil {
		return fmt.Errorf("%w: dbMaster exec: %v", ErrInternal, err)
	}

	return nil
}

var updateStatusQuery = fmt.Sprintf(`
UPDATE %s
SET status = $2
WHERE order_id = $1
`, paymentStatusesTable)

func (q *pgQueries) updateStatus(ctx context.Context, orderID uint64, status PaymentStatus) error {
	if _, err := q.db.Exec(ctx, updateStatusQuery, orderID, status); err != nil {
		return fmt.Errorf("%w: dbMaster exec: %v", ErrInternal, err)
	}

	return nil
}
