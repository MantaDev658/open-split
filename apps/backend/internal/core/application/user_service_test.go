package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
)

type mockUserRepo struct {
	saveFunc       func(ctx context.Context, u domain.User) error
	listAllFunc    func(ctx context.Context) ([]domain.User, error)
	updateFunc     func(ctx context.Context, u domain.UserID, d string) error
	softDeleteFunc func(ctx context.Context, u domain.UserID) error
}

func (m *mockUserRepo) Save(ctx context.Context, u domain.User) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, u)
	}
	return nil
}

func (m *mockUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u domain.UserID, d string) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, u, d)
	}
	return nil
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, u domain.UserID) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, u)
	}
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CreateUserCommand
		mockSave    func(ctx context.Context, u domain.User) error
		expectError bool
	}{
		{
			name:        "Success path",
			cmd:         CreateUserCommand{ID: "Alice", DisplayName: "Alice Smith"},
			mockSave:    func(ctx context.Context, u domain.User) error { return nil },
			expectError: false,
		},
		{
			name:        "Fails on empty ID",
			cmd:         CreateUserCommand{ID: "", DisplayName: "No Name"},
			mockSave:    func(ctx context.Context, u domain.User) error { return nil },
			expectError: true,
		},
		{
			name:        "Fails on infrastructure error",
			cmd:         CreateUserCommand{ID: "Bob", DisplayName: "Bob Builder"},
			mockSave:    func(ctx context.Context, u domain.User) error { return errors.New("db down") },
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{saveFunc: tt.mockSave}
			service := NewUserService(repo)

			err := service.CreateUser(context.Background(), tt.cmd)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateUser() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	repo := &mockUserRepo{
		listAllFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{{ID: "Alice", DisplayName: "Alice"}}, nil
		},
	}
	service := NewUserService(repo)

	users, err := service.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 || users[0].ID != "Alice" {
		t.Errorf("unexpected user list returned")
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	repo := &mockUserRepo{}
	service := NewUserService(repo)

	t.Run("Fails with empty name", func(t *testing.T) {
		err := service.UpdateUser(context.Background(), "Alice", "")
		if err == nil {
			t.Error("expected error for empty display name")
		}
	})

	t.Run("Succeeds with valid name", func(t *testing.T) {
		err := service.UpdateUser(context.Background(), "Alice", "Alice S.")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	repo := &mockUserRepo{}
	service := NewUserService(repo)

	err := service.DeleteUser(context.Background(), "Alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
