package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"opensplit/apps/backend/internal/core/application"
	openhttp "opensplit/apps/backend/internal/core/infrastructure/http"
	"opensplit/apps/backend/internal/core/infrastructure/postgres"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@localhost:5432/opensplit?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Could not connect to DB: %v", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	groupRepo := postgres.NewGroupRepository(db)
	expenseRepo := postgres.NewExpenseRepository(db)

	userService := application.NewUserService(userRepo)
	groupService := application.NewGroupService(groupRepo)
	expenseService := application.NewExpenseService(expenseRepo, groupRepo)

	handler := openhttp.NewAPIHandler(expenseService, userService, groupService)

	mux := http.NewServeMux()

	// Expense
	mux.HandleFunc("POST /expenses", handler.CreateExpense)
	mux.HandleFunc("GET /expenses", handler.ListExpenses)
	mux.HandleFunc("GET /balances", handler.GetBalances)

	// User
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("GET /users", handler.ListUsers)

	// Group
	mux.HandleFunc("POST /groups", handler.CreateGroup)
	mux.HandleFunc("POST /groups/{id}/members", handler.AddGroupMember)
	mux.HandleFunc("GET /groups", handler.ListGroups)

	port := ":8080"
	fmt.Printf("🚀 Open Split API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
