package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	secret := []byte("test-secret")
	middleware := AuthMiddleware(secret)

	// panics if the user_id is not injected into the context correctly
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || userID == "" {
			t.Errorf("expected UserID in context, got none")
		}
		if userID != "Alice" {
			t.Errorf("expected UserID 'Alice', got %s", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware(nextHandler)

	t.Run("Rejects missing Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("Rejects malformed token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.string")
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("Succeeds and injects context with valid token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "Alice",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(secret)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
