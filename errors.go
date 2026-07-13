package main

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors shared across the business layer. Handlers translate them
// into HTTP status codes via httpStatus.
var (
	ErrNotFound            = errors.New("ressource introuvable")
	ErrUnauthorized        = errors.New("authentification requise")
	ErrForbidden           = errors.New("action non autorisée")
	ErrConflict            = errors.New("conflit avec l'état actuel")
	ErrInsufficientCredits = errors.New("crédits insuffisants")
)

// ValidationError signals invalid client input and maps to HTTP 400.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s : %s", e.Field, e.Message)
}

// httpStatus maps a business error to the HTTP status it should produce.
func httpStatus(err error) int {
	var ve *ValidationError
	switch {
	case errors.As(err, &ve):
		return http.StatusBadRequest
	case errors.Is(err, ErrInsufficientCredits):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
