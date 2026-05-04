package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"opensplit/apps/backend/internal/core/application"
	openhttp "opensplit/apps/backend/internal/core/infrastructure/http"
	"opensplit/apps/backend/internal/core/infrastructure/postgres"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	rawDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer rawDB.Close()

	db := postgres.NewDB(rawDB)

	auditRepo := postgres.NewAuditRepository(rawDB)
	userRepo := postgres.NewUserRepository(rawDB)
	groupRepo := postgres.NewGroupRepository(rawDB)
	expenseRepo := postgres.NewExpenseRepository(rawDB)

	userService := application.NewUserService(userRepo, []byte(jwtSecret))
	groupService := application.NewGroupService(groupRepo, expenseRepo, auditRepo, db)
	expenseService := application.NewExpenseService(expenseRepo, groupRepo, auditRepo, db)

	handler := openhttp.NewAPIHandler(expenseService, userService, groupService)

	// Protected Routes (Requires Auth)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("POST /expenses", handler.CreateExpense)
	protectedMux.HandleFunc("GET /expenses", handler.ListExpenses)
	protectedMux.HandleFunc("GET /balances", handler.GetBalances)
	protectedMux.HandleFunc("GET /friends/{user_id}/balances", handler.GetFriendBalances)
	protectedMux.HandleFunc("PUT /expenses/{id}", handler.UpdateExpense)
	protectedMux.HandleFunc("DELETE /expenses/{id}", handler.DeleteExpense)
	protectedMux.HandleFunc("POST /settlements", handler.CreateSettlement)

	protectedMux.HandleFunc("GET /users", handler.ListUsers)
	protectedMux.HandleFunc("PUT /users/{id}", handler.UpdateUser)
	protectedMux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)

	protectedMux.HandleFunc("POST /groups", handler.CreateGroup)
	protectedMux.HandleFunc("POST /groups/{id}/members", handler.AddGroupMember)
	protectedMux.HandleFunc("GET /groups", handler.ListGroups)
	protectedMux.HandleFunc("PUT /groups/{id}", handler.UpdateGroup)
	protectedMux.HandleFunc("DELETE /groups/{id}", handler.DeleteGroup)
	protectedMux.HandleFunc("DELETE /groups/{id}/members/{user_id}", handler.RemoveGroupMember)
	protectedMux.HandleFunc("GET /groups/{id}/activity", handler.GetGroupActivity)

	authMiddleware := openhttp.AuthMiddleware([]byte(jwtSecret))
	protectedHandler := authMiddleware(protectedMux)

	// Public Routes
	mainMux := http.NewServeMux()
	mainMux.HandleFunc("POST /auth/register", handler.RegisterUser)
	mainMux.HandleFunc("POST /auth/login", handler.LoginUser)

	// Delegate all other routes to the protected handler
	mainMux.Handle("/", protectedHandler)

	port := ":8080"
	fmt.Printf("API running on port %s\n", port)
	server := &http.Server{
		Addr:              port,
		Handler:           http.TimeoutHandler(mainMux, 10*time.Second, `{"error":"request timeout"}`),
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server crashed: %v", err)
	}
}
