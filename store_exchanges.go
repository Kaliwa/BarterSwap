package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

const exchangeColumns = `id, service_id, requester_id, owner_id, status, created_at, updated_at`

func scanExchange(sc rowScanner) (Exchange, error) {
	var e Exchange
	var createdAt, updatedAt time.Time
	if err := sc.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &createdAt, &updatedAt); err != nil {
		return Exchange{}, err
	}
	e.CreatedAt = createdAt.Format(time.RFC3339)
	e.UpdatedAt = updatedAt.Format(time.RFC3339)
	return e, nil
}

// isUniqueViolation reports whether the error is a Postgres unique-constraint
// violation (SQLSTATE 23505), used to translate the active-exchange index into
// a clean conflict.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// CreateExchange inserts a pending exchange. The partial unique index maps a
// second active exchange on the same service to ErrConflict (rule 2).
func (s *Store) CreateExchange(ctx context.Context, serviceID, requesterID, ownerID int) (Exchange, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO exchanges (service_id, requester_id, owner_id, status)
		 VALUES ($1, $2, $3, 'pending')
		 RETURNING `+exchangeColumns,
		serviceID, requesterID, ownerID)
	ex, err := scanExchange(row)
	if err != nil {
		if isUniqueViolation(err) {
			return Exchange{}, ErrConflict
		}
		return Exchange{}, fmt.Errorf("insertion échange : %w", err)
	}
	return ex, nil
}

func (s *Store) GetExchange(ctx context.Context, id int) (Exchange, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+exchangeColumns+` FROM exchanges WHERE id = $1`, id)
	ex, err := scanExchange(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Exchange{}, ErrNotFound
	}
	if err != nil {
		return Exchange{}, err
	}
	return ex, nil
}

// ListExchanges returns exchanges the user is involved in (as requester or
// owner), optionally filtered by status, newest first.
func (s *Store) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	query := `SELECT ` + exchangeColumns + ` FROM exchanges WHERE (requester_id = $1 OR owner_id = $1)`
	args := []any{userID}
	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exchanges := []Exchange{}
	for rows.Next() {
		ex, err := scanExchange(rows)
		if err != nil {
			return nil, err
		}
		exchanges = append(exchanges, ex)
	}
	return exchanges, rows.Err()
}

// HasActiveExchange reports whether a service currently has a pending or
// accepted exchange.
func (s *Store) HasActiveExchange(ctx context.Context, serviceID int) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM exchanges WHERE service_id = $1 AND status IN ('pending', 'accepted'))`,
		serviceID).Scan(&exists)
	return exists, err
}

// UserBalance returns the credit balance derived from the transaction journal.
func (s *Store) UserBalance(ctx context.Context, userID int) (int, error) {
	var balance int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(montant), 0) FROM credit_transactions WHERE user_id = $1`,
		userID).Scan(&balance)
	return balance, err
}

// AcceptExchange moves a pending exchange to accepted and blocks the credits:
// the requester is debited (their balance is checked under a row lock so two
// concurrent accepts cannot overdraw).
func (s *Store) AcceptExchange(ctx context.Context, id int) (Exchange, error) {
	return s.transition(ctx, id, func(tx *sql.Tx, ex exchangeRow) error {
		if ex.status != statusPending {
			return ErrConflict
		}
		credits, err := serviceCredits(ctx, tx, ex.serviceID)
		if err != nil {
			return err
		}
		// Serialize credit changes for this requester, then verify the balance.
		if _, err := tx.ExecContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, ex.requesterID); err != nil {
			return err
		}
		var balance int
		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(montant), 0) FROM credit_transactions WHERE user_id = $1`,
			ex.requesterID).Scan(&balance); err != nil {
			return err
		}
		if balance < credits {
			return ErrInsufficientCredits
		}
		if err := insertTransaction(ctx, tx, ex.requesterID, id, -credits, txSpend); err != nil {
			return err
		}
		return setStatus(ctx, tx, id, statusAccepted)
	})
}

// RejectExchange moves a pending exchange to rejected. No credits were blocked
// yet, so nothing is refunded.
func (s *Store) RejectExchange(ctx context.Context, id int) (Exchange, error) {
	return s.transition(ctx, id, func(tx *sql.Tx, ex exchangeRow) error {
		if ex.status != statusPending {
			return ErrConflict
		}
		return setStatus(ctx, tx, id, statusRejected)
	})
}

// CompleteExchange moves an accepted exchange to completed and transfers the
// blocked credits to the owner for good.
func (s *Store) CompleteExchange(ctx context.Context, id int) (Exchange, error) {
	return s.transition(ctx, id, func(tx *sql.Tx, ex exchangeRow) error {
		if ex.status != statusAccepted {
			return ErrConflict
		}
		credits, err := serviceCredits(ctx, tx, ex.serviceID)
		if err != nil {
			return err
		}
		if err := insertTransaction(ctx, tx, ex.ownerID, id, credits, txEarn); err != nil {
			return err
		}
		return setStatus(ctx, tx, id, statusCompleted)
	})
}

// CancelExchange cancels a pending or accepted exchange, refunding the
// requester if the credits had been blocked.
func (s *Store) CancelExchange(ctx context.Context, id int) (Exchange, error) {
	return s.transition(ctx, id, func(tx *sql.Tx, ex exchangeRow) error {
		if ex.status != statusPending && ex.status != statusAccepted {
			return ErrConflict
		}
		if ex.status == statusAccepted {
			credits, err := serviceCredits(ctx, tx, ex.serviceID)
			if err != nil {
				return err
			}
			if err := insertTransaction(ctx, tx, ex.requesterID, id, credits, txRefund); err != nil {
				return err
			}
		}
		return setStatus(ctx, tx, id, statusCancelled)
	})
}

// exchangeRow holds the locked exchange fields a transition operates on.
type exchangeRow struct {
	serviceID   int
	requesterID int
	ownerID     int
	status      string
}

// transition runs fn inside a transaction after locking the exchange row, then
// returns the fresh exchange. It centralizes the boilerplate shared by all
// status changes.
func (s *Store) transition(ctx context.Context, id int, fn func(tx *sql.Tx, ex exchangeRow) error) (Exchange, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var ex exchangeRow
	err = tx.QueryRowContext(ctx,
		`SELECT service_id, requester_id, owner_id, status FROM exchanges WHERE id = $1 FOR UPDATE`, id).
		Scan(&ex.serviceID, &ex.requesterID, &ex.ownerID, &ex.status)
	if errors.Is(err, sql.ErrNoRows) {
		return Exchange{}, ErrNotFound
	}
	if err != nil {
		return Exchange{}, err
	}

	if err := fn(tx, ex); err != nil {
		return Exchange{}, err
	}
	if err := tx.Commit(); err != nil {
		return Exchange{}, err
	}
	return s.GetExchange(ctx, id)
}

func serviceCredits(ctx context.Context, tx *sql.Tx, serviceID int) (int, error) {
	var credits int
	if err := tx.QueryRowContext(ctx, `SELECT credits FROM services WHERE id = $1`, serviceID).Scan(&credits); err != nil {
		return 0, err
	}
	return credits, nil
}

func insertTransaction(ctx context.Context, tx *sql.Tx, userID, exchangeID, montant int, txType string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES ($1, $2, $3, $4)`,
		userID, exchangeID, montant, txType)
	return err
}

func setStatus(ctx context.Context, tx *sql.Tx, id int, status string) error {
	_, err := tx.ExecContext(ctx, `UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}
