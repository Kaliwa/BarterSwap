package main

import (
	"context"
	"net/http"
	"testing"
)

// jardinier owns the "Jardinage" skill; every test below uses him as provider.
func storeWithJardinageSkill() *fakeStore {
	return &fakeStore{
		getUserSkills: func(_ context.Context, userID int) ([]Skill, error) {
			return []Skill{{Nom: "Jardinage", Niveau: "expert"}}, nil
		},
	}
}

func validServiceBody() map[string]any {
	return map[string]any{
		"titre":         "Taille de haies",
		"categorie":     "Jardinage",
		"duree_minutes": 90,
		"credits":       2,
		"ville":         "Lyon",
	}
}

func TestCreateService(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		mutate     func(body map[string]any)
		wantStatus int
	}{
		{"création valide", 1, func(map[string]any) {}, http.StatusCreated},
		{"sans authentification", 0, func(map[string]any) {}, http.StatusUnauthorized},
		{"titre manquant", 1, func(b map[string]any) { b["titre"] = "  " }, http.StatusBadRequest},
		{"catégorie inconnue", 1, func(b map[string]any) { b["categorie"] = "Magie" }, http.StatusBadRequest},
		{"durée nulle", 1, func(b map[string]any) { b["duree_minutes"] = 0 }, http.StatusBadRequest},
		{"crédits négatifs", 1, func(b map[string]any) { b["credits"] = -1 }, http.StatusBadRequest},
		{"compétence non déclarée", 1, func(b map[string]any) { b["categorie"] = "Cuisine" }, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storeWithJardinageSkill()
			store.createService = func(_ context.Context, in Service) (Service, error) {
				if !in.Actif {
					t.Error("un service créé doit être actif")
				}
				in.ID = 10
				return in, nil
			}
			body := validServiceBody()
			tt.mutate(body)
			rec := doRequest(t, newTestServer(store), http.MethodPost, "/api/services", tt.userID, body)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestGetService(t *testing.T) {
	store := &fakeStore{
		getService: func(_ context.Context, id int) (Service, error) {
			if id != 10 {
				return Service{}, ErrNotFound
			}
			return Service{ID: 10, Titre: "Taille de haies"}, nil
		},
	}
	srv := newTestServer(store)

	rec := doRequest(t, srv, http.MethodGet, "/api/services/10", 0, nil)
	assertStatus(t, rec, http.StatusOK)

	rec = doRequest(t, srv, http.MethodGet, "/api/services/999", 0, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestListServicesForwardsFilters(t *testing.T) {
	var gotFilter ServiceFilter
	store := &fakeStore{
		listServices: func(_ context.Context, f ServiceFilter) ([]Service, error) {
			gotFilter = f
			return []Service{{ID: 1}}, nil
		},
	}
	rec := doRequest(t, newTestServer(store), http.MethodGet,
		"/api/services?categorie=Jardinage&ville=Lyon&search=haies", 0, nil)
	assertStatus(t, rec, http.StatusOK)

	want := ServiceFilter{Categorie: "Jardinage", Ville: "Lyon", Search: "haies"}
	if gotFilter != want {
		t.Errorf("filtre transmis = %+v, attendu %+v", gotFilter, want)
	}
}

func TestUpdateService(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		wantStatus int
	}{
		{"mise à jour par le propriétaire", 1, http.StatusOK},
		{"mise à jour par un tiers", 2, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storeWithJardinageSkill()
			store.getService = func(_ context.Context, id int) (Service, error) {
				return Service{ID: id, ProviderID: 1}, nil
			}
			store.updateService = func(_ context.Context, in Service) (Service, error) { return in, nil }
			rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/services/10", tt.userID, validServiceBody())
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestUpdateServiceNotFound(t *testing.T) {
	store := &fakeStore{
		getService: func(_ context.Context, id int) (Service, error) { return Service{}, ErrNotFound },
	}
	rec := doRequest(t, newTestServer(store), http.MethodPut, "/api/services/999", 1, validServiceBody())
	assertStatus(t, rec, http.StatusNotFound)
}

func TestDeleteService(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		active     bool
		wantStatus int
	}{
		{"suppression par le propriétaire", 1, false, http.StatusNoContent},
		{"suppression par un tiers", 2, false, http.StatusForbidden},
		{"échange actif en cours", 1, true, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{
				getService: func(_ context.Context, id int) (Service, error) {
					return Service{ID: id, ProviderID: 1}, nil
				},
				hasActiveExchange: func(_ context.Context, serviceID int) (bool, error) { return tt.active, nil },
				deleteService:     func(_ context.Context, id int) error { return nil },
			}
			rec := doRequest(t, newTestServer(store), http.MethodDelete, "/api/services/10", tt.userID, nil)
			assertStatus(t, rec, tt.wantStatus)
		})
	}
}
