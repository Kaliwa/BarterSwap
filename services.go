package main

import (
	"context"
	"strings"
)

// CreateService validates a new announcement and links it to a skill the
// provider actually declared (matched to the service category).
func (a *App) CreateService(ctx context.Context, providerID int, input Service) (Service, error) {
	if err := validateServiceInput(input); err != nil {
		return Service{}, err
	}
	if err := a.ensureProviderHasSkill(ctx, providerID, input.Categorie); err != nil {
		return Service{}, err
	}
	return a.store.CreateService(ctx, Service{
		ProviderID:   providerID,
		Titre:        strings.TrimSpace(input.Titre),
		Description:  strings.TrimSpace(input.Description),
		Categorie:    input.Categorie,
		DureeMinutes: input.DureeMinutes,
		Credits:      input.Credits,
		Ville:        strings.TrimSpace(input.Ville),
		Actif:        true,
	})
}

// GetService returns a single announcement.
func (a *App) GetService(ctx context.Context, id int) (Service, error) {
	return a.store.GetService(ctx, id)
}

// ListServices returns the announcements matching the optional filters.
func (a *App) ListServices(ctx context.Context, f ServiceFilter) ([]Service, error) {
	return a.store.ListServices(ctx, f)
}

// UpdateService lets a provider edit their own announcement only.
func (a *App) UpdateService(ctx context.Context, actorID, id int, input Service) (Service, error) {
	existing, err := a.store.GetService(ctx, id)
	if err != nil {
		return Service{}, err
	}
	if existing.ProviderID != actorID {
		return Service{}, ErrForbidden
	}
	if err := validateServiceInput(input); err != nil {
		return Service{}, err
	}
	if err := a.ensureProviderHasSkill(ctx, actorID, input.Categorie); err != nil {
		return Service{}, err
	}
	return a.store.UpdateService(ctx, Service{
		ID:           id,
		Titre:        strings.TrimSpace(input.Titre),
		Description:  strings.TrimSpace(input.Description),
		Categorie:    input.Categorie,
		DureeMinutes: input.DureeMinutes,
		Credits:      input.Credits,
		Ville:        strings.TrimSpace(input.Ville),
		Actif:        input.Actif,
	})
}

// DeleteService lets a provider remove their own announcement only.
func (a *App) DeleteService(ctx context.Context, actorID, id int) error {
	existing, err := a.store.GetService(ctx, id)
	if err != nil {
		return err
	}
	if existing.ProviderID != actorID {
		return ErrForbidden
	}
	return a.store.DeleteService(ctx, id)
}

func validateServiceInput(s Service) error {
	if strings.TrimSpace(s.Titre) == "" {
		return &ValidationError{Field: "titre", Message: "le titre est obligatoire"}
	}
	if !validCategories[s.Categorie] {
		return &ValidationError{Field: "categorie", Message: "catégorie invalide"}
	}
	if s.DureeMinutes <= 0 {
		return &ValidationError{Field: "duree_minutes", Message: "la durée doit être positive"}
	}
	if s.Credits <= 0 {
		return &ValidationError{Field: "credits", Message: "le coût en crédits doit être positif"}
	}
	return nil
}

// ensureProviderHasSkill enforces that a service is backed by a declared skill:
// the provider must own a skill whose name matches the service category.
func (a *App) ensureProviderHasSkill(ctx context.Context, providerID int, categorie string) error {
	skills, err := a.store.GetUserSkills(ctx, providerID)
	if err != nil {
		return err
	}
	for _, sk := range skills {
		if strings.EqualFold(strings.TrimSpace(sk.Nom), strings.TrimSpace(categorie)) {
			return nil
		}
	}
	return &ValidationError{Field: "categorie", Message: "vous devez posséder une compétence correspondant à cette catégorie"}
}
