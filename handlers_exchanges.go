package main

import (
	"context"
	"net/http"
)

// POST /api/exchanges — create an exchange request.
func (s *Server) handleCreateExchange(w http.ResponseWriter, r *http.Request) {
	actor, err := authUserID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	var body struct {
		ServiceID int `json:"service_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondError(w, err)
		return
	}
	if body.ServiceID <= 0 {
		respondError(w, &ValidationError{Field: "service_id", Message: "service_id est obligatoire"})
		return
	}
	ex, err := s.app.RequestExchange(r.Context(), actor, body.ServiceID)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ex)
}

// GET /api/exchanges — list the caller's exchanges (requests + received),
// optionally filtered by ?status=.
func (s *Server) handleListExchanges(w http.ResponseWriter, r *http.Request) {
	actor, err := authUserID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	exchanges, err := s.app.ListExchanges(r.Context(), actor, r.URL.Query().Get("status"))
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchanges)
}

// GET /api/exchanges/{id} — detail, restricted to participants.
func (s *Server) handleGetExchange(w http.ResponseWriter, r *http.Request) {
	actor, id, err := authAndID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	ex, err := s.app.GetExchange(r.Context(), actor, id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

// PUT /api/exchanges/{id}/accept
func (s *Server) handleAcceptExchange(w http.ResponseWriter, r *http.Request) {
	s.runExchangeAction(w, r, s.app.AcceptExchange)
}

// PUT /api/exchanges/{id}/reject
func (s *Server) handleRejectExchange(w http.ResponseWriter, r *http.Request) {
	s.runExchangeAction(w, r, s.app.RejectExchange)
}

// PUT /api/exchanges/{id}/complete
func (s *Server) handleCompleteExchange(w http.ResponseWriter, r *http.Request) {
	s.runExchangeAction(w, r, s.app.CompleteExchange)
}

// PUT /api/exchanges/{id}/cancel
func (s *Server) handleCancelExchange(w http.ResponseWriter, r *http.Request) {
	s.runExchangeAction(w, r, s.app.CancelExchange)
}

// runExchangeAction wires the shared plumbing of the status-change endpoints:
// authenticate, read the id, run the action and return the updated exchange.
func (s *Server) runExchangeAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, actorID, id int) (Exchange, error)) {
	actor, id, err := authAndID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	ex, err := action(r.Context(), actor, id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

// authAndID resolves both the caller identity and the {id} path parameter.
func authAndID(r *http.Request) (int, int, error) {
	actor, err := authUserID(r)
	if err != nil {
		return 0, 0, err
	}
	id, err := pathID(r)
	if err != nil {
		return 0, 0, err
	}
	return actor, id, nil
}
