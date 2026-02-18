package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/domain"
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
	if !actor.Role.CanManageFlags() {
		return nil, models.ErrForbidden
	}

	err := domain.ValidateValueMatchesType(flag.DefaultValue, flag.ValueType)
	if err != nil {
		var fe *models.FieldError
		if errors.As(err, &fe) {
			return nil, apierrors.DumbValidationError(fe.Field, fe.RejectedValue, fe.Issue, fe)
		}
		return nil, err
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
	if !actor.Role.CanManageFlags() {
		return nil, models.ErrForbidden
	}

	err := domain.ValidateValueMatchesType(flag.DefaultValue, flag.ValueType)
	if err != nil {
		var fe *models.FieldError
		if errors.As(err, &fe) {
			return nil, apierrors.DumbValidationError(fe.Field, fe.RejectedValue, fe.Issue, fe)
		}
		return nil, err
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
	if !actor.Role.CanManageFlags() {
		return models.ErrForbidden
	}
	return uc.repo.Delete(ctx, id)
}
