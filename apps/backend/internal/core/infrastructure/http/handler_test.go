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

func newTestServices(eRepo *mocks.MockExpenseRepo, uRepo *mocks.MockUserRepo, gRepo *mocks.MockGroupRepo, aRepo *mocks.MockAuditRepo) (*application.ExpenseService, *application.UserService, *application.GroupService) {
	tx := &mocks.MockTransactor{}
	es := application.NewExpenseService(eRepo, gRepo, aRepo, tx)
	us := application.NewUserService(uRepo, []byte("test-secret"))
	gs := application.NewGroupService(gRepo, eRepo, aRepo, tx)
	return es, us, gs
}

func TestAPIHandler_GetBalances(t *testing.T) {
	aRepo := &mocks.MockAuditRepo{}
	eRepo := &mocks.MockExpenseRepo{ListAllFunc: func(ctx context.Context, page domain.Page) ([]*domain.Expense, error) {
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
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(eRepo, uRepo, &mocks.MockGroupRepo{}, aRepo)
	handler := NewAPIHandler(es, us, gs)

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
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
		ListAllFunc: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{{ID: "Alice", DisplayName: "Alice"}}, nil
		},
	}
	es, us, gs := newTestServices(&mocks.MockExpenseRepo{}, uRepo, &mocks.MockGroupRepo{}, &mocks.MockAuditRepo{})
	handler := NewAPIHandler(es, us, gs)

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

func TestAPIHandler_Auth(t *testing.T) {
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(&mocks.MockExpenseRepo{}, uRepo, &mocks.MockGroupRepo{}, &mocks.MockAuditRepo{})
	handler := NewAPIHandler(es, us, gs)

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
	es, us, gs := newTestServices(&mocks.MockExpenseRepo{}, &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			return &domain.User{ID: id, IsActive: true}, nil
		},
	}, &mocks.MockGroupRepo{}, &mocks.MockAuditRepo{})
	handler := NewAPIHandler(es, us, gs)

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
	aRepo := &mocks.MockAuditRepo{}
	eRepo := &mocks.MockExpenseRepo{
		ListAllFunc: func(ctx context.Context, page domain.Page) ([]*domain.Expense, error) {
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
		},
		GetByIDFunc: func(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
			total, _ := money.New(3000)
			splitAmt, _ := money.New(3000)
			exp, err := domain.NewExpense(id, nil, "Test", total, "Alice", []domain.Split{
				{User: "Alice", Amount: splitAmt},
			})
			if err != nil {
				panic("invalid mock setup: " + err.Error())
			}
			return exp, nil
		},
		DeleteFunc: func(ctx context.Context, id domain.ExpenseID) error { return nil },
	}

	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(eRepo, uRepo, &mocks.MockGroupRepo{}, aRepo)
	handler := NewAPIHandler(es, us, gs)

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

	t.Run("POST /expenses rejects zero total", func(t *testing.T) {
		body, _ := json.Marshal(application.CreateExpenseCommand{
			TotalCents: 0, Payer: "Alice", SplitType: "EQUAL",
			Splits: []application.SplitDetail{{UserID: "Alice"}},
		})
		req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.CreateExpense(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for zero total, got %d", rr.Code)
		}
	})

	t.Run("PUT /expenses/{id} successfully updates", func(t *testing.T) {
		cmd := application.UpdateExpenseCommand{
			Description: "Updated Dinner",
			TotalCents:  4000,
			Payer:       "Alice",
			SplitType:   "EQUAL",
			Splits: []application.SplitDetail{
				{UserID: "Alice", Value: 2000},
				{UserID: "Bob", Value: 2000},
			},
		}
		body, _ := json.Marshal(cmd)

		req := httptest.NewRequest("PUT", "/expenses/exp-123", bytes.NewBuffer(body))
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
	aRepo := &mocks.MockAuditRepo{}
	eRepo := &mocks.MockExpenseRepo{
		SaveFunc: func(ctx context.Context, expense *domain.Expense) error { return nil },
	}
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(eRepo, uRepo, &mocks.MockGroupRepo{}, aRepo)
	handler := NewAPIHandler(es, us, gs)

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
	aRepo := &mocks.MockAuditRepo{}
	eRepo := &mocks.MockExpenseRepo{
		GetFriendBalanceSummaryFunc: func(ctx context.Context, userID domain.UserID) ([]domain.FriendBalance, error) {
			return []domain.FriendBalance{
				{FriendID: "Charlie", NetCents: 1000}, // Charlie owes Alice $10
			}, nil
		},
	}
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(eRepo, uRepo, &mocks.MockGroupRepo{}, aRepo)
	handler := NewAPIHandler(es, us, gs)

	t.Run("GET /friends/{user_id}/balances succeeds for auth user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/friends/Alice/balances", nil)
		req.SetPathValue("user_id", "Alice")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)
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

	t.Run("GET /friends/{user_id}/balances rejects mismatched user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/friends/Bob/balances", nil)
		req.SetPathValue("user_id", "Bob")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice") // auth user is Alice, path is Bob
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.GetFriendBalances(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for mismatched user, got %d", rr.Code)
		}
	})
}

