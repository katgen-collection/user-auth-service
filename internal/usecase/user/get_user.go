package user

import (
	"context"

	"mikhailjbs/user-auth-service/internal/domain/user"
)

type GetUserUseCase interface {
	Execute(ctx context.Context, id string) (*user.User, error)
}

type getUserUseCase struct {
	service user.Service
}

func NewGetUserUseCase(service user.Service) GetUserUseCase {
	return &getUserUseCase{service: service}
}

func (uc *getUserUseCase) Execute(ctx context.Context, id string) (*user.User, error) {
	return uc.service.Get(id)
}
