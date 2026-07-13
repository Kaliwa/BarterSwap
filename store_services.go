package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// rowScanner is satisfied by both *sql.Row and *sql.Rows, so a single scan
// helper serves single-row and multi-row queries (interface defined where it
// is consumed).
type rowScanner interface {
	Scan(dest ...any) error
}

const serviceColumns = `id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at`

func scanService(sc rowScanner) (Service, error) {
	var svc Service
	var createdAt time.Time
	if err := sc.Scan(
		&svc.ID, &svc.ProviderID, &svc.Titre, &svc.Description, &svc.Categorie,
		&svc.DureeMinutes, &svc.Credits, &svc.Ville, &svc.Actif, &createdAt,
	); err != nil {
		return Service{}, err
	}
	svc.CreatedAt = createdAt.Format(time.RFC3339)
	return svc, nil
}

func (s *Store) CreateService(ctx context.Context, in Service) (Service, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING `+serviceColumns,
		in.ProviderID, in.Titre, in.Description, in.Categorie, in.DureeMinutes, in.Credits, in.Ville, in.Actif)
	created, err := scanService(row)
	if err != nil {
		return Service{}, fmt.Errorf("insertion service : %w", err)
	}
	return created, nil
}

func (s *Store) GetService(ctx context.Context, id int) (Service, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+serviceColumns+` FROM services WHERE id = $1`, id)
	svc, err := scanService(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Service{}, ErrNotFound
	}
	if err != nil {
		return Service{}, err
	}
	return svc, nil
}

func (s *Store) UpdateService(ctx context.Context, in Service) (Service, error) {
	row := s.db.QueryRowContext(ctx,
		`UPDATE services
		 SET titre = $1, description = $2, categorie = $3, duree_minutes = $4, credits = $5, ville = $6, actif = $7
		 WHERE id = $8
		 RETURNING `+serviceColumns,
		in.Titre, in.Description, in.Categorie, in.DureeMinutes, in.Credits, in.Ville, in.Actif, in.ID)
	updated, err := scanService(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Service{}, ErrNotFound
	}
	if err != nil {
		return Service{}, err
	}
	return updated, nil
}

func (s *Store) DeleteService(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListServices applies the optional filters server-side and returns the
// matching services, newest first.
func (s *Store) ListServices(ctx context.Context, f ServiceFilter) ([]Service, error) {
	query := `SELECT ` + serviceColumns + ` FROM services`
	var conds []string
	var args []any
	n := 1

	if v := strings.TrimSpace(f.Categorie); v != "" {
		conds = append(conds, fmt.Sprintf("LOWER(categorie) = LOWER($%d)", n))
		args = append(args, v)
		n++
	}
	if v := strings.TrimSpace(f.Ville); v != "" {
		conds = append(conds, fmt.Sprintf("LOWER(ville) = LOWER($%d)", n))
		args = append(args, v)
		n++
	}
	if v := strings.TrimSpace(f.Search); v != "" {
		conds = append(conds, fmt.Sprintf("(titre ILIKE $%d OR description ILIKE $%d)", n, n))
		args = append(args, "%"+v+"%")
		n++
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []Service{}
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, rows.Err()
}
