package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"time"
)

type AccessToken string

type TokenProvider interface {
	GenerateToken(data models.UserAuthData) (string, error)
	GetTTL() time.Duration
}

type AuthUseCase struct {
	userRepo       repository.UserRepository
	passwordHasher PasswordHasher
	tokenProvider  TokenProvider
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	passwordHasher PasswordHasher,
	tokenProvider TokenProvider,
) *AuthUseCase {
	return &AuthUseCase{userRepo, passwordHasher, tokenProvider}
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (AccessToken, *models.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, models.ErrInvalidCredentials
	}

	if !uc.passwordHasher.Compare(user.PasswordHash, password) {
		return "", nil, models.ErrInvalidCredentials
	}

	token, err := uc.tokenProvider.GenerateToken(models.UserAuthData{ID: user.ID, Role: user.Role})
	if err != nil {
		return "", nil, err
	}

	return AccessToken(token), &user, nil
}
