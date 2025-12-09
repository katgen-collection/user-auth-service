package auth

import (
	"errors"

	"mikhailjbs/user-auth-service/internal/domain/session"
	"mikhailjbs/user-auth-service/internal/domain/user"
	"mikhailjbs/user-auth-service/internal/infra/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
)

type Repository interface {
	// Define repository methods if needed
}

type Service interface {
	RegisterUser(r *RegisterRequest) (*user.User, error)
	LoginUser(r *LoginRequest) (*user.User, error)
	ValidateToken(token string) (*session.Session, error)
	InvalidateSession(sessionID string) error
	GetMe(token string) (*user.User, error)
}

type service struct {
	userService    user.Service
	sessionService session.Service
	userRepo       user.Repository
}

func NewService(uSvc user.Service, sSvc session.Service, uRepo user.Repository) Service {
	return &service{
		userService:    uSvc,
		sessionService: sSvc,
		userRepo:       uRepo,
	}
}

func (s *service) RegisterUser(r *RegisterRequest) (*user.User, error) {
	newUserCreate := &user.CreateUserRequest{
		Email:    r.Email,
		Password: r.Password,
		Fullname: r.Fullname,
		Username: r.Username,
		Role:     r.Role,
	}

	newUser, err := s.userService.Create(newUserCreate)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s *service) LoginUser(r *LoginRequest) (*user.User, error) {
	existingUser, err := s.userRepo.GetByEmail(r.Email)
	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		return nil, ErrInvalidCredentials
	}

	if err := security.ComparePassword(existingUser.PasswordHash, r.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return existingUser, nil
}

func (s *service) ValidateToken(token string) (*session.Session, error) {
	sess, err := s.sessionService.GetSessionByID(token)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

func (s *service) InvalidateSession(sessionID string) error {
	return s.sessionService.InvalidateSession(sessionID)
}

func (s *service) GetMe(token string) (*user.User, error) {
	sess, err := s.sessionService.GetSessionByID(token)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	user, err := s.userRepo.Get(sess.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
