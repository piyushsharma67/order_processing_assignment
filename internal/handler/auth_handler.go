package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"order_processing/internal/auth"
)

type AuthHandler struct {
	auth *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{auth: svc}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/signup", h.signup)
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token      string `json:"token"`
	CustomerID string `json:"customer_id"`
	ExpiresAt  string `json:"expires_at"`
	TokenType  string `json:"token_type"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	h.authenticate(w, r, h.auth.Login)
}

func (h *AuthHandler) signup(w http.ResponseWriter, r *http.Request) {
	h.authenticate(w, r, h.auth.Register)
}

func (h *AuthHandler) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	authenticate func(context.Context, string, string) (string, time.Time, string, error),
) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	token, expiresAt, customerID, err := authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, auth.ErrUsernameTaken):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, auth.ErrWeakPassword):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:      token,
		CustomerID: customerID,
		ExpiresAt:  expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:  "Bearer",
	})
}
