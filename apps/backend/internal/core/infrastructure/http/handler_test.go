package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"opensplit/apps/backend/internal/core/application"
	"opensplit/apps/backend/internal/core/domain"
	"opensplit/libs/shared/money"

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
		nil,
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

type mockGroupRepo struct{}

func (m *mockGroupRepo) Save(ctx context.Context, g *domain.Group) error { return nil }
func (m *mockGroupRepo) ListForUser(ctx context.Context, u domain.UserID) ([]*domain.Group, error) {
	return []*domain.Group{
		{ID: "g1", Name: "Ski Trip 2026", Members: []domain.UserID{"Alice"}},
	}, nil
}
func (m *mockGroupRepo) GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	return &domain.Group{
		ID:      id,
		Name:    "Ski Trip 2026",
		Members: []domain.UserID{"Alice"},
	}, nil
}

func TestAPIHandler_GetBalances(t *testing.T) {
	expenseRepo := &mockExpenseRepo{}
	userRepo := &mockUserRepo{}
	groupRepo := &mockGroupRepo{}
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)
	userService := application.NewUserService(userRepo)
	groupService := application.NewGroupService(groupRepo)
	handler := NewAPIHandler(expenseService, userService, groupService)

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

func TestAPIHandler_Users(t *testing.T) {
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{})
	handler := NewAPIHandler(eService, uService, gService)

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

func TestAPIHandler_Expenses(t *testing.T) {
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{})
	handler := NewAPIHandler(eService, uService, gService)

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

func TestAPIHandler_Groups(t *testing.T) {
	// Initialize services with mocks
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{})
	handler := NewAPIHandler(eService, uService, gService)

	t.Run("POST /groups creates a group", func(t *testing.T) {
		body, _ := json.Marshal(application.CreateGroupCommand{Name: "Ski Trip 2026", Creator: "Alice"})
		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateGroup(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("POST /groups/{id}/members adds a member", func(t *testing.T) {
		body := []byte(`{"user_id": "Bob"}`)

		req := httptest.NewRequest("POST", "/groups/g1/members", bytes.NewBuffer(body))

		req.SetPathValue("id", "g1")

		rr := httptest.NewRecorder()
		handler.AddGroupMember(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("GET /groups returns list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/groups?user_id=Alice", nil)
		rr := httptest.NewRecorder()

		handler.ListGroups(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("Ski Trip 2026")) {
			t.Errorf("expected body to contain 'Ski Trip 2026', got %s", rr.Body.String())
		}
	})
}
