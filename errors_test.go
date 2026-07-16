package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"validation", &ValidationError{Field: "pseudo", Message: "obligatoire"}, http.StatusBadRequest},
		{"crédits insuffisants", ErrInsufficientCredits, http.StatusBadRequest},
		{"introuvable", ErrNotFound, http.StatusNotFound},
		{"non authentifié", ErrUnauthorized, http.StatusUnauthorized},
		{"interdit", ErrForbidden, http.StatusForbidden},
		{"conflit", ErrConflict, http.StatusConflict},
		{"erreur enveloppée", fmt.Errorf("contexte : %w", ErrNotFound), http.StatusNotFound},
		{"erreur inconnue", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpStatus(tt.err); got != tt.want {
				t.Errorf("httpStatus(%v) = %d, attendu %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestValidationErrorMessage(t *testing.T) {
	withField := &ValidationError{Field: "note", Message: "invalide"}
	if withField.Error() != "note : invalide" {
		t.Errorf("Error() = %q", withField.Error())
	}
	withoutField := &ValidationError{Message: "corps invalide"}
	if withoutField.Error() != "corps invalide" {
		t.Errorf("Error() = %q", withoutField.Error())
	}
}
