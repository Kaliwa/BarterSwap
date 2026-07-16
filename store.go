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

CREATE TABLE IF NOT EXISTS services (
    id            SERIAL PRIMARY KEY,
    provider_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titre         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    categorie     TEXT NOT NULL,
    duree_minutes INTEGER NOT NULL,
    credits       INTEGER NOT NULL,
    ville         TEXT NOT NULL DEFAULT '',
    actif         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_skills_user ON skills(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_tx_user ON credit_transactions(user_id);
CREATE TABLE IF NOT EXISTS exchanges (
    id           SERIAL PRIMARY KEY,
    service_id   INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    requester_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    owner_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_services_provider ON services(provider_id);
CREATE INDEX IF NOT EXISTS idx_services_categorie ON services(categorie);
CREATE INDEX IF NOT EXISTS idx_services_ville ON services(ville);
CREATE INDEX IF NOT EXISTS idx_exchanges_requester ON exchanges(requester_id);
CREATE INDEX IF NOT EXISTS idx_exchanges_owner ON exchanges(owner_id);

-- Enforces "one active exchange per service" (rule 2) at the database level,
-- so concurrency is handled without any mutex.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_active_exchange_per_service
    ON exchanges(service_id) WHERE status IN ('pending', 'accepted');

-- UNIQUE on exchange_id enforces "one review per exchange" at the database
-- level; the CHECK keeps the note in the 1-5 range even outside the API.
CREATE TABLE IF NOT EXISTS reviews (
    id          SERIAL PRIMARY KEY,
    exchange_id INTEGER NOT NULL UNIQUE REFERENCES exchanges(id) ON DELETE CASCADE,
    service_id  INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewee_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note        INTEGER NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reviews_reviewee ON reviews(reviewee_id);
CREATE INDEX IF NOT EXISTS idx_reviews_service ON reviews(service_id);
`

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migration : %w", err)
	}
	return nil
}
