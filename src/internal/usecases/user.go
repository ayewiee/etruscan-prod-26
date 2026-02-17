package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/common/pagination"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type UserUseCase struct {
	repo           repository.UserRepository
	passwordHasher PasswordHasher
}

func NewUserUseCase(repo repository.UserRepository, hasher PasswordHasher) *UserUseCase {
	return &UserUseCase{repo: repo, passwordHasher: hasher}
}

func (uc *UserUseCase) Create(
	ctx context.Context,
	actor models.UserAuthData,
	user models.User,
	password string,
) (*models.User, error) {
	if actor.Role != models.UserRoleAdmin {
		return nil, models.ErrForbidden
	}

	passwordHash, err := uc.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = passwordHash

	usr, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func (uc *UserUseCase) AdminUpdate(
	ctx context.Context,
	actor models.UserAuthData,
	user models.User,
	updUser models.User,
	password *string,
) (*models.User, error) {
	if actor.Role != models.UserRoleAdmin {
		return nil, models.ErrForbidden
	}

	updUser.ID = user.ID

	if password != nil {
		passwordHash, err := uc.passwordHasher.Hash(*password)
		if err != nil {
			return nil, err
		}
		updUser.PasswordHash = passwordHash
	} else {
		updUser.PasswordHash = user.PasswordHash
	}

	usr, err := uc.repo.AdminUpdate(ctx, updUser)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func (uc *UserUseCase) GetByID(ctx context.Context, actor models.UserAuthData, id uuid.UUID) (*models.User, error) {
	if actor.Role != models.UserRoleAdmin && id != actor.ID {
		return nil, models.ErrForbidden
	}

	user, err := uc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("User not found", nil, err)
		}
	}
	return &user, nil
}

func (uc *UserUseCase) List(
	ctx context.Context,
	actor models.UserAuthData,
	pgn pagination.Pagination,
) ([]models.User, int, error) {
	if actor.Role != models.UserRoleAdmin {
		return nil, 0, models.ErrForbidden
	}

	return uc.repo.List(ctx, pgn.Limit(), pgn.Offset())
}

func (uc *UserUseCase) SoftDelete(ctx context.Context, actor models.UserAuthData, id uuid.UUID) error {
	if actor.Role != models.UserRoleAdmin {
		return models.ErrForbidden
	}

	return uc.repo.SoftDelete(ctx, id)
}
