package session

import "errors"

var (
	ErrNotFound = errors.New("session not found")
)

type Repository interface {
	Create(s *Session) error
	GetByID(id string) (*Session, error)
	Invalidate(id string) error
	Delete(id string) error
	Update(id string, s *Session) (*Session, error)
	List(params *SessionQueryParams) ([]*Session, error)
}

type Service interface {
	CreateSession(s *Session) error
	GetSessionByID(id string) (*Session, error)
	InvalidateSession(id string) error
	DeleteSession(id string) error
	UpdateSession(id string, s *Session) (*Session, error)
	ListSessions(params *SessionQueryParams) ([]*Session, error)
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) CreateSession(sess *Session) error {
	return s.repo.Create(sess)
}

func (s *service) GetSessionByID(id string) (*Session, error) {
	return s.repo.GetByID(id)
}

func (s *service) InvalidateSession(id string) error {
	return s.repo.Invalidate(id)
}

func (s *service) DeleteSession(id string) error {
	return s.repo.Delete(id)
}

func (s *service) UpdateSession(id string, sess *Session) (*Session, error) {
	return s.repo.Update(id, sess)
}

func (s *service) ListSessions(params *SessionQueryParams) ([]*Session, error) {
	return s.repo.List(params)
}
