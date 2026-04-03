package postgres

import (
	"context"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
)

func TestUserRepository_Lifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := domain.User{
		ID:          "Charlie",
		DisplayName: "Charlie Kelly",
	}

	err := repo.Save(ctx, testUser)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	users, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	found := false
	for _, u := range users {
		if u.ID == "Charlie" && u.DisplayName == "Charlie Kelly" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("saved user 'Charlie' was not found in the database")
	}
}
