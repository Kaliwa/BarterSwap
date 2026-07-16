package main

import (
	"context"
	"net/http"
	"testing"
)

// exchangeFixture: the service 10 (2 credits) belongs to user 1; user 2 asks.
func exchangeFixtureStore() *fakeStore {
	return &fakeStore{
		getService: func(_ context.Context, id int) (Service, error) {
			return Service{ID: id, ProviderID: 1, Credits: 2, Actif: true}, nil
		},
		userBalance: func(_ context.Context, userID int) (int, error) { return 10, nil },
		createExchange: func(_ context.Context, serviceID, requesterID, ownerID int) (Exchange, error) {
			return Exchange{ID: 1, ServiceID: serviceID, RequesterID: requesterID, OwnerID: ownerID, Status: statusPending}, nil
		},
	}
}

func TestRequestExchange(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		body       any
		mutate     func(store *fakeStore)
		wantStatus int
	}{
		{
			"demande valide", 2, map[string]int{"service_id": 10},
			func(*fakeStore) {}, http.StatusCreated,
		},
		{
			"sans authentification", 0, map[string]int{"service_id": 10},
			func(*fakeStore) {}, http.StatusUnauthorized,
		},
		{
			"service_id manquant", 2, map[string]int{},
			func(*fakeStore) {}, http.StatusBadRequest,
		},
		{
			"service introuvable", 2, map[string]int{"service_id": 10},
			func(s *fakeStore) {
				s.getService = func(_ context.Context, id int) (Service, error) { return Service{}, ErrNotFound }
			},
			http.StatusNotFound,
		},
		{
			"service inactif", 2, map[string]int{"service_id": 10},
			func(s *fakeStore) {
				s.getService = func(_ context.Context, id int) (Service, error) {
					return Service{ID: id, ProviderID: 1, Credits: 2, Actif: false}, nil
				}
			},
			http.StatusBadRequest,
		},
		{
			"son propre service", 1, map[string]int{"service_id": 10},
			func(*fakeStore) {}, http.StatusBadRequest,
		},
		{
			"crédits insuffisants", 2, map[string]int{"service_id": 10},
			func(s *fakeStore) {
				s.userBalance = func(_ context.Context, userID int) (int, error) { return 1, nil }
			},
			http.StatusBadRequest,
		},
		{
			"échange actif déjà présent", 2, map[string]int{"service_id": 10},
			func(s *fakeStore) {
				s.createExchange = func(_ context.Context, serviceID, requesterID, ownerID int) (Exchange, error) {
					return Exchange{}, ErrConflict
				}
			},
			http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := exchangeFixtureStore()
			tt.mutate(store)
			rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/exchanges", tt.userID, tt.body)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetExchangeRestrictedToParticipants(t *testing.T) {
	store := &fakeStore{
		getExchange: func(_ context.Context, id int) (Exchange, error) {
			return Exchange{ID: id, RequesterID: 2, OwnerID: 1, Status: statusPending}, nil
		},
	}
	srv := newTestServer(store)

	for _, userID := range []int{1, 2} {
		rec := doRequest(t, srv, http.MethodGet, "/api/exchanges/1", userID, nil)
		assertStatus(t, rec, http.StatusOK)
	}

	rec := doRequest(t, srv, http.MethodGet, "/api/exchanges/1", 3, nil)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestListExchangesForwardsStatusFilter(t *testing.T) {
	var gotUser int
	var gotStatus string
	store := &fakeStore{
		listExchanges: func(_ context.Context, userID int, status string) ([]Exchange, error) {
			gotUser, gotStatus = userID, status
			return []Exchange{}, nil
		},
	}
	rec := doRequest(t, newTestServer(store), http.MethodGet, "/api/exchanges?status=pending", 2, nil)
	assertStatus(t, rec, http.StatusOK)
	if gotUser != 2 || gotStatus != "pending" {
		t.Errorf("appel store = (user %d, status %q), attendu (2, pending)", gotUser, gotStatus)
	}
}

// TestExchangeTransitions covers the four status-change endpoints and their
// permission rules: accept/reject are owner-only, complete/cancel are open to
// both participants.
func TestExchangeTransitions(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		userID     int
		wantStatus int
	}{
		{"accept par l'offreur", "accept", 1, http.StatusOK},
		{"accept par le demandeur", "accept", 2, http.StatusForbidden},
		{"reject par l'offreur", "reject", 1, http.StatusOK},
		{"reject par un tiers", "reject", 3, http.StatusForbidden},
		{"complete par le demandeur", "complete", 2, http.StatusOK},
		{"complete par l'offreur", "complete", 1, http.StatusOK},
		{"complete par un tiers", "complete", 3, http.StatusForbidden},
		{"cancel par le demandeur", "cancel", 2, http.StatusOK},
		{"cancel par un tiers", "cancel", 3, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transition := func(_ context.Context, id int) (Exchange, error) {
				return Exchange{ID: id, RequesterID: 2, OwnerID: 1}, nil
			}
			store := &fakeStore{
				getExchange: func(_ context.Context, id int) (Exchange, error) {
					return Exchange{ID: id, RequesterID: 2, OwnerID: 1, Status: statusPending}, nil
				},
				acceptExchange:   transition,
				rejectExchange:   transition,
				completeExchange: transition,
				cancelExchange:   transition,
			}
			rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/exchanges/1/"+tt.action, tt.userID, nil)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

// TestExchangeTransitionConflict checks that a store-level state conflict
// (e.g. accepting a non-pending exchange) surfaces as HTTP 409.
func TestExchangeTransitionConflict(t *testing.T) {
	store := &fakeStore{
		getExchange: func(_ context.Context, id int) (Exchange, error) {
			return Exchange{ID: id, RequesterID: 2, OwnerID: 1, Status: statusCompleted}, nil
		},
		acceptExchange: func(_ context.Context, id int) (Exchange, error) { return Exchange{}, ErrConflict },
	}
	rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/exchanges/1/accept", 1, nil)
	assertStatus(t, rec, http.StatusConflict)
}
