package main

import (
	"context"
	"strings"
)

// RegisterUser validates and creates a new account (welcome credits are granted
// by the store).
func (a *App) RegisterUser(ctx context.Context, input User) (User, error) {
	pseudo := strings.TrimSpace(input.Pseudo)
	if pseudo == "" {
		return User{}, &ValidationError{Field: "pseudo", Message: "le pseudo est obligatoire"}
	}
	return a.store.CreateUser(ctx, User{
		Pseudo: pseudo,
		Bio:    strings.TrimSpace(input.Bio),
		Ville:  strings.TrimSpace(input.Ville),
	})
}

// GetUser returns a public profile.
func (a *App) GetUser(ctx context.Context, id int) (User, error) {
	return a.store.GetUser(ctx, id)
}

// UpdateUser lets a user edit their own profile only.
func (a *App) UpdateUser(ctx context.Context, actorID, id int, input User) (User, error) {
	if actorID != id {
		return User{}, ErrForbidden
	}
	pseudo := strings.TrimSpace(input.Pseudo)
	if pseudo == "" {
		return User{}, &ValidationError{Field: "pseudo", Message: "le pseudo est obligatoire"}
	}
	return a.store.UpdateUser(ctx, User{
		ID:     id,
		Pseudo: pseudo,
		Bio:    strings.TrimSpace(input.Bio),
		Ville:  strings.TrimSpace(input.Ville),
	})
}

// GetUserSkills returns the skills of an existing user.
func (a *App) GetUserSkills(ctx context.Context, id int) ([]Skill, error) {
	exists, err := a.store.UserExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	return a.store.GetUserSkills(ctx, id)
}

// SetUserSkills validates then replaces the skills of the caller's own account.
func (a *App) SetUserSkills(ctx context.Context, actorID, id int, skills []Skill) ([]Skill, error) {
	if actorID != id {
		return nil, ErrForbidden
	}
	exists, err := a.store.UserExists(ctx, id)
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

	if err := a.store.SetUserSkills(ctx, id, cleaned); err != nil {
		return nil, err
	}
	return a.store.GetUserSkills(ctx, id)
}
