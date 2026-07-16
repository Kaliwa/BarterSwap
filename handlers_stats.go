package main

import "net/http"

// GET /api/users/{id}/stats — a user's activity summary.
func (s *Server) handleGetUserStats(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	stats, err := s.app.GetUserStats(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
