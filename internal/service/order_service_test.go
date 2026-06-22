package service_test

import (
	"context"
	"testing"
	"time"

	"order_processing/internal/domain"
	"order_processing/internal/repository"
	"order_processing/internal/service"
)

func TestOrderServiceFlow(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 0)
	ctx := context.Background()

	order, err := svc.CreateOrder(ctx, service.CreateOrderInput{
		CustomerID: "cust-1",
		Items: []domain.OrderItem{
			{ProductID: "p1", ProductName: "Book", Quantity: 2, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.Status != domain.StatusPending {
		t.Fatalf("expected PENDING, got %s", order.Status)
	}

	got, err := svc.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.Total != 20 {
		t.Fatalf("expected total 20, got %v", got.Total)
	}

	cancelled, err := svc.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", cancelled.Status)
	}
}

func TestCancelOnlyPending(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 0)
	ctx := context.Background()

	order, err := svc.CreateOrder(ctx, service.CreateOrderInput{
		CustomerID: "cust-1",
		Items:      []domain.OrderItem{{ProductID: "p1", Quantity: 1, UnitPrice: 5}},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if _, err := svc.ProcessPendingOrders(ctx); err != nil {
		t.Fatalf("ProcessPendingOrders failed: %v", err)
	}

	if _, err := svc.CancelOrder(ctx, order.ID); err != domain.ErrCannotCancel {
		t.Fatalf("expected ErrCannotCancel, got %v", err)
	}
}

func TestProcessPendingOrders(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 0)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateOrder(ctx, service.CreateOrderInput{
			CustomerID: "cust-1",
			Items:      []domain.OrderItem{{ProductID: "p1", Quantity: 1, UnitPrice: 1}},
		})
		if err != nil {
			t.Fatalf("CreateOrder failed: %v", err)
		}
	}

	count, err := svc.ProcessPendingOrders(ctx)
	if err != nil {
		t.Fatalf("ProcessPendingOrders failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 updates, got %d", count)
	}

	status := domain.StatusProcessing
	orders, err := svc.ListOrders(ctx, repository.OrderListFilter{Status: &status})
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if len(orders) != 3 {
		t.Fatalf("expected 3 processing orders, got %d", len(orders))
	}
}

func TestProcessPendingOrdersRespectsDelay(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 100*time.Millisecond)
	ctx := context.Background()

	_, err := svc.CreateOrder(ctx, service.CreateOrderInput{
		CustomerID: "cust-1",
		Items:      []domain.OrderItem{{ProductID: "p1", Quantity: 1, UnitPrice: 1}},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	count, err := svc.ProcessPendingOrders(ctx)
	if err != nil {
		t.Fatalf("ProcessPendingOrders failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 updates before delay elapsed, got %d", count)
	}

	time.Sleep(150 * time.Millisecond)

	count, err = svc.ProcessPendingOrders(ctx)
	if err != nil {
		t.Fatalf("ProcessPendingOrders failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 update after delay elapsed, got %d", count)
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	repo := repository.NewInMemoryOrderRepository()
	svc := service.NewOrderService(repo, 0)
	ctx := context.Background()

	order, err := svc.CreateOrder(ctx, service.CreateOrderInput{
		CustomerID: "cust-1",
		Items:      []domain.OrderItem{{ProductID: "p1", Quantity: 1, UnitPrice: 10}},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if _, err := svc.ProcessPendingOrders(ctx); err != nil {
		t.Fatalf("ProcessPendingOrders failed: %v", err)
	}

	updated, err := svc.UpdateOrderStatus(ctx, order.ID, domain.StatusShipped)
	if err != nil {
		t.Fatalf("UpdateOrderStatus failed: %v", err)
	}
	if updated.Status != domain.StatusShipped {
		t.Fatalf("expected SHIPPED, got %s", updated.Status)
	}
}
