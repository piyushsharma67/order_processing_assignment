package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"order_processing/internal/domain"
	"order_processing/internal/repository"
)

type OrderService struct {
	repo                repository.OrderRepository
	pendingProcessDelay time.Duration
}

func NewOrderService(repo repository.OrderRepository, pendingProcessDelay time.Duration) *OrderService {
	return &OrderService{
		repo:                repo,
		pendingProcessDelay: pendingProcessDelay,
	}
}

type CreateOrderInput struct {
	CustomerID string             `json:"customer_id"`
	Items      []domain.OrderItem `json:"items"`
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if err := domain.ValidateItems(input.Items); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order := &domain.Order{
		ID:         uuid.NewString(),
		CustomerID: input.CustomerID,
		Items:      append([]domain.OrderItem(nil), input.Items...),
		Status:     domain.StatusPending,
		Total:      domain.CalculateTotal(input.Items),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, filter repository.OrderListFilter) ([]*domain.Order, error) {
	if filter.Status != nil && !filter.Status.IsValid() {
		return nil, domain.ErrInvalidStatus
	}
	return s.repo.List(ctx, filter)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidStatus
	}
	if status == domain.StatusCancelled {
		return nil, domain.ErrCannotCancel
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanTransition(order.Status, status) {
		return nil, domain.ErrInvalidTransition
	}

	order.Status = status
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != domain.StatusPending {
		return nil, domain.ErrCannotCancel
	}

	order.Status = domain.StatusCancelled
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) ProcessPendingOrders(ctx context.Context) (int, error) {
	pending, err := s.repo.ListByStatus(ctx, domain.StatusPending)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	updated := 0
	for _, order := range pending {
		if now.Before(order.CreatedAt.Add(s.pendingProcessDelay)) {
			continue
		}

		order.Status = domain.StatusProcessing
		order.UpdatedAt = now
		if err := s.repo.Update(ctx, order); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
