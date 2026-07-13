package main

import "net/http"

// POST /api/users — create an account.
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input User
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, err)
		return
	}
	user, err := s.app.RegisterUser(r.Context(), input)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// GET /api/users/{id} — public profile.
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	user, err := s.app.GetUser(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// PUT /api/users/{id} — edit own profile.
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
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
	var input User
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, err)
		return
	}
	user, err := s.app.UpdateUser(r.Context(), actor, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// GET /api/users/{id}/skills — a user's skills.
func (s *Server) handleGetUserSkills(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondError(w, err)
		return
	}
	skills, err := s.app.GetUserSkills(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

// PUT /api/users/{id}/skills — replace own skills.
func (s *Server) handleSetUserSkills(w http.ResponseWriter, r *http.Request) {
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
	var skills []Skill
	if err := decodeJSON(r, &skills); err != nil {
		respondError(w, err)
		return
	}
	updated, err := s.app.SetUserSkills(r.Context(), actor, id, skills)
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
