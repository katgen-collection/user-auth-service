package user

import (
	"context"

	"mikhailjbs/user-auth-service/internal/domain/user"
)

type CreateUserUseCase interface {
	Execute(ctx context.Context, req *user.CreateUserRequest) (*user.User, error)
}

type createUserUseCase struct {
	service user.Service
}

func NewCreateUserUseCase(service user.Service) CreateUserUseCase {
	return &createUserUseCase{service: service}
}

func (uc *createUserUseCase) Execute(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {

	if req.Role == "" {
		req.Role = string(user.RoleUser)
	}

	newUser, err := uc.service.Create(req)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}
