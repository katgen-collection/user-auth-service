package user

import (
	"context"

	"mikhailjbs/user-auth-service/internal/domain/user"
)

type DeleteUserUseCase interface {
	Execute(ctx context.Context, id string) error
}

type deleteUserUseCase struct {
	service user.Service
}

func NewDeleteUserUseCase(service user.Service) DeleteUserUseCase {
	return &deleteUserUseCase{service: service}
}

func (uc *deleteUserUseCase) Execute(ctx context.Context, id string) error {
	return uc.service.Delete(id)
}
