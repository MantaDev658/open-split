package http

import (
	"encoding/json"
	"net/http"

	"opensplit/apps/backend/internal/expense/application"
	"opensplit/apps/backend/internal/expense/domain"
)

type APIHandler struct {
	expenseService *application.ExpenseService
	userService    *application.UserService
}

func NewAPIHandler(es *application.ExpenseService, us *application.UserService) *APIHandler {
	return &APIHandler{
		expenseService: es,
		userService:    us,
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

// GET /expenses
func (h *APIHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.expenseService.ListAllExpenses(r.Context())
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
			Total:       exp.TotalAmount().Int64(),
			Payer:       string(exp.Payer()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return
	}
}

// GET /balances
func (h *APIHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.expenseService.ListAllExpenses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	netBalances := domain.CalculateNetBalances(expenses)
	transactions := domain.SimplifyDebts(netBalances)

	response := struct {
		Balances     map[domain.UserID]int64 `json:"net_balances"`
		Transactions []domain.Transaction    `json:"suggested_settlements"`
	}{
		Balances:     netBalances,
		Transactions: transactions,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
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
