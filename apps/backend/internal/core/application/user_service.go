package application

import (
	"context"
	"errors"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type CreateUserCommand struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func (s *UserService) CreateUser(ctx context.Context, cmd CreateUserCommand) error {
	if cmd.ID == "" || cmd.DisplayName == "" {
		return fmt.Errorf("user ID and display name are required")
	}

	user := domain.User{
		ID:          domain.UserID(cmd.ID),
		DisplayName: cmd.DisplayName,
	}

	return s.repo.Save(ctx, user)
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListAll(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, displayName string) error {
	if displayName == "" {
		return errors.New("display name cannot be empty")
	}
	return s.repo.Update(ctx, domain.UserID(id), displayName)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, domain.UserID(id))
}
