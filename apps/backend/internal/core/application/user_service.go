package application

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"opensplit/apps/backend/internal/core/domain"
)

type UserService struct {
	repo      domain.UserRepository
	jwtSecret []byte
}

func NewUserService(repo domain.UserRepository, secret []byte) *UserService {
	return &UserService{repo: repo, jwtSecret: secret}
}

type CreateUserCommand struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// secret should ideally be loaded from an .env file
var JwtSecret = []byte("super-secret-splitwise-key-change-in-prod")

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

func (s *UserService) RegisterUser(ctx context.Context, id, displayName, plainPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{
		ID:           domain.UserID(id),
		DisplayName:  displayName,
		PasswordHash: string(hash),
	}
	return s.repo.Save(ctx, user)
}

func (s *UserService) LoginUser(ctx context.Context, id, plainPassword string) (string, error) {
	user, err := s.repo.GetByID(ctx, domain.UserID(id))
	if err != nil || !user.IsActive {
		return "", domain.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListAll(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, displayName string) error {
	if displayName == "" {
		return domain.ErrEmptyDisplayName
	}
	return s.repo.Update(ctx, domain.UserID(id), displayName)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, domain.UserID(id))
}
