package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/libs/go-core/money"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set. Skipping Postgres integration test.")
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

func TestExpenseRepository_Lifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expenseID := domain.ExpenseID(uuid.NewString())
	total, _ := money.New(5000)
	split1, _ := money.New(2500)
	split2, _ := money.New(2500)

	exp, err := domain.NewExpense(
		expenseID,
		"Integration Test Dinner",
		total,
		"Alice",
		[]domain.Split{
			{User: "Alice", Amount: split1},
			{User: "Bob", Amount: split2},
		},
	)
	if err != nil {
		t.Fatalf("failed to create domain expense: %v", err)
	}

	err = repo.Save(ctx, exp)
	if err != nil {
		t.Fatalf("failed to save expense: %v", err)
	}

	fetchedExp, err := repo.GetByID(ctx, expenseID)
	if err != nil {
		t.Fatalf("failed to get expense by id: %v", err)
	}

	if fetchedExp.ID() != exp.ID() {
		t.Errorf("expected ID %s, got %s", exp.ID(), fetchedExp.ID())
	}
	if fetchedExp.TotalAmount().Int64() != 5000 {
		t.Errorf("expected total 5000, got %d", fetchedExp.TotalAmount().Int64())
	}
	if len(fetchedExp.Splits()) != 2 {
		t.Errorf("expected 2 splits, got %d", len(fetchedExp.Splits()))
	}

	allExpenses, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list all expenses: %v", err)
	}

	if len(allExpenses) != 1 {
		t.Errorf("expected 1 expense in db, got %d", len(allExpenses))
	}
	if allExpenses[0].Description() != "Integration Test Dinner" {
		t.Errorf("expected description 'Integration Test Dinner', got %s", allExpenses[0].Description())
	}
}

func TestExpenseRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewExpenseRepository(db)

	randomID := domain.ExpenseID(uuid.NewString())
	_, err := repo.GetByID(context.Background(), randomID)
	if err != domain.ErrExpenseNotFound {
		t.Errorf("expected ErrExpenseNotFound, got %v", err)
	}
}
