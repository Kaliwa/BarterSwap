package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lib/pq"
)

// stubScanner feeds fixed values into scan destinations, mimicking *sql.Row,
// so the scan helpers are testable without a database.
type stubScanner struct {
	vals []any
	err  error
}

func (s stubScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i, d := range dest {
		switch p := d.(type) {
		case *int:
			*p = s.vals[i].(int)
		case *string:
			*p = s.vals[i].(string)
		case *bool:
			*p = s.vals[i].(bool)
		case *float64:
			*p = s.vals[i].(float64)
		case *time.Time:
			*p = s.vals[i].(time.Time)
		default:
			return fmt.Errorf("type de destination inattendu %T", d)
		}
	}
	return nil
}

var scanTime = time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)

func TestScanService(t *testing.T) {
	sc := stubScanner{vals: []any{10, 1, "Taille de haies", "desc", "Jardinage", 90, 2, "Lyon", true, scanTime}}
	svc, err := scanService(sc)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if svc.ID != 10 || svc.Titre != "Taille de haies" || !svc.Actif || svc.CreatedAt != "2026-07-16T10:00:00Z" {
		t.Errorf("service scanné : %+v", svc)
	}

	if _, err := scanService(stubScanner{err: errors.New("boom")}); err == nil {
		t.Error("l'erreur de scan doit être propagée")
	}
}

func TestScanExchange(t *testing.T) {
	sc := stubScanner{vals: []any{1, 10, 2, 1, statusPending, scanTime, scanTime}}
	ex, err := scanExchange(sc)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if ex.ID != 1 || ex.ServiceID != 10 || ex.Status != statusPending || ex.UpdatedAt != "2026-07-16T10:00:00Z" {
		t.Errorf("échange scanné : %+v", ex)
	}

	if _, err := scanExchange(stubScanner{err: errors.New("boom")}); err == nil {
		t.Error("l'erreur de scan doit être propagée")
	}
}

func TestScanReview(t *testing.T) {
	sc := stubScanner{vals: []any{1, 5, 10, 2, 1, 4, "Très bien", scanTime}}
	rv, err := scanReview(sc)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if rv.ID != 1 || rv.ExchangeID != 5 || rv.Note != 4 || rv.CreatedAt != "2026-07-16T10:00:00Z" {
		t.Errorf("avis scanné : %+v", rv)
	}

	if _, err := scanReview(stubScanner{err: errors.New("boom")}); err == nil {
		t.Error("l'erreur de scan doit être propagée")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"violation d'unicité", &pq.Error{Code: "23505"}, true},
		{"violation enveloppée", fmt.Errorf("insertion : %w", &pq.Error{Code: "23505"}), true},
		{"autre erreur postgres", &pq.Error{Code: "23503"}, false},
		{"erreur quelconque", errors.New("boom"), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUniqueViolation(tt.err); got != tt.want {
				t.Errorf("isUniqueViolation(%v) = %v, attendu %v", tt.err, got, tt.want)
			}
		})
	}
}
