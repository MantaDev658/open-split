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
	"opensplit/apps/backend/internal/core/mocks"
	"opensplit/libs/shared/money"

	"golang.org/x/crypto/bcrypt"
)

func TestAPIHandler_GetBalances(t *testing.T) {
	expenseRepo := &mocks.MockExpenseRepo{ListAllFunc: func(ctx context.Context) ([]*domain.Expense, error) {
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
	}}
	userRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	groupRepo := &mocks.MockGroupRepo{}
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)
	userService := application.NewUserService(userRepo, []byte("test-secret"))
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
	eService := application.NewExpenseService(&mocks.MockExpenseRepo{}, &mocks.MockGroupRepo{})
	uService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
		ListAllFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{{ID: "Alice", DisplayName: "Alice"}}, nil
		},
	}, []byte("test-secret"))
	gService := application.NewGroupService(&mocks.MockGroupRepo{}, &mocks.MockExpenseRepo{})
	handler := NewAPIHandler(eService, uService, gService)

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

// Add this block to test the Auth Endpoints
func TestAPIHandler_Auth(t *testing.T) {
	uService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))

	handler := NewAPIHandler(
		application.NewExpenseService(&mocks.MockExpenseRepo{}, &mocks.MockGroupRepo{}),
		uService,
		application.NewGroupService(&mocks.MockGroupRepo{}, &mocks.MockExpenseRepo{}),
	)

	t.Run("POST /auth/register creates user", func(t *testing.T) {
		body := []byte(`{"id": "Alice", "display_name": "Alice", "password": "password123"}`)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("POST /auth/login returns token", func(t *testing.T) {
		body := []byte(`{"id": "Alice", "password": "password123"}`)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.LoginUser(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("token")) {
			t.Errorf("expected JSON with token, got %s", rr.Body.String())
		}
	})
}

func TestAPIHandler_ExpensesGroupFiltering(t *testing.T) {
	eService := application.NewExpenseService(&mocks.MockExpenseRepo{}, &mocks.MockGroupRepo{})
	uService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))
	gService := application.NewGroupService(&mocks.MockGroupRepo{}, &mocks.MockExpenseRepo{})
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
	expenseRepo := &mocks.MockExpenseRepo{ListAllFunc: func(ctx context.Context) ([]*domain.Expense, error) {
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
	}}
	groupRepo := &mocks.MockGroupRepo{}
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)
	userService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))
	groupService := application.NewGroupService(groupRepo, expenseRepo)
	handler := NewAPIHandler(expenseService, userService, groupService)

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
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

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
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.DeleteExpense(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_CreateSettlement(t *testing.T) {
	expenseRepo := &mocks.MockExpenseRepo{
		SaveFunc: func(ctx context.Context, expense *domain.Expense) error {
			return nil
		},
	}
	expenseService := application.NewExpenseService(expenseRepo, &mocks.MockGroupRepo{})
	userService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))
	groupService := application.NewGroupService(&mocks.MockGroupRepo{}, expenseRepo)

	handler := NewAPIHandler(expenseService, userService, groupService)

	t.Run("POST /settlements succeeds", func(t *testing.T) {
		cmd := application.SettleUpCommand{
			PayerID:     "Alice",
			ReceiverID:  "Bob",
			AmountCents: 2000,
		}
		body, _ := json.Marshal(cmd)
		req := httptest.NewRequest("POST", "/settlements", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.CreateSettlement(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201 Created, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("POST /settlements handles bad JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/settlements", bytes.NewBufferString("{bad-json}"))
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.CreateSettlement(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_GetFriendBalances(t *testing.T) {
	eRepo := &mocks.MockExpenseRepo{
		ListNonGroupExpensesByUserFunc: func(ctx context.Context, userID domain.UserID) ([]*domain.Expense, error) {
			total, _ := money.New(2000)
			split, _ := money.New(1000)
			exp, _ := domain.NewExpense("exp-1", nil, "Drinks", total, "Alice", []domain.Split{
				{User: "Alice", Amount: split}, {User: "Charlie", Amount: split},
			})
			return []*domain.Expense{exp}, nil
		},
	}
	userService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))

	handler := NewAPIHandler(
		application.NewExpenseService(eRepo, &mocks.MockGroupRepo{}),
		userService,
		application.NewGroupService(&mocks.MockGroupRepo{}, eRepo),
	)

	t.Run("GET /friends/{user_id}/balances succeeds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/friends/Alice/balances", nil)
		req.SetPathValue("user_id", "Alice")
		rr := httptest.NewRecorder()

		handler.GetFriendBalances(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}

		expectedBody := `"From":"Charlie","To":"Alice","Amount":1000`
		if !bytes.Contains(rr.Body.Bytes(), []byte(expectedBody)) {
			t.Errorf("expected body to contain settlement, got: %s", rr.Body.String())
		}
	})
}

func TestAPIHandler_Groups(t *testing.T) {
	expenseRepo := &mocks.MockExpenseRepo{}
	groupRepo := &mocks.MockGroupRepo{ListForUserFunc: func(ctx context.Context, userID domain.UserID) ([]*domain.Group, error) {
		return []*domain.Group{
			{ID: "g1", Name: "Ski Trip 2026", Members: []domain.UserID{"Alice"}},
		}, nil
	}}
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)
	userService := application.NewUserService(&mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}, []byte("test-secret"))
	groupService := application.NewGroupService(groupRepo, expenseRepo)
	handler := NewAPIHandler(expenseService, userService, groupService)

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
