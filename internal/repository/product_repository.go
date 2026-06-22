package repository

import (
	"context"
	"sync"

	"order_processing/internal/domain"
)

type ProductRepository interface {
	Count(ctx context.Context) (int64, error)
	CreateMany(ctx context.Context, products []domain.Product) error
	List(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id string) (*domain.Product, error)
}

type InMemoryProductRepository struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: make(map[string]domain.Product),
	}
}

func (r *InMemoryProductRepository) Count(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.products)), nil
}

func (r *InMemoryProductRepository) CreateMany(_ context.Context, products []domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, product := range products {
		r.products[product.ID] = product
	}
	return nil
}

func (r *InMemoryProductRepository) List(_ context.Context) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Product, 0, len(r.products))
	for _, product := range r.products {
		result = append(result, product)
	}
	return result, nil
}

func (r *InMemoryProductRepository) GetByID(_ context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	return &product, nil
}
