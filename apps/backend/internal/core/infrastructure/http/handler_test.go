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
)

// Expense
type mockExpenseRepo struct{}

func (m *mockExpenseRepo) Save(ctx context.Context, e *domain.Expense) error { return nil }
func (m *mockExpenseRepo) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	return nil, domain.ErrExpenseNotFound
}
func (m *mockExpenseRepo) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	total, _ := money.New(3000)
	split, _ := money.New(1500)
	exp, _ := domain.NewExpense(
		domain.ExpenseID("test-id"),
		nil,
		"Test Dinner",
		total,
		"Alice",
		[]domain.Split{{User: "Alice", Amount: split}, {User: "Bob", Amount: split}},
	)
	return []*domain.Expense{exp}, nil
}
func (m *mockExpenseRepo) ListByGroup(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error) {
	return []*domain.Expense{}, nil
}

func (m *mockExpenseRepo) Update(ctx context.Context, expense *domain.Expense) error { return nil }
func (m *mockExpenseRepo) Delete(ctx context.Context, id domain.ExpenseID) error     { return nil }

// User
type mockUserRepo struct{}

func (m *mockUserRepo) Save(ctx context.Context, u domain.User) error { return nil }
func (m *mockUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	return []domain.User{{ID: "Alice", DisplayName: "Alice"}}, nil
}
func (m *mockUserRepo) Update(ctx context.Context, u domain.UserID, d string) error { return nil }
func (m *mockUserRepo) SoftDelete(ctx context.Context, u domain.UserID) error       { return nil }

// Group
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
func (m *mockGroupRepo) UpdateName(ctx context.Context, id domain.GroupID, n string) error {
	return nil
}
func (m *mockGroupRepo) Delete(ctx context.Context, id domain.GroupID) error { return nil }
func (m *mockGroupRepo) RemoveMember(ctx context.Context, id domain.GroupID, u domain.UserID) error {
	return nil
}

func TestAPIHandler_GetBalances(t *testing.T) {
	expenseRepo := &mockExpenseRepo{}
	userRepo := &mockUserRepo{}
	groupRepo := &mockGroupRepo{}
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)
	userService := application.NewUserService(userRepo)
	groupService := application.NewGroupService(groupRepo, expenseRepo)
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
	gService := application.NewGroupService(&mockGroupRepo{}, &mockExpenseRepo{})
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

	t.Run("PUT /users/{id} updates display name", func(t *testing.T) {
		body := []byte(`{"display_name": "Alice Updated"}`)
		req := httptest.NewRequest("PUT", "/users/Alice", bytes.NewBuffer(body))
		req.SetPathValue("id", "Alice")
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /users/{id} soft deletes", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/Alice", nil)
		req.SetPathValue("id", "Alice")
		rr := httptest.NewRecorder()

		handler.DeleteUser(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_ExpensesGroupFiltering(t *testing.T) {
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{}, &mockExpenseRepo{})
	handler := NewAPIHandler(eService, uService, gService)

	t.Run("GET /balances filters by group_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/balances?group_id=g1", nil)
		rr := httptest.NewRecorder()

		handler.GetBalances(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_Expenses(t *testing.T) {
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{}, &mockExpenseRepo{})
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

	t.Run("PUT /expenses/{id} successfully updates", func(t *testing.T) {
		cmd := application.UpdateExpenseCommand{
			Description: "Updated Dinner",
			TotalCents:  4000,
			Payer:       "Alice",
			Splits:      map[string]int64{"Alice": 2000, "Bob": 2000},
		}
		body, _ := json.Marshal(cmd)

		req := httptest.NewRequest("PUT", "/expenses/exp-123", bytes.NewBuffer(body))
		// Inject the path value for the test router
		req.SetPathValue("id", "exp-123")

		rr := httptest.NewRecorder()
		handler.UpdateExpense(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /expenses/{id} successfully deletes", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/expenses/exp-123", nil)
		// Inject the path value for the test router
		req.SetPathValue("id", "exp-123")

		rr := httptest.NewRecorder()
		handler.DeleteExpense(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_Groups(t *testing.T) {
	eService := application.NewExpenseService(&mockExpenseRepo{}, &mockGroupRepo{})
	uService := application.NewUserService(&mockUserRepo{})
	gService := application.NewGroupService(&mockGroupRepo{}, &mockExpenseRepo{})
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

	t.Run("PUT /groups/{id} updates group name", func(t *testing.T) {
		body := []byte(`{"name": "New Trip Name"}`)
		req := httptest.NewRequest("PUT", "/groups/g1", bytes.NewBuffer(body))
		req.SetPathValue("id", "g1")
		rr := httptest.NewRecorder()

		handler.UpdateGroup(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /groups/{id} deletes group", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/g1", nil)
		req.SetPathValue("id", "g1")
		rr := httptest.NewRecorder()

		handler.DeleteGroup(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /groups/{id}/members/{user_id} triggers validation", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/g1/members/Bob", nil)
		req.SetPathValue("id", "g1")
		req.SetPathValue("user_id", "Bob")
		rr := httptest.NewRecorder()

		handler.RemoveGroupMember(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
