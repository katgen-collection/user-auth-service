package auth

import (
	"context"
	"mikhailjbs/user-auth-service/internal/domain/auth"
	"mikhailjbs/user-auth-service/internal/domain/user"
)

type GetMeUseCase interface {
	Execute(ctx context.Context, token string) (*user.User, error)
}

type getMeUseCase struct {
	authService auth.Service
}

func NewGetMeUseCase(authService auth.Service) GetMeUseCase {
	return &getMeUseCase{
		authService: authService,
	}
}

func (uc *getMeUseCase) Execute(ctx context.Context, token string) (*user.User, error) {
	userRecord, err := uc.authService.GetMe(token)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}