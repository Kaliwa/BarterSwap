package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateUser inserts a new user and grants the welcome credits in the same
// transaction, so an account never exists without its opening journal entry.
func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var created User
	var createdAt time.Time
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (pseudo, bio, ville)
		 VALUES ($1, $2, $3)
		 RETURNING id, pseudo, bio, ville, created_at`,
		u.Pseudo, u.Bio, u.Ville,
	).Scan(&created.ID, &created.Pseudo, &created.Bio, &created.Ville, &createdAt)
	if err != nil {
		return User{}, fmt.Errorf("insertion utilisateur : %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, montant, type)
		 VALUES ($1, $2, $3)`,
		created.ID, welcomeCredits, txEarn,
	); err != nil {
		return User{}, fmt.Errorf("crédits de bienvenue : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit création utilisateur : %w", err)
	}

	created.CreatedAt = createdAt.Format(time.RFC3339)
	created.CreditBalance = welcomeCredits
	created.Skills = []Skill{}
	return created, nil
}

// GetUser returns a full user profile, including derived credit balance and
// skills. It returns ErrNotFound when the user does not exist.
func (s *Store) GetUser(ctx context.Context, id int) (User, error) {
	var u User
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.pseudo, u.bio, u.ville, u.created_at,
		        COALESCE((SELECT SUM(montant) FROM credit_transactions WHERE user_id = u.id), 0)
		 FROM users u
		 WHERE u.id = $1`, id,
	).Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &createdAt, &u.CreditBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.CreatedAt = createdAt.Format(time.RFC3339)

	skills, err := s.GetUserSkills(ctx, id)
	if err != nil {
		return User{}, err
	}
	u.Skills = skills
	return u, nil
}

// UpdateUser updates the editable profile fields and returns the fresh profile.
func (s *Store) UpdateUser(ctx context.Context, u User) (User, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET pseudo = $1, bio = $2, ville = $3 WHERE id = $4`,
		u.Pseudo, u.Bio, u.Ville, u.ID)
	if err != nil {
		return User{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return User{}, err
	}
	if n == 0 {
		return User{}, ErrNotFound
	}
	return s.GetUser(ctx, u.ID)
}

// UserExists reports whether a user id is present.
func (s *Store) UserExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

// GetUserSkills returns the skills of a user, in insertion order.
func (s *Store) GetUserSkills(ctx context.Context, userID int) ([]Skill, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT nom, niveau FROM skills WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []Skill{}
	for rows.Next() {
		var sk Skill
		if err := rows.Scan(&sk.Nom, &sk.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, sk)
	}
	return skills, rows.Err()
}

// SetUserSkills replaces all skills of a user atomically.
func (s *Store) SetUserSkills(ctx context.Context, userID int, skills []Skill) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("suppression compétences : %w", err)
	}
	for _, sk := range skills {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`,
			userID, sk.Nom, sk.Niveau,
		); err != nil {
			return fmt.Errorf("insertion compétence : %w", err)
		}
	}
	return tx.Commit()
}
