package main

import "net/http"

// POST /api/exchanges/{id}/review — rate a completed exchange (requester only).
func (s *Server) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	actor, id, err := authAndID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	var input Review
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, err)
		return
	}
	review, err := s.app.CreateReview(r.Context(), actor, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, review)
}

// GET /api/users/{id}/reviews — reviews received by a user.
func (s *Server) handleListUserReviews(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	reviews, err := s.app.ListUserReviews(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}

// GET /api/services/{id}/reviews — reviews posted on a service.
func (s *Server) handleListServiceReviews(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	reviews, err := s.app.ListServiceReviews(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}
