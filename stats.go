package main

import "context"

// GetUserStats returns the activity summary of an existing user.
func (a *App) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	exists, err := a.store.UserExists(ctx, userID)
	if err != nil {
		return UserStats{}, err
	}
	if !exists {
		return UserStats{}, ErrNotFound
	}
	return a.store.GetUserStats(ctx, userID)
}
