package repository

import "order_processing/internal/domain"

type OrderListFilter struct {
	CustomerID string
	Status     *domain.OrderStatus
}
