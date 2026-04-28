package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"opensplit/apps/backend/internal/core/application"
	"opensplit/apps/backend/internal/core/domain"
)

type APIHandler struct {
	expenseService *application.ExpenseService
	userService    *application.UserService
	groupService   *application.GroupService
}

func NewAPIHandler(es *application.ExpenseService, us *application.UserService, gs *application.GroupService) *APIHandler {
	return &APIHandler{
		expenseService: es,
		userService:    us,
		groupService:   gs,
	}
}

// POST /expenses
func (h *APIHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateExpenseCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.expenseService.AddExpense(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"status": "expense created"}`))
}

// GET /expenses?group_id={optional}
func (h *APIHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")

	var expenses []*domain.Expense
	var err error

	if groupID != "" {
		expenses, err = h.expenseService.ListExpensesByGroup(r.Context(), groupID)
	} else {
		expenses, err = h.expenseService.ListAllExpenses(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type responseData struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Total       int64  `json:"total_cents"`
		Payer       string `json:"payer"`
	}

	var res []responseData
	for _, exp := range expenses {
		res = append(res, responseData{
			ID:          string(exp.ID()),
			Description: exp.Description(),
			Total:       exp.Total().Int64(),
			Payer:       string(exp.Payer()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return
	}
}

// GET /balances?group_id={optional}
func (h *APIHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")

	var expenses []*domain.Expense
	var err error

	if groupID != "" {
		expenses, err = h.expenseService.ListExpensesByGroup(r.Context(), groupID)
	} else {
		expenses, err = h.expenseService.ListAllExpenses(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balances := domain.CalculateNetBalances(expenses)
	suggestions := domain.SimplifyDebts(balances)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"net_balances":          balances,
		"suggested_settlements": suggestions,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

// PUT /expenses/{id}
func (h *APIHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	expenseID := r.PathValue("id")

	var cmd application.UpdateExpenseCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	cmd.ID = expenseID

	// 4. Pass to the Application Service
	if err := h.expenseService.UpdateExpense(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "expense updated"}`))
}

// DELETE /expenses/{id}
func (h *APIHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	expenseID := r.PathValue("id")

	if err := h.expenseService.DeleteExpense(r.Context(), expenseID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "expense deleted"}`))
}

// POST /users
func (h *APIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.userService.CreateUser(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"status": "user created"}`))
}

// GET /users
func (h *APIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		return
	}
}

// POST /groups
func (h *APIHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateGroupCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	groupID, err := h.groupService.CreateGroup(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"status": "group created", "group_id": "%s"}`, groupID)))
}

// POST /groups/{id}/members
func (h *APIHandler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	var cmd struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if err := h.groupService.AddMemberToGroup(r.Context(), groupID, cmd.UserID); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyInGroup) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "member added"}`))
}

// GET /groups
func (h *APIHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error": "user_id is required"}`, http.StatusBadRequest)
		return
	}

	groups, err := h.groupService.ListGroupsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(groups)
}
