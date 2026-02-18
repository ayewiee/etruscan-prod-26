package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/domain"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ExperimentUseCase struct {
	repo     repository.ExperimentRepository
	flagRepo repository.FlagRepository
}

func NewExperimentUseCase(
	repo repository.ExperimentRepository,
	flagRepo repository.FlagRepository,
) *ExperimentUseCase {
	return &ExperimentUseCase{repo: repo, flagRepo: flagRepo}
}

func (uc ExperimentUseCase) Create(
	ctx context.Context,
	actor models.UserAuthData,
	experiment *models.Experiment,
) (*models.Experiment, error) {
	if !actor.Role.CanManageExperiments() {
		return nil, models.ErrForbidden
	}

	// check that flag exists
	flag, err := uc.flagRepo.GetByID(ctx, experiment.FlagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("Flag not found", echo.Map{"flagId": experiment.FlagID}, err)
		}
		return nil, err
	}

	err = validateVariants(flag, experiment.Variants)
	if err != nil {
		return nil, err
	}

	experiment.CreatedBy = actor.ID
	experiment.Status = models.ExperimentStatusDraft

	exp, err := uc.repo.Create(ctx, experiment)
	if err != nil {
		return nil, err
	}

	err = uc.repo.CreateVariants(ctx, exp.ID, experiment.Variants)
	if err != nil {
		return nil, err
	}
	variants, err := uc.repo.ListVariantsByExperimentID(ctx, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants

	return exp, nil
}

func (uc ExperimentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error) {
	experiment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewErrNotFound("Experiment not found", nil, err)
		}
		return nil, err
	}

	variants, err := uc.repo.ListVariantsByExperimentID(ctx, id)
	if err != nil {
		return nil, err
	}
	experiment.Variants = variants

	return experiment, nil
}

func validateVariants(flag *models.Flag, variants []*models.Variant) error {
	var ferrs []models.FieldError

	controlCounter := 0
	weightsSum := 0

	for idx, v := range variants {
		if v.IsControl {
			controlCounter++
		}
		weightsSum += v.Weight

		err := domain.ValidateValueMatchesType(v.Value, flag.ValueType)
		if err != nil {
			var fe *models.FieldError
			if errors.As(err, &fe) {
				fe.Field = fmt.Sprintf("variants[%d].value", idx)
				ferrs = append(ferrs, *fe)
			} else {
				return err
			}
		}
	}

	if controlCounter != 1 {
		ferrs = append(ferrs, models.FieldError{
			Field:         "variants.IsControl",
			Issue:         fmt.Sprintf("there must be exactly 1 control variant (isControl: true) — %d now", controlCounter),
			RejectedValue: nil,
		})
	}

	if weightsSum != 100 {
		ferrs = append(ferrs, models.FieldError{
			Field:         "variants.weight",
			Issue:         fmt.Sprintf("weights must add up to exactly 100 — %d now", weightsSum),
			RejectedValue: nil,
		})
	}

	if len(ferrs) > 0 {
		return apierrors.MultipleDumbValidationErrors(ferrs...)
	}

	return nil
}
