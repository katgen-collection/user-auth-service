package user

import (
	"context"
	"mikhailjbs/user-auth-service/internal/domain/user"
)

type GetUsersUseCase interface {
	Execute(ctx context.Context, params *user.UserQueryParams) ([]*user.User, error)
}

type getUsersUseCase struct {
	service user.Service
}

func NewGetUsersUseCase(service user.Service) GetUsersUseCase {
	return &getUsersUseCase{service: service}
}

func (uc *getUsersUseCase) Execute(ctx context.Context, params *user.UserQueryParams) ([]*user.User, error) {
	return uc.service.List(params)
}
