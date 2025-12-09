package repository

import (
	"errors"

	"mikhailjbs/user-auth-service/internal/domain/user"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *user.User) (*user.User, error) {
	// Store the user to db

	createdUser := u
	if err := r.db.Create(createdUser).Error; err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (r *userRepository) GetByEmail(email string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetAll(params *user.UserQueryParams) ([]*user.User, error) {
	var users []*user.User
	query := r.db.Model(&user.User{})

	if params != nil {
		if params.Email != nil {
			query = query.Where("email = ?", *params.Email)
		}
		if params.Role != nil {
			query = query.Where("role = ?", *params.Role)
		}
		if params.Search != nil {
			search := "%" + *params.Search + "%"
			query = query.Where("username LIKE ? OR fullname LIKE ? OR email LIKE ?", search, search, search)
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Get(id string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&user.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *userRepository) Update(id string, u *user.User) (*user.User, error) {
	if err := r.db.Model(&user.User{}).Where("id = ?", id).Updates(u).Error; err != nil {
		return nil, err
	}
	return r.Get(id)
}
