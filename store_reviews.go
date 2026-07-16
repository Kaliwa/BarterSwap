package main

import (
	"context"
	"fmt"
	"time"
)

const reviewColumns = `id, exchange_id, service_id, reviewer_id, reviewee_id, note, commentaire, created_at`

func scanReview(sc rowScanner) (Review, error) {
	var rv Review
	var createdAt time.Time
	if err := sc.Scan(
		&rv.ID, &rv.ExchangeID, &rv.ServiceID, &rv.ReviewerID, &rv.RevieweeID,
		&rv.Note, &rv.Commentaire, &createdAt,
	); err != nil {
		return Review{}, err
	}
	rv.CreatedAt = createdAt.Format(time.RFC3339)
	return rv, nil
}

// CreateReview inserts a review. The UNIQUE constraint on exchange_id maps a
// second review on the same exchange to ErrConflict.
func (s *Store) CreateReview(ctx context.Context, in Review) (Review, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO reviews (exchange_id, service_id, reviewer_id, reviewee_id, note, commentaire)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+reviewColumns,
		in.ExchangeID, in.ServiceID, in.ReviewerID, in.RevieweeID, in.Note, in.Commentaire)
	created, err := scanReview(row)
	if err != nil {
		if isUniqueViolation(err) {
			return Review{}, ErrConflict
		}
		return Review{}, fmt.Errorf("insertion avis : %w", err)
	}
	return created, nil
}

// ListUserReviews returns the reviews received by a user, newest first.
func (s *Store) ListUserReviews(ctx context.Context, revieweeID int) ([]Review, error) {
	return s.listReviews(ctx, `reviewee_id`, revieweeID)
}

// ListServiceReviews returns the reviews posted on a service, newest first.
func (s *Store) ListServiceReviews(ctx context.Context, serviceID int) ([]Review, error) {
	return s.listReviews(ctx, `service_id`, serviceID)
}

// listReviews factors the shared listing query; column is one of the two
// trusted literals above, never user input.
func (s *Store) listReviews(ctx context.Context, column string, id int) ([]Review, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+reviewColumns+` FROM reviews WHERE `+column+` = $1 ORDER BY created_at DESC, id DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []Review{}
	for rows.Next() {
		rv, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}
