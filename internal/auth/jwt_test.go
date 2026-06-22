package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"order_processing/internal/auth"
	"order_processing/internal/repository"
)

func newTestService(t *testing.T) *auth.Service {
	t.Helper()

	users := repository.NewInMemoryUserRepository()
	svc, err := auth.NewService("test-secret", users, time.Hour)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	ctx := context.Background()
	if err := svc.EnsureUser(ctx, "admin", "secret"); err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	return svc
}

func TestLoginAndValidateToken(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	token, expiresAt, customerID, err := svc.Login(ctx, "admin", "secret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if customerID == "" {
		t.Fatal("expected customer_id")
	}
	if !strings.HasPrefix(customerID, "cust-") {
		t.Fatalf("expected cust- prefix, got %q", customerID)
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expected future expiry")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.Username != "admin" {
		t.Fatalf("expected admin, got %q", claims.Username)
	}
	if claims.CustomerID != customerID {
		t.Fatalf("expected customer_id %q in token, got %q", customerID, claims.CustomerID)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _, _, err := svc.Login(ctx, "admin", "wrong")
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
}

func TestRegisterAndLogin(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	token, _, customerID, err := svc.Register(ctx, "newuser", "password1")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if customerID == "" {
		t.Fatal("expected customer_id")
	}

	_, _, loginCustomerID, err := svc.Login(ctx, "newuser", "password1")
	if err != nil {
		t.Fatalf("login after register: %v", err)
	}
	if loginCustomerID != customerID {
		t.Fatalf("expected same customer_id on login, got %q want %q", loginCustomerID, customerID)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, _, _, err := svc.Register(ctx, "newuser", "password1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, _, _, err := svc.Register(ctx, "newuser", "password2")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.ValidateToken("not-a-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}
