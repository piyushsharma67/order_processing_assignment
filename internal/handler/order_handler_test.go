package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"order_processing/internal/auth"
	"order_processing/internal/domain"
	"order_processing/internal/handler"
	"order_processing/internal/repository"
	"order_processing/internal/service"
)

const testCustomerID = "cust-test-42"

func setupHandler() *handler.OrderHandler {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 0)
	return handler.NewOrderHandler(svc)
}

func withClaims(req *http.Request) *http.Request {
	claims := &auth.Claims{
		Username:   "testuser",
		CustomerID: testCustomerID,
	}
	return req.WithContext(auth.ContextWithClaims(context.Background(), claims))
}

func TestCreateAndGetOrder(t *testing.T) {
	h := setupHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := map[string]any{
		"customer_id": testCustomerID,
		"items": []map[string]any{
			{"product_id": "p1", "product_name": "Widget", "quantity": 2, "unit_price": 9.99},
		},
	}
	payload, _ := json.Marshal(body)

	createReq := withClaims(httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(payload)))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRec.Code, createRec.Body.String())
	}

	var created domain.Order
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	getReq := withClaims(httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil))
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}
}

func TestListOrdersByStatus(t *testing.T) {
	h := setupHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"customer_id":"` + testCustomerID + `","items":[{"product_id":"p1","quantity":1,"unit_price":1}]}`)
	createReq := withClaims(httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body)))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	listReq := withClaims(httptest.NewRequest(http.MethodGet, "/orders?status=PENDING", nil))
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}

	var orders []domain.Order
	if err := json.Unmarshal(listRec.Body.Bytes(), &orders); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestListOrdersRequiresAuth(t *testing.T) {
	h := setupHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	listReq := httptest.NewRequest(http.MethodGet, "/orders", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", listRec.Code)
	}
}
