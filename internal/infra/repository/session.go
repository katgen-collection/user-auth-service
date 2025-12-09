package repository

import (
	"errors"

	"mikhailjbs/user-auth-service/internal/domain/session"

	"gorm.io/gorm"
)

type Repository interface {
	Create(s *session.Session) error
	GetByID(id string) (*session.Session, error)
	Invalidate(id string) error
	Delete(id string) error
	Update(id string, s *session.Session) (*session.Session, error)
	List(params *session.SessionQueryParams) ([]*session.Session, error)
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) session.Repository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(s *session.Session) error {
	return r.db.Create(s).Error
}

func (r *sessionRepository) GetByID(id string) (*session.Session, error) {
	var s session.Session
	if err := r.db.Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *sessionRepository) Invalidate(id string) error {
	return r.db.Model(&session.Session{}).Where("id = ?", id).Update("valid", false).Error
}

func (r *sessionRepository) Delete(id string) error {
	return r.db.Delete(&session.Session{}, "id = ?", id).Error
}

func (r *sessionRepository) Update(id string, s *session.Session) (*session.Session, error) {
	if err := r.db.Model(&session.Session{}).Where("id = ?", id).Updates(s).Error; err != nil {
		return nil, err
	}

	return s, nil
}

func (r *sessionRepository) List(params *session.SessionQueryParams) ([]*session.Session, error) {
	var sessions []*session.Session
	query := r.db.Model(&session.Session{})
	if params != nil {
		if params.UserID != nil {
			query = query.Where("user_id = ?", *params.UserID)
		}
		if params.Valid != nil {
			query = query.Where("valid = ?", *params.Valid)
		}
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
