package main

import "context"

// GetUserStats aggregates a user's activity in a single query. Amounts come
// from the credit journal: earned counts only exchange transfers (welcome
// credits excluded), spent is the blocked amount net of refunds. The average
// rating is rounded to two decimals and 0 when the user has no review.
func (s *Store) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	stats := UserStats{UserID: userID}
	err := s.db.QueryRowContext(ctx, `
		SELECT
		    COALESCE((SELECT SUM(montant) FROM credit_transactions WHERE user_id = $1), 0),
		    (SELECT COUNT(*) FROM exchanges
		       WHERE (requester_id = $1 OR owner_id = $1) AND status = 'completed'),
		    (SELECT COUNT(*) FROM services WHERE provider_id = $1 AND actif),
		    COALESCE((SELECT ROUND(AVG(note), 2) FROM reviews WHERE reviewee_id = $1), 0),
		    (SELECT COUNT(*) FROM reviews WHERE reviewee_id = $1),
		    COALESCE((SELECT SUM(montant) FROM credit_transactions
		       WHERE user_id = $1 AND type = 'earn' AND exchange_id IS NOT NULL), 0),
		    COALESCE((SELECT SUM(-montant) FROM credit_transactions
		       WHERE user_id = $1 AND type IN ('spend', 'refund')), 0)`,
		userID,
	).Scan(
		&stats.CreditBalance, &stats.CompletedExchanges, &stats.ActiveServices,
		&stats.AverageRating, &stats.ReviewCount, &stats.TotalEarned, &stats.TotalSpent,
	)
	if err != nil {
		return UserStats{}, err
	}
	return stats, nil
}
