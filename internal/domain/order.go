
package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "PENDING"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusShipped    OrderStatus = "SHIPPED"
	StatusDelivered  OrderStatus = "DELIVERED"
	StatusCancelled  OrderStatus = "CANCELLED"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidStatus      = errors.New("invalid order status")
	ErrCannotCancel       = errors.New("order can only be cancelled when status is PENDING")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrEmptyOrderItems    = errors.New("order must contain at least one item")
	ErrInvalidItem        = errors.New("item quantity must be positive and price must be non-negative")
)

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	Status     OrderStatus `json:"status"`
	Total      float64     `json:"total"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}

func CanTransition(from, to OrderStatus) bool {
	if from == to {
		return true
	}

	switch from {
	case StatusPending:
		return to == StatusProcessing || to == StatusCancelled
	case StatusProcessing:
		return to == StatusShipped
	case StatusShipped:
		return to == StatusDelivered
	default:
		return false
	}
}

func CalculateTotal(items []OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

func ValidateItems(items []OrderItem) error {
	if len(items) == 0 {
		return ErrEmptyOrderItems
	}
	for _, item := range items {
		if item.Quantity <= 0 || item.UnitPrice < 0 {
			return ErrInvalidItem
		}
	}
	return nil
}
