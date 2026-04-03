package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"opensplit/apps/backend/internal/expense/application"
	"opensplit/apps/backend/internal/expense/domain"
	"opensplit/libs/go-core/money"

	"github.com/google/uuid"
)

type mockExpenseRepo struct{}

func (m *mockExpenseRepo) Save(ctx context.Context, expense *domain.Expense) error { return nil }
func (m *mockExpenseRepo) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	total, _ := money.New(3000)
	split, _ := money.New(1500)
	exp, _ := domain.NewExpense(
		domain.ExpenseID(uuid.NewString()),
		"Test Dinner",
		total,
		"Alice",
		[]domain.Split{{User: "Alice", Amount: split}, {User: "Bob", Amount: split}},
	)
	return []*domain.Expense{exp}, nil
}

type mockUserRepo struct{}

func (m *mockUserRepo) Save(ctx context.Context, u domain.User) error { return nil }
func (m *mockUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	return []domain.User{{ID: "Alice", DisplayName: "Alice"}}, nil
}

func TestAPIHandler_GetBalances(t *testing.T) {
	expenseRepo := &mockExpenseRepo{}
	userRepo := &mockUserRepo{}
	expenseService := application.NewExpenseService(expenseRepo)
	userService := application.NewUserService(userRepo)
	handler := NewAPIHandler(expenseService, userService)

	req, err := http.NewRequest("GET", "/balances", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetBalances(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBodyFragment := `"suggested_settlements":[{"From":"Bob","To":"Alice","Amount":1500}]`
	if !bytes.Contains(rr.Body.Bytes(), []byte(expectedBodyFragment)) {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestAPIHandler_GetPostUsers(t *testing.T) {
	uService := application.NewUserService(&mockUserRepo{})
	eService := application.NewExpenseService(&mockExpenseRepo{})
	handler := NewAPIHandler(eService, uService)

	t.Run("POST /users creates a user", func(t *testing.T) {
		body, _ := json.Marshal(application.CreateUserCommand{ID: "Charlie", DisplayName: "Charlie"})
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateUser(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("GET /users returns list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("Alice")) {
			t.Errorf("expected body to contain Alice, got %s", rr.Body.String())
		}
	})
}

func TestAPIHandler_GetPostExpenses(t *testing.T) {
	uService := application.NewUserService(&mockUserRepo{})
	eService := application.NewExpenseService(&mockExpenseRepo{})
	handler := NewAPIHandler(eService, uService)

	t.Run("GET /expenses returns list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/expenses", nil)
		rr := httptest.NewRecorder()

		handler.ListExpenses(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		expectedBodyFragment := `"description":"Test Dinner","total_cents":3000,"payer":"Alice"`
		if !bytes.Contains(rr.Body.Bytes(), []byte(expectedBodyFragment)) {
			t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
		}
	})

	t.Run("POST /expenses handles invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer([]byte("{invalid}")))
		rr := httptest.NewRecorder()

		handler.CreateExpense(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})
}
