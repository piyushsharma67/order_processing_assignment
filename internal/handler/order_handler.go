package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"order_processing/internal/auth"
	"order_processing/internal/domain"
	"order_processing/internal/repository"
	"order_processing/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", h.createOrder)
	mux.HandleFunc("GET /orders", h.listOrders)
	mux.HandleFunc("GET /orders/{id}", h.getOrder)
	mux.HandleFunc("PATCH /orders/{id}/status", h.updateStatus)
	mux.HandleFunc("POST /orders/{id}/cancel", h.cancelOrder)
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input service.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(input.CustomerID) == "" {
		input.CustomerID = claims.CustomerID
	}
	if input.CustomerID != claims.CustomerID {
		writeError(w, http.StatusForbidden, "customer_id does not match signed-in user")
		return
	}

	order, err := h.service.CreateOrder(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := r.PathValue("id")
	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if order.CustomerID != claims.CustomerID {
		writeError(w, http.StatusNotFound, domain.ErrOrderNotFound.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	statusParam := strings.TrimSpace(r.URL.Query().Get("status"))
	filter := repository.OrderListFilter{CustomerID: claims.CustomerID}

	if statusParam != "" {
		status := domain.OrderStatus(strings.ToUpper(statusParam))
		filter.Status = &status
	}

	orders, err := h.service.ListOrders(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if orders == nil {
		orders = []*domain.Order{}
	}
	writeJSON(w, http.StatusOK, orders)
}

type updateStatusRequest struct {
	Status domain.OrderStatus `json:"status"`
}

func (h *OrderHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.service.UpdateOrderStatus(r.Context(), id, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := r.PathValue("id")
	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if order.CustomerID != claims.CustomerID {
		writeError(w, http.StatusNotFound, domain.ErrOrderNotFound.Error())
		return
	}

	order, err = h.service.CancelOrder(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrCannotCancel),
		errors.Is(err, domain.ErrEmptyOrderItems),
		errors.Is(err, domain.ErrInvalidItem):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
