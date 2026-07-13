package main

import "net/http"

// GET /api/services — list with optional server-side filters.
func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	filter := ServiceFilter{
		Categorie: r.URL.Query().Get("categorie"),
		Ville:     r.URL.Query().Get("ville"),
		Search:    r.URL.Query().Get("search"),
	}
	services, err := s.app.ListServices(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

// POST /api/services — create an announcement.
func (s *Server) handleCreateService(w http.ResponseWriter, r *http.Request) {
	actor, err := authUserID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	var input Service
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, err)
		return
	}
	service, err := s.app.CreateService(r.Context(), actor, input)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, service)
}

// GET /api/services/{id} — announcement detail.
func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	service, err := s.app.GetService(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, service)
}

// PUT /api/services/{id} — edit own announcement.
func (s *Server) handleUpdateService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	actor, err := authUserID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	var input Service
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, err)
		return
	}
	service, err := s.app.UpdateService(r.Context(), actor, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, service)
}

// DELETE /api/services/{id} — remove own announcement.
func (s *Server) handleDeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	actor, err := authUserID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	if err := s.app.DeleteService(r.Context(), actor, id); err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
