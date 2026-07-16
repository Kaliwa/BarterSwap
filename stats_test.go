package main

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestGetUserStats(t *testing.T) {
	want := UserStats{
		UserID:             1,
		CreditBalance:      14,
		CompletedExchanges: 3,
		ActiveServices:     2,
		AverageRating:      4.5,
		ReviewCount:        2,
		TotalEarned:        8,
		TotalSpent:         4,
	}
	store := &fakeStore{
		userExists:   func(_ context.Context, id int) (bool, error) { return id == 1, nil },
		getUserStats: func(_ context.Context, userID int) (UserStats, error) { return want, nil },
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/users/1/stats", 0, nil)
	assertStatus(t, rec, http.StatusOK)
	var got UserStats
	decodeBody(t, rec, &got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stats = %+v, attendu %+v", got, want)
	}

	rec = doRequest(t, srv, http.MethodGet, "/api/users/99/stats", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)

	rec = doRequest(t, srv, http.MethodGet, "/api/users/abc/stats", 0, nil)
	assertStatus(t, rec, http.StatusBadRequest)
}