func TestAPIHandler_Groups(t *testing.T) {
	aRepo := &mocks.MockAuditRepo{
		SaveFunc: func(ctx context.Context, log domain.AuditLog) error { return nil },
	}

	eRepo := &mocks.MockExpenseRepo{
		ListByGroupFunc: func(ctx context.Context, groupID domain.GroupID, page domain.Page) ([]*domain.Expense, error) {
			return []*domain.Expense{}, nil
		},
	}

	gRepo := &mocks.MockGroupRepo{
		ListForUserFunc: func(ctx context.Context, userID domain.UserID) ([]*domain.Group, error) {
			return []*domain.Group{
				{ID: "g1", Name: "Ski Trip 2026", Members: []domain.UserID{"Alice"}},
			}, nil
		},
		GetByIDFunc: func(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
			return &domain.Group{ID: "g1", Name: "Ski Trip 2026", Members: []domain.UserID{"Alice"}}, nil
		},
		UpdateNameFunc:   func(ctx context.Context, id domain.GroupID, newName string) error { return nil },
		DeleteFunc:       func(ctx context.Context, id domain.GroupID) error { return nil },
		RemoveMemberFunc: func(ctx context.Context, groupID domain.GroupID, userID domain.UserID) error { return nil },
		SaveFunc:         func(ctx context.Context, group *domain.Group) error { return nil },
	}
	uRepo := &mocks.MockUserRepo{
		SaveFunc: func(ctx context.Context, user domain.User) error { return nil },
		GetByIDFunc: func(ctx context.Context, id domain.UserID) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &domain.User{ID: "Alice", IsActive: true, PasswordHash: string(hash)}, nil
		},
	}
	es, us, gs := newTestServices(eRepo, uRepo, gRepo, aRepo)
	handler := NewAPIHandler(es, us, gs)

	t.Run("POST /groups creates a group using JWT identity as creator", func(t *testing.T) {
		// Creator comes from JWT, not from the request body
		body, _ := json.Marshal(map[string]string{"name": "Ski Trip 2026"})
		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.CreateGroup(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("POST /groups/{id}/members adds a member", func(t *testing.T) {
		body := []byte(`{"user_id": "Bob"}`)
		req := httptest.NewRequest("POST", "/groups/g1/members", bytes.NewBuffer(body))
		req.SetPathValue("id", "g1")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.AddGroupMember(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("GET /groups returns list for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/groups", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ListGroups(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("Ski Trip 2026")) {
			t.Errorf("expected body to contain 'Ski Trip 2026', got %s", rr.Body.String())
		}
	})

	t.Run("GET /groups rejects unauthenticated request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/groups", nil)
		rr := httptest.NewRecorder()

		handler.ListGroups(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("PUT /groups/{id} updates group name", func(t *testing.T) {
		body := []byte(`{"name": "New Trip Name"}`)
		req := httptest.NewRequest("PUT", "/groups/g1", bytes.NewBuffer(body))
		req.SetPathValue("id", "g1")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.UpdateGroup(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /groups/{id} deletes group", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/g1", nil)
		req.SetPathValue("id", "g1")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.DeleteGroup(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("DELETE /groups/{id}/members/{user_id} removes member", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/g1/members/Bob", nil)
		req.SetPathValue("id", "g1")
		req.SetPathValue("user_id", "Bob")
		ctx := context.WithValue(req.Context(), UserIDKey, "Alice")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.RemoveGroupMember(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
