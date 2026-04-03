package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"opensplit/apps/backend/internal/expense/domain"
	"opensplit/apps/backend/internal/expense/infrastructure/postgres"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set. Skipping User integration test.")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// clean tables before running to ensure an isolated environment
	_, _ = db.Exec("DELETE FROM splits")
	_, _ = db.Exec("DELETE FROM expenses")
	_, _ = db.Exec("DELETE FROM users")

	// seed required users for Foreign Key constraints
	_, err = db.Exec("INSERT INTO users (id, display_name) VALUES ('Alice', 'Alice'), ('Bob', 'Bob')")
	if err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	return db
}

func TestUserRepository_Lifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
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
