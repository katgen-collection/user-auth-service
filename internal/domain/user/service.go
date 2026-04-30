package user

import (
	"errors"
	"strings"
	"time"
	"context"

	"mikhailjbs/user-auth-service/internal/infra/security"

	"github.com/google/uuid"
)

var (
	ErrEmailTaken = errors.New("email already exists")
	ErrNotFound   = errors.New("user not found")
)

// Repository defines the interface for user persistence
type Repository interface {
	Create(u *User) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll(params *UserQueryParams) ([]*User, error)
	Get(id string) (*User, error)
	Delete(id string) error
	Update(id string, u *User) (*User, error)
}

// Service defines the interface for user domain logic
type Service interface {
	Create(u *CreateUserRequest) (*User, error)
	List(params *UserQueryParams) ([]*User, error)
	Get(id string) (*User, error)
	Delete(id string) error
	Update(id string, u *User) (*User, error)
}

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, u *User) error
	PublishUserUpdated(ctx context.Context, u *User) error
}

type service struct {
	repo      Repository
	publisher EventPublisher
}

func NewService(r Repository, p EventPublisher) Service {
	return &service{repo: r, publisher: p}
}

func (s *service) Create(u *CreateUserRequest) (*User, error) {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if u.Role == "" {
		u.Role = "user"
	}

	existing, err := s.repo.GetByEmail(u.Email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrEmailTaken
	}

	// Hash the password before storing
	hashedPassword, err := security.HashPassword(u.Password)
	if err != nil {
		return nil, err
	}

	newUser := &User{
		ID:           uuid.New().String(),
		Email:        u.Email,
		PasswordHash: hashedPassword,
		Fullname:     u.Fullname,
		Username:     u.Username,
		Role:         Role(u.Role),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUser, err := s.repo.Create(newUser)
	if err == nil && s.publisher != nil {
		s.publisher.PublishUserCreated(context.Background(), createdUser)
	}
	return createdUser, err
}

func (s *service) List(params *UserQueryParams) ([]*User, error) {
	return s.repo.GetAll(params)
}

func (s *service) Get(id string) (*User, error) {
	return s.repo.Get(id)
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *service) Update(id string, u *User) (*User, error) {
	updatedUser, err := s.repo.Update(id, u)
	if err == nil && s.publisher != nil {
		s.publisher.PublishUserUpdated(context.Background(), updatedUser)
	}
	return updatedUser, err
}
