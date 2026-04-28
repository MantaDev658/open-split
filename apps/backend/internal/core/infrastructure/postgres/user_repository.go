package postgres

import (
	"context"
	"database/sql"
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
		ON CONFLICT (id) DO UPDATE SET display_name = EXCLUDED.display_name
	`
	_, err := r.db.ExecContext(ctx, query, string(user.ID), user.DisplayName)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, display_name FROM users ORDER BY created_at ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var id string
		if err := rows.Scan(&id, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		u.ID = domain.UserID(id)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return users, nil
}
