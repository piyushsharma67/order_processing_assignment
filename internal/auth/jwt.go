package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"order_processing/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
)

type Claims struct {
	Username   string `json:"username"`
	CustomerID string `json:"customer_id"`
	jwt.RegisteredClaims
}

type Service struct {
	secret []byte
	users  repository.UserRepository
	expiry time.Duration
}

func NewService(secret string, users repository.UserRepository, expiry time.Duration) (*Service, error) {
	if secret == "" {
		return nil, errors.New("JWT secret must not be empty")
	}
	if users == nil {
		return nil, errors.New("user repository must not be nil")
	}
	return &Service{
		secret: []byte(secret),
		users:  users,
		expiry: expiry,
	}, nil
}

func (s *Service) EnsureUser(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err == nil {
		if user.CustomerID == "" {
			_, err = s.ensureCustomerID(ctx, username)
		}
		return err
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, username, hash, generateCustomerID())
}

func (s *Service) Register(ctx context.Context, username, password string) (string, time.Time, string, error) {
	username = strings.TrimSpace(username)
	if err := validateCredentials(username, password); err != nil {
		return "", time.Time{}, "", err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return "", time.Time{}, "", err
	}

	customerID := generateCustomerID()
	if err := s.users.Create(ctx, username, hash, customerID); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return "", time.Time{}, "", ErrUsernameTaken
		}
		return "", time.Time{}, "", err
	}

	token, expiresAt, err := s.issueToken(username, customerID)
	return token, expiresAt, customerID, err
}

func (s *Service) Login(ctx context.Context, username, password string) (string, time.Time, string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", time.Time{}, "", ErrInvalidCredentials
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", time.Time{}, "", ErrInvalidCredentials
		}
		return "", time.Time{}, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, "", ErrInvalidCredentials
	}

	customerID, err := s.ensureCustomerID(ctx, username)
	if err != nil {
		return "", time.Time{}, "", err
	}

	token, expiresAt, err := s.issueToken(username, customerID)
	return token, expiresAt, customerID, err
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) ensureCustomerID(ctx context.Context, username string) (string, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user.CustomerID != "" {
		return user.CustomerID, nil
	}

	customerID := generateCustomerID()
	if err := s.users.SetCustomerID(ctx, username, customerID); err != nil {
		return "", err
	}
	return customerID, nil
}

func (s *Service) issueToken(username, customerID string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.expiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username:   username,
		CustomerID: customerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	})

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func generateCustomerID() string {
	return "cust-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
}

func validateCredentials(username, password string) error {
	if username == "" || password == "" {
		return ErrInvalidCredentials
	}
	if utf8.RuneCountInString(password) < 6 {
		return ErrWeakPassword
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}
