package user

import (
	"context"
	"time"

	"mikhailjbs/user-auth-service/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
)

type UpdateUserUseCase interface {
	Execute(ctx context.Context, id string, req *user.UpdateUserRequest) (*user.User, error)
}

type updateUserUseCase struct {
	service user.Service
}

func NewUpdateUserUseCase(service user.Service) UpdateUserUseCase {
	return &updateUserUseCase{service: service}
}

func (uc *updateUserUseCase) Execute(ctx context.Context, id string, req *user.UpdateUserRequest) (*user.User, error) {
	existingUser, err := uc.service.Get(id)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		existingUser.Username = *req.Username
	}

	if req.Fullname != nil {
		existingUser.Fullname = *req.Fullname
	}

	if req.Email != nil {
		existingUser.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		existingUser.PasswordHash = string(hashedPassword)
	}

	existingUser.UpdatedAt = time.Now()

	return uc.service.Update(id, existingUser)
}
