package auth

import (
	"context"
	"mikhailjbs/user-auth-service/internal/domain/auth"
	"mikhailjbs/user-auth-service/internal/domain/user"
)

type LoginUseCase interface {
	Execute(ctx context.Context, req *auth.LoginRequest) (*user.User, error)
}

type loginUseCase struct {
	authService auth.Service
}

func NewLoginUseCase(authService auth.Service) LoginUseCase {
	return &loginUseCase{
		authService: authService,
	}
}

func (uc *loginUseCase) Execute(ctx context.Context, req *auth.LoginRequest) (*user.User, error) {
	userRecord, err := uc.authService.LoginUser(req)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}
