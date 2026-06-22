package repository

import (
	"context"
	"sync"

	"order_processing/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	List(ctx context.Context, filter OrderListFilter) ([]*domain.Order, error)
	ListByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)
}

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (r *InMemoryOrderRepository) Create(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) GetByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return cloneOrder(order), nil
}

func (r *InMemoryOrderRepository) Update(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[order.ID]; !ok {
		return domain.ErrOrderNotFound
	}
	r.orders[order.ID] = cloneOrder(order)
	return nil
}

func (r *InMemoryOrderRepository) List(_ context.Context, filter OrderListFilter) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		if filter.CustomerID != "" && order.CustomerID != filter.CustomerID {
			continue
		}
		if filter.Status != nil && order.Status != *filter.Status {
			continue
		}
		result = append(result, cloneOrder(order))
	}
	return result, nil
}

func (r *InMemoryOrderRepository) ListByStatus(_ context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	return r.List(context.Background(), OrderListFilter{Status: &status})
}

func cloneOrder(order *domain.Order) *domain.Order {
	items := make([]domain.OrderItem, len(order.Items))
	copy(items, order.Items)

	return &domain.Order{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Items:      items,
		Status:     order.Status,
		Total:      order.Total,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}
