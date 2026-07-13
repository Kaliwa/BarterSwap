package main

import "context"

// RequestExchange creates a pending exchange after enforcing the entry rules:
// the service must exist and be active, a user cannot request their own
// service, and they must have enough credits. Rule 2 (one active exchange per
// service) is enforced by the store.
func (a *App) RequestExchange(ctx context.Context, requesterID, serviceID int) (Exchange, error) {
	service, err := a.store.GetService(ctx, serviceID)
	if err != nil {
		return Exchange{}, err
	}
	if !service.Actif {
		return Exchange{}, &ValidationError{Field: "service_id", Message: "ce service n'est pas disponible"}
	}
	if service.ProviderID == requesterID {
		return Exchange{}, &ValidationError{Field: "service_id", Message: "on ne peut pas s'échanger son propre service"}
	}

	balance, err := a.store.UserBalance(ctx, requesterID)
	if err != nil {
		return Exchange{}, err
	}
	if balance < service.Credits {
		return Exchange{}, ErrInsufficientCredits
	}

	return a.store.CreateExchange(ctx, serviceID, requesterID, service.ProviderID)
}

// GetExchange returns an exchange, restricted to its participants.
func (a *App) GetExchange(ctx context.Context, actorID, id int) (Exchange, error) {
	ex, err := a.store.GetExchange(ctx, id)
	if err != nil {
		return Exchange{}, err
	}
	if !isParticipant(ex, actorID) {
		return Exchange{}, ErrForbidden
	}
	return ex, nil
}

// ListExchanges returns the exchanges the user is involved in.
func (a *App) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	return a.store.ListExchanges(ctx, userID, status)
}

// AcceptExchange accepts a pending request; only the owner may accept.
func (a *App) AcceptExchange(ctx context.Context, actorID, id int) (Exchange, error) {
	ex, err := a.store.GetExchange(ctx, id)
	if err != nil {
		return Exchange{}, err
	}
	if actorID != ex.OwnerID {
		return Exchange{}, ErrForbidden
	}
	return a.store.AcceptExchange(ctx, id)
}

// RejectExchange rejects a pending request; only the owner may reject.
func (a *App) RejectExchange(ctx context.Context, actorID, id int) (Exchange, error) {
	ex, err := a.store.GetExchange(ctx, id)
	if err != nil {
		return Exchange{}, err
	}
	if actorID != ex.OwnerID {
		return Exchange{}, ErrForbidden
	}
	return a.store.RejectExchange(ctx, id)
}

// CompleteExchange marks an accepted exchange as completed; either participant
// may confirm completion.
func (a *App) CompleteExchange(ctx context.Context, actorID, id int) (Exchange, error) {
	ex, err := a.store.GetExchange(ctx, id)
	if err != nil {
		return Exchange{}, err
	}
	if !isParticipant(ex, actorID) {
		return Exchange{}, ErrForbidden
	}
	return a.store.CompleteExchange(ctx, id)
}

// CancelExchange cancels a pending or accepted exchange; either participant may
// cancel.
func (a *App) CancelExchange(ctx context.Context, actorID, id int) (Exchange, error) {
	ex, err := a.store.GetExchange(ctx, id)
	if err != nil {
		return Exchange{}, err
	}
	if !isParticipant(ex, actorID) {
		return Exchange{}, ErrForbidden
	}
	return a.store.CancelExchange(ctx, id)
}

func isParticipant(ex Exchange, userID int) bool {
	return userID == ex.RequesterID || userID == ex.OwnerID
}
