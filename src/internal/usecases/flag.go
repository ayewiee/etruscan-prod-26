package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/google/uuid"
)

type FlagUseCase struct {
	repo repository.FlagRepository
}

func NewFlagUseCase(repo repository.FlagRepository) *FlagUseCase {
	return &FlagUseCase{repo: repo}
}

func (uc *FlagUseCase) Create(ctx context.Context, actor models.UserAuthData, flag *models.Flag) (*models.Flag, error) {
	if !actor.Role.CanEditFlags() {
		return nil, models.ErrForbidden
	}

	return uc.repo.Create(ctx, flag)
}

func (uc *FlagUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Flag, error) {
	flag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("Flag not found", nil, err)
		}
	}
	return flag, nil
}

func (uc *FlagUseCase) GetByKey(ctx context.Context, key string) (*models.Flag, error) {
	flag, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("Flag not found", nil, err)
		}
	}
	return flag, nil
}

func (uc *FlagUseCase) List(ctx context.Context) ([]*models.Flag, error) {
	return uc.repo.List(ctx)
}

func (uc *FlagUseCase) Update(ctx context.Context, actor models.UserAuthData, flag *models.Flag) (*models.Flag, error) {
	if !actor.Role.CanEditFlags() {
		return nil, models.ErrForbidden
	}
	updFlag, err := uc.repo.Update(ctx, flag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("Flag not found", nil, err)
		}
	}
	return updFlag, nil
}

func (uc *FlagUseCase) Delete(ctx context.Context, actor models.UserAuthData, id uuid.UUID) error {
	if !actor.Role.CanEditFlags() {
		return models.ErrForbidden
	}
	return uc.repo.Delete(ctx, id)
}
