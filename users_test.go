package main

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"création valide", map[string]string{"pseudo": "alice", "bio": "  Jardinière  ", "ville": "Lyon"}, http.StatusCreated},
		{"pseudo manquant", map[string]string{"bio": "sans pseudo"}, http.StatusBadRequest},
		{"pseudo espaces uniquement", map[string]string{"pseudo": "   "}, http.StatusBadRequest},
		{"JSON invalide", `{"pseudo":`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{
				createUser: func(_ context.Context, u User) (User, error) {
					if u.Pseudo != "alice" || u.Bio != "Jardinière" {
						t.Errorf("champs non nettoyés : %+v", u)
					}
					u.ID = 1
					u.CreditBalance = welcomeCredits
					return u, nil
				},
			}
			rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/users", 0, tt.body)
			assertStatus(t, rec, tt.wantStatus)

			if tt.wantStatus == http.StatusCreated {
				var got User
				decodeBody(t, rec, &got)
				if got.CreditBalance != welcomeCredits {
					t.Errorf("credit_balance = %d, attendu %d", got.CreditBalance, welcomeCredits)
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	store := &fakeStore{
		getUser: func(_ context.Context, id int) (User, error) {
			if id != 7 {
				return User{}, ErrNotFound
			}
			return User{ID: 7, Pseudo: "bob", CreditBalance: 12}, nil
		},
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/users/7", 0, nil)
	assertStatus(t, rec, http.StatusOK)
	var got User
	decodeBody(t, rec, &got)
	if got.Pseudo != "bob" || got.CreditBalance != 12 {
		t.Errorf("profil inattendu : %+v", got)
	}

	rec = doRequest(t, srv, http.MethodGet, "/api/users/99", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		body       any
		wantStatus int
	}{
		{"mise à jour valide", 3, map[string]string{"pseudo": "carla", "ville": "Nice"}, http.StatusOK},
		{"sans authentification", 0, map[string]string{"pseudo": "carla"}, http.StatusUnauthorized},
		{"profil d'un autre utilisateur", 8, map[string]string{"pseudo": "carla"}, http.StatusForbidden},
		{"pseudo vide", 3, map[string]string{"pseudo": ""}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{
				updateUser: func(_ context.Context, u User) (User, error) { return u, nil },
			}
			rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/users/3", tt.userID, tt.body)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetUserSkills(t *testing.T) {
	store := &fakeStore{
		userExists: func(_ context.Context, id int) (bool, error) { return id == 4, nil },
		getUserSkills: func(_ context.Context, userID int) ([]Skill, error) {
			return []Skill{{Nom: "Jardinage", Niveau: "expert"}}, nil
		},
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/users/4/skills", 0, nil)
	assertStatus(t, rec, http.StatusOK)
	var skills []Skill
	decodeBody(t, rec, &skills)
	if len(skills) != 1 || skills[0].Nom != "Jardinage" {
		t.Errorf("compétences inattendues : %+v", skills)
	}

	rec = doRequest(t, srv, http.MethodGet, "/api/users/5/skills", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestSetUserSkills(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		body       any
		wantStatus int
	}{
		{"remplacement valide", 4, []Skill{{Nom: " Cuisine ", Niveau: "débutant"}}, http.StatusOK},
		{"liste vide autorisée", 4, []Skill{}, http.StatusOK},
		{"compétences d'un autre", 9, []Skill{{Nom: "Cuisine", Niveau: "débutant"}}, http.StatusForbidden},
		{"nom vide", 4, []Skill{{Nom: "  ", Niveau: "expert"}}, http.StatusBadRequest},
		{"niveau invalide", 4, []Skill{{Nom: "Cuisine", Niveau: "dieu"}}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{
				userExists: func(_ context.Context, id int) (bool, error) { return true, nil },
				setUserSkills: func(_ context.Context, userID int, skills []Skill) error {
					for _, sk := range skills {
						if sk.Nom != "Cuisine" {
							t.Errorf("nom non nettoyé : %q", sk.Nom)
						}
					}
					return nil
				},
				getUserSkills: func(_ context.Context, userID int) ([]Skill, error) { return []Skill{}, nil },
			}
			rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/users/4/skills", tt.userID, tt.body)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestSetUserSkillsUserNotFound(t *testing.T) {
	store := &fakeStore{
		userExists: func(_ context.Context, id int) (bool, error) { return false, nil },
	}
	rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/users/4/skills", 4, []Skill{})
	assertStatus(t, rec, http.StatusNotFound)
}
