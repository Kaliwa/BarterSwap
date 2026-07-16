package main

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"
)

// newTestStore connects to a dedicated barterswap_test database (created on
// the fly next to the dev one) and truncates every table so each test starts
// from a clean state. The whole test is skipped when PostgreSQL is not
// reachable, so `go test` still passes without the compose stack running.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	baseDSN := env("DATABASE_URL", "postgres://barterswap:barterswap@localhost:5432/barterswap?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	admin, err := sql.Open("postgres", baseDSN)
	if err != nil {
		t.Skipf("base indisponible : %v", err)
	}
	if err := admin.PingContext(ctx); err != nil {
		_ = admin.Close()
		t.Skipf("base indisponible : %v", err)
	}
	// Ignore the "already exists" error: the test database is created once.
	_, _ = admin.ExecContext(ctx, `CREATE DATABASE barterswap_test`)
	_ = admin.Close()

	u, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("DSN invalide : %v", err)
	}
	u.Path = "/barterswap_test"

	store, err := NewStore(ctx, u.String())
	if err != nil {
		t.Fatalf("connexion base de test : %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if _, err := store.db.ExecContext(ctx,
		`TRUNCATE users, skills, credit_transactions, services, exchanges, reviews RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("nettoyage base de test : %v", err)
	}
	return store
}

func createTestUser(t *testing.T, s *Store, pseudo string) User {
	t.Helper()
	u, err := s.CreateUser(context.Background(), User{Pseudo: pseudo, Ville: "Lyon"})
	if err != nil {
		t.Fatalf("création utilisateur %s : %v", pseudo, err)
	}
	return u
}

func createTestService(t *testing.T, s *Store, providerID, credits int) Service {
	t.Helper()
	svc, err := s.CreateService(context.Background(), Service{
		ProviderID:   providerID,
		Titre:        "Taille de haies",
		Categorie:    "Jardinage",
		DureeMinutes: 90,
		Credits:      credits,
		Ville:        "Lyon",
		Actif:        true,
	})
	if err != nil {
		t.Fatalf("création service : %v", err)
	}
	return svc
}

// completedExchange walks a fresh exchange through accept + complete.
func completedExchange(t *testing.T, s *Store, serviceID, requesterID, ownerID int) Exchange {
	t.Helper()
	ctx := context.Background()
	ex, err := s.CreateExchange(ctx, serviceID, requesterID, ownerID)
	if err != nil {
		t.Fatalf("création échange : %v", err)
	}
	if _, err := s.AcceptExchange(ctx, ex.ID); err != nil {
		t.Fatalf("acceptation : %v", err)
	}
	ex, err = s.CompleteExchange(ctx, ex.ID)
	if err != nil {
		t.Fatalf("complétion : %v", err)
	}
	return ex
}

func TestStoreUsers(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	alice := createTestUser(t, store, "alice")
	if alice.CreditBalance != welcomeCredits {
		t.Errorf("solde initial = %d, attendu %d", alice.CreditBalance, welcomeCredits)
	}

	got, err := store.GetUser(ctx, alice.ID)
	if err != nil || got.Pseudo != "alice" || got.CreditBalance != welcomeCredits {
		t.Errorf("GetUser = %+v, %v", got, err)
	}

	if _, err := store.GetUser(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetUser(999) = %v, attendu ErrNotFound", err)
	}

	updated, err := store.UpdateUser(ctx, User{ID: alice.ID, Pseudo: "alice2", Ville: "Paris"})
	if err != nil || updated.Pseudo != "alice2" || updated.Ville != "Paris" {
		t.Errorf("UpdateUser = %+v, %v", updated, err)
	}
	if _, err := store.UpdateUser(ctx, User{ID: 999, Pseudo: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("UpdateUser(999) = %v, attendu ErrNotFound", err)
	}

	exists, err := store.UserExists(ctx, alice.ID)
	if err != nil || !exists {
		t.Errorf("UserExists = %v, %v", exists, err)
	}
	exists, _ = store.UserExists(ctx, 999)
	if exists {
		t.Error("UserExists(999) doit être faux")
	}

	skills := []Skill{{Nom: "Jardinage", Niveau: "expert"}, {Nom: "Cuisine", Niveau: "débutant"}}
	if err := store.SetUserSkills(ctx, alice.ID, skills); err != nil {
		t.Fatalf("SetUserSkills : %v", err)
	}
	got2, err := store.GetUserSkills(ctx, alice.ID)
	if err != nil || len(got2) != 2 || got2[0].Nom != "Jardinage" {
		t.Errorf("GetUserSkills = %+v, %v", got2, err)
	}
	// Replacement is wholesale.
	if err := store.SetUserSkills(ctx, alice.ID, []Skill{{Nom: "Musique", Niveau: "expert"}}); err != nil {
		t.Fatalf("SetUserSkills (remplacement) : %v", err)
	}
	got2, _ = store.GetUserSkills(ctx, alice.ID)
	if len(got2) != 1 || got2[0].Nom != "Musique" {
		t.Errorf("compétences après remplacement = %+v", got2)
	}
}

func TestStoreServices(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	alice := createTestUser(t, store, "alice")

	svc := createTestService(t, store, alice.ID, 2)

	got, err := store.GetService(ctx, svc.ID)
	if err != nil || got.Titre != "Taille de haies" || !got.Actif {
		t.Errorf("GetService = %+v, %v", got, err)
	}
	if _, err := store.GetService(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetService(999) = %v, attendu ErrNotFound", err)
	}

	svc.Titre = "Tonte de pelouse"
	svc.Actif = false
	updated, err := store.UpdateService(ctx, svc)
	if err != nil || updated.Titre != "Tonte de pelouse" || updated.Actif {
		t.Errorf("UpdateService = %+v, %v", updated, err)
	}
	if _, err := store.UpdateService(ctx, Service{ID: 999, Titre: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("UpdateService(999) = %v, attendu ErrNotFound", err)
	}

	// A second service in another city/category to exercise the filters.
	other, err := store.CreateService(ctx, Service{
		ProviderID: alice.ID, Titre: "Cours de guitare", Categorie: "Musique",
		DureeMinutes: 60, Credits: 1, Ville: "Paris", Actif: true,
	})
	if err != nil {
		t.Fatalf("création second service : %v", err)
	}

	filterTests := []struct {
		name    string
		filter  ServiceFilter
		wantIDs []int
	}{
		{"sans filtre", ServiceFilter{}, []int{other.ID, svc.ID}},
		{"par catégorie", ServiceFilter{Categorie: "musique"}, []int{other.ID}},
		{"par ville", ServiceFilter{Ville: "PARIS"}, []int{other.ID}},
		{"par recherche", ServiceFilter{Search: "pelouse"}, []int{svc.ID}},
		{"filtres cumulés sans résultat", ServiceFilter{Categorie: "Musique", Ville: "Lyon"}, []int{}},
	}
	for _, tt := range filterTests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := store.ListServices(ctx, tt.filter)
			if err != nil {
				t.Fatalf("ListServices : %v", err)
			}
			ids := make([]int, 0, len(list))
			for _, s := range list {
				ids = append(ids, s.ID)
			}
			if len(ids) != len(tt.wantIDs) {
				t.Fatalf("ids = %v, attendu %v", ids, tt.wantIDs)
			}
			for i := range ids {
				if ids[i] != tt.wantIDs[i] {
					t.Fatalf("ids = %v, attendu %v", ids, tt.wantIDs)
				}
			}
		})
	}

	if err := store.DeleteService(ctx, svc.ID); err != nil {
		t.Errorf("DeleteService : %v", err)
	}
	if err := store.DeleteService(ctx, svc.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("DeleteService (déjà supprimé) = %v, attendu ErrNotFound", err)
	}
}

func TestStoreExchangeLifecycle(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	owner := createTestUser(t, store, "offreur")
	requester := createTestUser(t, store, "demandeur")
	svc := createTestService(t, store, owner.ID, 3)

	ex, err := store.CreateExchange(ctx, svc.ID, requester.ID, owner.ID)
	if err != nil || ex.Status != statusPending {
		t.Fatalf("CreateExchange = %+v, %v", ex, err)
	}

	// Rule 2: a second active exchange on the same service is a conflict.
	if _, err := store.CreateExchange(ctx, svc.ID, requester.ID, owner.ID); !errors.Is(err, ErrConflict) {
		t.Errorf("second échange actif = %v, attendu ErrConflict", err)
	}

	active, err := store.HasActiveExchange(ctx, svc.ID)
	if err != nil || !active {
		t.Errorf("HasActiveExchange = %v, %v", active, err)
	}

	list, err := store.ListExchanges(ctx, requester.ID, statusPending)
	if err != nil || len(list) != 1 || list[0].ID != ex.ID {
		t.Errorf("ListExchanges(pending) = %+v, %v", list, err)
	}
	list, _ = store.ListExchanges(ctx, requester.ID, statusCompleted)
	if len(list) != 0 {
		t.Errorf("ListExchanges(completed) = %+v, attendu vide", list)
	}

	// Accept blocks the requester's credits.
	ex, err = store.AcceptExchange(ctx, ex.ID)
	if err != nil || ex.Status != statusAccepted {
		t.Fatalf("AcceptExchange = %+v, %v", ex, err)
	}
	balance, _ := store.UserBalance(ctx, requester.ID)
	if balance != welcomeCredits-3 {
		t.Errorf("solde après blocage = %d, attendu %d", balance, welcomeCredits-3)
	}

	// Accepting twice is a state conflict.
	if _, err := store.AcceptExchange(ctx, ex.ID); !errors.Is(err, ErrConflict) {
		t.Errorf("double acceptation = %v, attendu ErrConflict", err)
	}

	// Complete transfers the credits to the owner.
	ex, err = store.CompleteExchange(ctx, ex.ID)
	if err != nil || ex.Status != statusCompleted {
		t.Fatalf("CompleteExchange = %+v, %v", ex, err)
	}
	balance, _ = store.UserBalance(ctx, owner.ID)
	if balance != welcomeCredits+3 {
		t.Errorf("solde offreur après transfert = %d, attendu %d", balance, welcomeCredits+3)
	}

	if _, err := store.GetExchange(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetExchange(999) = %v, attendu ErrNotFound", err)
	}
	if _, err := store.AcceptExchange(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("AcceptExchange(999) = %v, attendu ErrNotFound", err)
	}
}

func TestStoreExchangeRejectAndCancel(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	owner := createTestUser(t, store, "offreur")
	requester := createTestUser(t, store, "demandeur")
	svc := createTestService(t, store, owner.ID, 4)

	// Reject: no credits were blocked, balances stay put.
	ex, err := store.CreateExchange(ctx, svc.ID, requester.ID, owner.ID)
	if err != nil {
		t.Fatalf("CreateExchange : %v", err)
	}
	if ex, err = store.RejectExchange(ctx, ex.ID); err != nil || ex.Status != statusRejected {
		t.Fatalf("RejectExchange = %+v, %v", ex, err)
	}
	if _, err := store.RejectExchange(ctx, ex.ID); !errors.Is(err, ErrConflict) {
		t.Errorf("double refus = %v, attendu ErrConflict", err)
	}

	// Cancel after accept: the blocked credits are refunded.
	ex2, err := store.CreateExchange(ctx, svc.ID, requester.ID, owner.ID)
	if err != nil {
		t.Fatalf("CreateExchange : %v", err)
	}
	if _, err := store.AcceptExchange(ctx, ex2.ID); err != nil {
		t.Fatalf("AcceptExchange : %v", err)
	}
	if ex2, err = store.CancelExchange(ctx, ex2.ID); err != nil || ex2.Status != statusCancelled {
		t.Fatalf("CancelExchange = %+v, %v", ex2, err)
	}
	balance, _ := store.UserBalance(ctx, requester.ID)
	if balance != welcomeCredits {
		t.Errorf("solde après restitution = %d, attendu %d", balance, welcomeCredits)
	}
	if _, err := store.CancelExchange(ctx, ex2.ID); !errors.Is(err, ErrConflict) {
		t.Errorf("double annulation = %v, attendu ErrConflict", err)
	}
}

func TestStoreExchangeInsufficientCreditsOnAccept(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	owner := createTestUser(t, store, "offreur")
	requester := createTestUser(t, store, "demandeur")
	svc := createTestService(t, store, owner.ID, welcomeCredits+5)

	ex, err := store.CreateExchange(ctx, svc.ID, requester.ID, owner.ID)
	if err != nil {
		t.Fatalf("CreateExchange : %v", err)
	}
	if _, err := store.AcceptExchange(ctx, ex.ID); !errors.Is(err, ErrInsufficientCredits) {
		t.Errorf("acceptation sans provision = %v, attendu ErrInsufficientCredits", err)
	}
}

func TestStoreReviews(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	owner := createTestUser(t, store, "offreur")
	requester := createTestUser(t, store, "demandeur")
	svc := createTestService(t, store, owner.ID, 2)
	ex := completedExchange(t, store, svc.ID, requester.ID, owner.ID)

	review, err := store.CreateReview(ctx, Review{
		ExchangeID: ex.ID, ServiceID: svc.ID, ReviewerID: requester.ID,
		RevieweeID: owner.ID, Note: 5, Commentaire: "Impeccable",
	})
	if err != nil || review.ID == 0 || review.Note != 5 {
		t.Fatalf("CreateReview = %+v, %v", review, err)
	}

	// One review per exchange, enforced by the UNIQUE constraint.
	if _, err := store.CreateReview(ctx, Review{
		ExchangeID: ex.ID, ServiceID: svc.ID, ReviewerID: requester.ID,
		RevieweeID: owner.ID, Note: 1,
	}); !errors.Is(err, ErrConflict) {
		t.Errorf("second avis = %v, attendu ErrConflict", err)
	}

	byUser, err := store.ListUserReviews(ctx, owner.ID)
	if err != nil || len(byUser) != 1 || byUser[0].Commentaire != "Impeccable" {
		t.Errorf("ListUserReviews = %+v, %v", byUser, err)
	}
	byService, err := store.ListServiceReviews(ctx, svc.ID)
	if err != nil || len(byService) != 1 || byService[0].ID != review.ID {
		t.Errorf("ListServiceReviews = %+v, %v", byService, err)
	}
	empty, err := store.ListUserReviews(ctx, requester.ID)
	if err != nil || len(empty) != 0 {
		t.Errorf("ListUserReviews (aucun avis) = %+v, %v", empty, err)
	}
}

func TestStoreUserStats(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	owner := createTestUser(t, store, "offreur")
	requester := createTestUser(t, store, "demandeur")
	svc := createTestService(t, store, owner.ID, 3)
	ex := completedExchange(t, store, svc.ID, requester.ID, owner.ID)

	if _, err := store.CreateReview(ctx, Review{
		ExchangeID: ex.ID, ServiceID: svc.ID, ReviewerID: requester.ID,
		RevieweeID: owner.ID, Note: 4,
	}); err != nil {
		t.Fatalf("CreateReview : %v", err)
	}

	ownerStats, err := store.GetUserStats(ctx, owner.ID)
	if err != nil {
		t.Fatalf("GetUserStats : %v", err)
	}
	want := UserStats{
		UserID:             owner.ID,
		CreditBalance:      welcomeCredits + 3,
		CompletedExchanges: 1,
		ActiveServices:     1,
		AverageRating:      4,
		ReviewCount:        1,
		TotalEarned:        3,
		TotalSpent:         0,
	}
	if ownerStats != want {
		t.Errorf("stats offreur = %+v, attendu %+v", ownerStats, want)
	}

	reqStats, err := store.GetUserStats(ctx, requester.ID)
	if err != nil {
		t.Fatalf("GetUserStats : %v", err)
	}
	if reqStats.TotalSpent != 3 || reqStats.CreditBalance != welcomeCredits-3 || reqStats.AverageRating != 0 {
		t.Errorf("stats demandeur = %+v", reqStats)
	}
}
