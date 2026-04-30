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

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return body, err
	}
	return body, nil
}

func getAuthUserID(r *http.Request) (string, error) {
	id, ok := r.Context().Value(UserIDKey).(string)
	if !ok || id == "" {
		return "", domain.ErrUnauthorized
	}
	return id, nil
}

// POST /expenses
func (h *APIHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	cmd, err := decodeJSON[application.CreateExpenseCommand](w, r)
	if err != nil {
		return
	}

	if cmd.Payer, err = getAuthUserID(r); err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := h.expenseService.AddExpense(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
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

// GET /friends/{user_id}/balances
func (h *APIHandler) GetFriendBalances(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	debts, err := h.expenseService.GetFriendBalances(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(debts); err != nil {
		return
	}
}

// PUT /expenses/{id}
func (h *APIHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	expenseID := r.PathValue("id")

	cmd, err := decodeJSON[application.UpdateExpenseCommand](w, r)
	if err != nil {
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

// POST /settlements
func (h *APIHandler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	cmd, err := decodeJSON[application.SettleUpCommand](w, r)
	if err != nil {
		return
	}

	if cmd.PayerID, err = getAuthUserID(r); err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := h.expenseService.SettleUp(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// POST /auth/register
func (h *APIHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	cmd, err := decodeJSON[struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
	}](w, r)
	if err != nil {
		return
	}

	if err := h.userService.RegisterUser(r.Context(), cmd.ID, cmd.DisplayName, cmd.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// POST /auth/login
func (h *APIHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	cmd, err := decodeJSON[struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		return
	}

	token, err := h.userService.LoginUser(r.Context(), cmd.ID, cmd.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		return
	}
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

// PUT /users/{id}
func (h *APIHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	cmd, err := decodeJSON[struct {
		DisplayName string `json:"display_name"`
	}](w, r)
	if err != nil {
		return
	}

	if err := h.userService.UpdateUser(r.Context(), userID, cmd.DisplayName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /users/{id}
func (h *APIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /groups
func (h *APIHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	cmd, err := decodeJSON[application.CreateGroupCommand](w, r)
	if err != nil {
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

	cmd, err := decodeJSON[struct {
		UserID string `json:"user_id"`
	}](w, r)
	if err != nil {
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

// PUT /groups/{id}
func (h *APIHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	cmd, err := decodeJSON[struct {
		Name string `json:"name"`
	}](w, r)
	if err != nil {
		return
	}

	if err := h.groupService.UpdateGroup(r.Context(), groupID, cmd.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /groups/{id}
func (h *APIHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")
	if err := h.groupService.DeleteGroup(r.Context(), groupID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /groups/{id}/members/{user_id}
func (h *APIHandler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")
	userID := r.PathValue("user_id")

	if err := h.groupService.RemoveMember(r.Context(), groupID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
