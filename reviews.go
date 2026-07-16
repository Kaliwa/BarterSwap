package main

import (
	"context"
	"fmt"
	"strings"
)

// CreateReview posts the requester's rating on a completed exchange. Rules:
// only the requester may review (the review targets the owner), the exchange
// must be completed, the note is 1-5, and there is at most one review per
// exchange (enforced by the store). Reviews are immutable once created.
func (a *App) CreateReview(ctx context.Context, actorID, exchangeID int, input Review) (Review, error) {
	ex, err := a.store.GetExchange(ctx, exchangeID)
	if err != nil {
		return Review{}, err
	}
	if actorID != ex.RequesterID {
		return Review{}, ErrForbidden
	}
	if ex.Status != statusCompleted {
		return Review{}, &ValidationError{Field: "status", Message: "seul un échange terminé peut être évalué"}
	}
	if input.Note < minNote || input.Note > maxNote {
		return Review{}, &ValidationError{Field: "note", Message: fmt.Sprintf("la note doit être comprise entre %d et %d", minNote, maxNote)}
	}
	return a.store.CreateReview(ctx, Review{
		ExchangeID:  exchangeID,
		ServiceID:   ex.ServiceID,
		ReviewerID:  actorID,
		RevieweeID:  ex.OwnerID,
		Note:        input.Note,
		Commentaire: strings.TrimSpace(input.Commentaire),
	})
}

// ListUserReviews returns the reviews received by an existing user.
func (a *App) ListUserReviews(ctx context.Context, userID int) ([]Review, error) {
	exists, err := a.store.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	return a.store.ListUserReviews(ctx, userID)
}

// ListServiceReviews returns the reviews posted on an existing service.
func (a *App) ListServiceReviews(ctx context.Context, serviceID int) ([]Review, error) {
	if _, err := a.store.GetService(ctx, serviceID); err != nil {
		return nil, err
	}
	return a.store.ListServiceReviews(ctx, serviceID)
}
