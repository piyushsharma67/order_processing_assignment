package domain_test

import (
	"testing"

	"order_processing/internal/domain"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from domain.OrderStatus
		to   domain.OrderStatus
		want bool
	}{
		{domain.StatusPending, domain.StatusProcessing, true},
		{domain.StatusPending, domain.StatusCancelled, true},
		{domain.StatusPending, domain.StatusShipped, false},
		{domain.StatusProcessing, domain.StatusShipped, true},
		{domain.StatusProcessing, domain.StatusPending, false},
		{domain.StatusShipped, domain.StatusDelivered, true},
		{domain.StatusDelivered, domain.StatusPending, false},
	}

	for _, tt := range tests {
		if got := domain.CanTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestValidateItems(t *testing.T) {
	if err := domain.ValidateItems(nil); err != domain.ErrEmptyOrderItems {
		t.Fatalf("expected ErrEmptyOrderItems, got %v", err)
	}

	if err := domain.ValidateItems([]domain.OrderItem{{Quantity: 0, UnitPrice: 10}}); err != domain.ErrInvalidItem {
		t.Fatalf("expected ErrInvalidItem, got %v", err)
	}

	items := []domain.OrderItem{{Quantity: 2, UnitPrice: 15.5}}
	if err := domain.ValidateItems(items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total := domain.CalculateTotal(items); total != 31 {
		t.Fatalf("expected total 31, got %v", total)
	}
}
