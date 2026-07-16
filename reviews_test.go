package main

import (
	"context"
	"net/http"
	"testing"
)

// reviewFixture: exchange 5 on service 10 — user 2 requested, user 1 offered.
func reviewFixtureStore(status string) *fakeStore {
	return &fakeStore{
		getExchange: func(_ context.Context, id int) (Exchange, error) {
			return Exchange{ID: id, ServiceID: 10, RequesterID: 2, OwnerID: 1, Status: status}, nil
		},
		createReview: func(_ context.Context, in Review) (Review, error) {
			in.ID = 1
			return in, nil
		},
	}
}

func TestCreateReview(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		status     string
		body       any
		wantStatus int
	}{
		{"avis valide", 2, statusCompleted, map[string]any{"note": 5, "commentaire": " Super ! "}, http.StatusCreated},
		{"sans authentification", 0, statusCompleted, map[string]any{"note": 5}, http.StatusUnauthorized},
		{"par l'offreur", 1, statusCompleted, map[string]any{"note": 5}, http.StatusForbidden},
		{"par un tiers", 3, statusCompleted, map[string]any{"note": 5}, http.StatusForbidden},
		{"échange non terminé", 2, statusAccepted, map[string]any{"note": 5}, http.StatusBadRequest},
		{"échange annulé", 2, statusCancelled, map[string]any{"note": 5}, http.StatusBadRequest},
		{"note trop basse", 2, statusCompleted, map[string]any{"note": 0}, http.StatusBadRequest},
		{"note trop haute", 2, statusCompleted, map[string]any{"note": 6}, http.StatusBadRequest},
		{"JSON invalide", 2, statusCompleted, `{"note":`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := reviewFixtureStore(tt.status)
			rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/exchanges/5/review", tt.userID, tt.body)
			assertStatus(t, rec, tt.wantStatus)

			if tt.wantStatus == http.StatusCreated {
				var got Review
				decodeBody(t, rec, &got)
				if got.ExchangeID != 5 || got.ServiceID != 10 || got.ReviewerID != 2 || got.RevieweeID != 1 {
					t.Errorf("liens de l'avis incorrects : %+v", got)
				}
				if got.Commentaire != "Super !" {
					t.Errorf("commentaire non nettoyé : %q", got.Commentaire)
				}
			}
		})
	}
}

func TestCreateReviewOnlyOncePerExchange(t *testing.T) {
	store := reviewFixtureStore(statusCompleted)
	store.createReview = func(_ context.Context, in Review) (Review, error) { return Review{}, ErrConflict }
	rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/exchanges/5/review", 2, map[string]any{"note": 4})
	assertStatus(t, rec, http.StatusConflict)
}

func TestCreateReviewExchangeNotFound(t *testing.T) {
	store := &fakeStore{
		getExchange: func(_ context.Context, id int) (Exchange, error) { return Exchange{}, ErrNotFound },
	}
	rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/exchanges/999/review", 2, map[string]any{"note": 4})
	assertStatus(t, rec, http.StatusNotFound)
}

func TestReviewIsImmutable(t *testing.T) {
	srv := newTestServer(&fakeStore{})
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		rec := doRequest(t, srv, method, "/api/exchanges/5/review", 2, nil)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /review = %d, aucune route de modification ne doit exister", method, rec.Code)
		}
	}
}

func TestListUserReviews(t *testing.T) {
	store := &fakeStore{
		userExists: func(_ context.Context, id int) (bool, error) { return id == 1, nil },
		listUserReviews: func(_ context.Context, revieweeID int) ([]Review, error) {
			return []Review{{ID: 1, RevieweeID: revieweeID, Note: 5}}, nil
		},
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/users/1/reviews", 0, nil)
	assertStatus(t, rec, http.StatusOK)
	var reviews []Review
	decodeBody(t, rec, &reviews)
	if len(reviews) != 1 || reviews[0].Note != 5 {
		t.Errorf("avis inattendus : %+v", reviews)
	}

	rec = doRequest(t, srv, http.MethodGet, "/api/users/99/reviews", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestListServiceReviews(t *testing.T) {
	store := &fakeStore{
		getService: func(_ context.Context, id int) (Service, error) {
			if id != 10 {
				return Service{}, ErrNotFound
			}
			return Service{ID: 10}, nil
		},
		listServiceReviews: func(_ context.Context, serviceID int) ([]Review, error) {
			return []Review{{ID: 1, ServiceID: serviceID, Note: 4}}, nil
		},
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/services/10/reviews", 0, nil)
	assertStatus(t, rec, http.StatusOK)
	var reviews []Review
	decodeBody(t, rec, &reviews)
	if len(reviews) != 1 || reviews[0].ServiceID != 10 {
		t.Errorf("avis inattendus : %+v", reviews)
	}

	rec = doRequest(t, srv, http.MethodGet, "/api/services/999/reviews", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}
