package main

import (
	"context"
	"strings"
)

// Service holds the business logic. It validates input and enforces the domain
// rules, delegating persistence to the Store. It knows nothing about HTTP.
type Service struct {
	store *Store
}

// NewService wires a Service to its Store.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// RegisterUser validates and creates a new account (welcome credits are granted
// by the store).
func (svc *Service) RegisterUser(ctx context.Context, input User) (User, error) {
	pseudo := strings.TrimSpace(input.Pseudo)
	if pseudo == "" {
		return User{}, &ValidationError{Field: "pseudo", Message: "le pseudo est obligatoire"}
	}
	return svc.store.CreateUser(ctx, User{
		Pseudo: pseudo,
		Bio:    strings.TrimSpace(input.Bio),
		Ville:  strings.TrimSpace(input.Ville),
	})
}

// GetUser returns a public profile.
func (svc *Service) GetUser(ctx context.Context, id int) (User, error) {
	return svc.store.GetUser(ctx, id)
}

// UpdateUser lets a user edit their own profile only.
func (svc *Service) UpdateUser(ctx context.Context, actorID, id int, input User) (User, error) {
	if actorID != id {
		return User{}, ErrForbidden
	}
	pseudo := strings.TrimSpace(input.Pseudo)
	if pseudo == "" {
		return User{}, &ValidationError{Field: "pseudo", Message: "le pseudo est obligatoire"}
	}
	return svc.store.UpdateUser(ctx, User{
		ID:     id,
		Pseudo: pseudo,
		Bio:    strings.TrimSpace(input.Bio),
		Ville:  strings.TrimSpace(input.Ville),
	})
}

// GetUserSkills returns the skills of an existing user.
func (svc *Service) GetUserSkills(ctx context.Context, id int) ([]Skill, error) {
	exists, err := svc.store.UserExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	return svc.store.GetUserSkills(ctx, id)
}

// SetUserSkills validates then replaces the skills of the caller's own account.
func (svc *Service) SetUserSkills(ctx context.Context, actorID, id int, skills []Skill) ([]Skill, error) {
	if actorID != id {
		return nil, ErrForbidden
	}
	exists, err := svc.store.UserExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	cleaned := make([]Skill, 0, len(skills))
	for _, sk := range skills {
		nom := strings.TrimSpace(sk.Nom)
		if nom == "" {
			return nil, &ValidationError{Field: "nom", Message: "le nom de la compétence est obligatoire"}
		}
		if !validNiveaux[sk.Niveau] {
			return nil, &ValidationError{Field: "niveau", Message: "niveau invalide (débutant, intermédiaire, expert)"}
		}
		cleaned = append(cleaned, Skill{Nom: nom, Niveau: sk.Niveau})
	}

	if err := svc.store.SetUserSkills(ctx, id, cleaned); err != nil {
		return nil, err
	}
	return svc.store.GetUserSkills(ctx, id)
}
