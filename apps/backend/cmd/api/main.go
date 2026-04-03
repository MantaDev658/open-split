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
	expenseRepo := postgres.NewExpenseRepository(db)

	userService := application.NewUserService(userRepo)
	expenseService := application.NewExpenseService(expenseRepo)

	handler := openhttp.NewAPIHandler(expenseService, userService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /expenses", handler.CreateExpense)
	mux.HandleFunc("GET /expenses", handler.ListExpenses)
	mux.HandleFunc("GET /balances", handler.GetBalances)
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("GET /users", handler.ListUsers)

	port := ":8080"
	fmt.Printf("🚀 Open Split API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
