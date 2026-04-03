package postgres

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set. Skipping integration test.")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Clean tables in reverse order of foreign keys
	_, _ = db.Exec("DELETE FROM splits")
	_, _ = db.Exec("DELETE FROM expenses")
	_, _ = db.Exec("DELETE FROM users")

	// Seed basic users needed for most tests
	_, err = db.Exec("INSERT INTO users (id, display_name) VALUES ('Alice', 'Alice'), ('Bob', 'Bob')")
	if err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	return db
}
