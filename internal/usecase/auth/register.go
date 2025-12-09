package auth

import (
	"context"
	"mikhailjbs/user-auth-service/internal/domain/auth"
	"mikhailjbs/user-auth-service/internal/domain/user"
)

type RegisterUseCase interface {
	Execute(ctx context.Context, req *auth.RegisterRequest) (*user.User, error)
}

type registerUseCase struct {
	authService auth.Service
}

func NewRegisterUseCase(authService auth.Service) RegisterUseCase {
	return &registerUseCase{
		authService: authService,
	}
}

func (uc *registerUseCase) Execute(ctx context.Context, req *auth.RegisterRequest) (*user.User, error) {
	userRecord, err := uc.authService.RegisterUser(req)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}