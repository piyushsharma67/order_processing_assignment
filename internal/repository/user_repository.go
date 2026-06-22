package repository

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRecord struct {
	Username     string
	PasswordHash string
	CustomerID   string
}

type UserRepository interface {
	Create(ctx context.Context, username, passwordHash, customerID string) error
	GetByUsername(ctx context.Context, username string) (UserRecord, error)
	SetCustomerID(ctx context.Context, username, customerID string) error
}

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]UserRecord
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]UserRecord),
	}
}

func (r *InMemoryUserRepository) Create(_ context.Context, username, passwordHash, customerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[username]; ok {
		return ErrUserExists
	}
	r.users[username] = UserRecord{
		Username:     username,
		PasswordHash: passwordHash,
		CustomerID:   customerID,
	}
	return nil
}

func (r *InMemoryUserRepository) GetByUsername(_ context.Context, username string) (UserRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return UserRecord{}, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) SetCustomerID(_ context.Context, username, customerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[username]
	if !ok {
		return ErrUserNotFound
	}
	user.CustomerID = customerID
	r.users[username] = user
	return nil
}
