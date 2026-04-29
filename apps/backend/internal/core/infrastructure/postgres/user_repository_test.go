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

	user := domain.User{ID: "Charlie", DisplayName: "Charlie Brown"}
	err := repo.Save(ctx, user)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	err = repo.Update(ctx, "Charlie", "Charles Brown")
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	err = repo.SoftDelete(ctx, "Charlie")
	if err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	users, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	for _, u := range users {
		if u.ID == "Charlie" {
			t.Errorf("expected Charlie to be hidden by soft delete, but he was returned")
		}
	}
}
