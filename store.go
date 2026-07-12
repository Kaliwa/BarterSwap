package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Store is the storage layer: it owns the database handle and exposes typed
// query methods. It contains no business rules.
type Store struct {
	db *sql.DB
}

// NewStore opens the database, waits for it to be reachable, applies the schema
// and returns a ready-to-use Store.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture base : %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := waitForDB(ctx, db); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// Close releases the underlying connection pool.
func (s *Store) Close() error { return s.db.Close() }

// waitForDB retries the ping so the API tolerates the database still starting
// up (typical with Docker Compose ordering).
func waitForDB(ctx context.Context, db *sql.DB) error {
	var err error
	for attempt := 0; attempt < 30; attempt++ {
		if err = db.PingContext(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("base injoignable après plusieurs tentatives : %w", err)
}

// schema is idempotent so it can run on every start-up.
const schema = `
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    pseudo     TEXT NOT NULL,
    bio        TEXT NOT NULL DEFAULT '',
    ville      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS skills (
    id      SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nom     TEXT NOT NULL,
    niveau  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exchange_id INTEGER,
    montant     INTEGER NOT NULL,
    type        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_skills_user ON skills(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_tx_user ON credit_transactions(user_id);
`

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migration : %w", err)
	}
	return nil
}
