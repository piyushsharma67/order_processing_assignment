package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order_processing/internal/auth"
	"order_processing/internal/handler"
	"order_processing/internal/repository"
)

func newTestAuthHandler(t *testing.T) *handler.AuthHandler {
	t.Helper()

	users := repository.NewInMemoryUserRepository()
	svc, err := auth.NewService("test-secret", users, time.Hour)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	if err := svc.EnsureUser(context.Background(), "admin", "secret"); err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	return handler.NewAuthHandler(svc)
}

func TestAuthLogin(t *testing.T) {
	h := newTestAuthHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"username":"admin","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["token"] == "" {
		t.Fatal("expected token in response")
	}
	if resp["customer_id"] == "" {
		t.Fatal("expected customer_id in response")
	}
}

func TestAuthLoginInvalidCredentials(t *testing.T) {
	h := newTestAuthHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthSignup(t *testing.T) {
	h := newTestAuthHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"username":"alice","password":"password1"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["token"] == "" {
		t.Fatal("expected token in response")
	}
	if resp["customer_id"] == "" {
		t.Fatal("expected customer_id in response")
	}
}

func TestAuthSignupDuplicateUsername(t *testing.T) {
	h := newTestAuthHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"username":"alice","password":"password1"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}
