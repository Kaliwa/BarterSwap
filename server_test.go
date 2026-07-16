package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// newTestServer builds the full HTTP stack (router + middlewares) on top of a
// fake store, so tests exercise routing, handlers and business logic together.
func newTestServer(store Storer) *Server {
	return NewServer(NewApp(store))
}

// doRequest performs an in-memory HTTP request against the server. body may be
// nil, a raw string (sent as-is, useful for malformed JSON) or any value to be
// JSON-encoded. userID > 0 sets the X-User-ID header.
func doRequest(t *testing.T, srv http.Handler, method, path string, userID int, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(b)
	default:
		buf, err := json.Marshal(b)
		if err != nil {
			t.Fatalf("encodage du corps de requête : %v", err)
		}
		reader = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, path, reader)
	if userID > 0 {
		req.Header.Set("X-User-ID", strconv.Itoa(userID))
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

// decodeBody parses the JSON response into dst.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("réponse JSON invalide : %v (corps : %s)", err, rec.Body.String())
	}
}

// assertStatus fails the test when the recorded status differs.
func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("statut = %d, attendu %d (corps : %s)", rec.Code, want, rec.Body.String())
	}
}

func TestHealth(t *testing.T) {
	rec := doRequest(t, newTestServer(&fakeStore{}), http.MethodGet, "/health", 0, nil)
	assertStatus(t, rec, http.StatusOK)

	var body map[string]string
	decodeBody(t, rec, &body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, attendu %q", body["status"], "ok")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, attendu application/json", ct)
	}
}

func TestCORSPreflight(t *testing.T) {
	rec := doRequest(t, newTestServer(&fakeStore{}), http.MethodOptions, "/api/services", 0, nil)
	assertStatus(t, rec, http.StatusNoContent)
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, attendu *", origin)
	}
	if methods := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "PUT") {
		t.Errorf("Access-Control-Allow-Methods = %q, PUT manquant", methods)
	}
}

func TestRecoverMiddlewareTurnsPanicInto500(t *testing.T) {
	// getUser is left nil: the fake panics, the middleware must answer 500.
	rec := doRequest(t, newTestServer(&fakeStore{}), http.MethodGet, "/api/users/1", 0, nil)
	assertStatus(t, rec, http.StatusInternalServerError)

	var body map[string]string
	decodeBody(t, rec, &body)
	if body["error"] != "erreur interne" {
		t.Errorf("error = %q, attendu message générique", body["error"])
	}
}

func TestInternalErrorIsMasked(t *testing.T) {
	store := &fakeStore{
		getUser: func(_ context.Context, id int) (User, error) { return User{}, errors.New("détail interne secret") },
	}
	rec := doRequest(t, newTestServer(store), http.MethodGet, "/api/users/1", 0, nil)
	assertStatus(t, rec, http.StatusInternalServerError)
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("le détail interne ne doit pas fuiter : %s", rec.Body.String())
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	rec := doRequest(t, newTestServer(&fakeStore{}), http.MethodGet, "/api/inconnu", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestPathIDValidation(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"non numérique", "/api/users/abc"},
		{"zéro", "/api/users/0"},
		{"négatif", "/api/users/-3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, newTestServer(&fakeStore{}), http.MethodGet, tt.path, 0, nil)
			assertStatus(t, rec, http.StatusBadRequest)
		})
	}
}

func TestAuthUserIDValidation(t *testing.T) {
	srv := newTestServer(&fakeStore{})

	// Absent header.
	rec := doRequest(t, srv, http.MethodPost, "/api/exchanges", 0, map[string]int{"service_id": 1})
	assertStatus(t, rec, http.StatusUnauthorized)

	// Malformed header.
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", strings.NewReader(`{"service_id":1}`))
	req.Header.Set("X-User-ID", "pas-un-nombre")
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("statut = %d, attendu 401", rec2.Code)
	}
}
