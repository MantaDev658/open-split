package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) error {
	query := `
		INSERT INTO users (id, display_name)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE 
		SET display_name = EXCLUDED.display_name, 
		    is_active = TRUE
	`
	_, err := r.db.ExecContext(ctx, query, string(user.ID), user.DisplayName)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	query := "SELECT id, display_name FROM users WHERE is_active = TRUE ORDER BY created_at ASC"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, id domain.UserID, newName string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE users SET display_name = $1 WHERE id = $2 AND is_active = TRUE", newName, string(id))
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found or inactive")
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id domain.UserID) error {
	res, err := r.db.ExecContext(ctx, "UPDATE users SET is_active = FALSE WHERE id = $1", string(id))
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
