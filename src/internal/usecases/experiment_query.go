package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/domain/models"
	"fmt"

	"github.com/google/uuid"
)

func (uc ExperimentUseCase) Create(
	ctx context.Context,
	actor models.UserAuthData,
	experiment *models.Experiment,
) (*models.Experiment, error) {
	if !actor.Role.CanManageExperiments() {
		return nil, models.ErrForbidden
	}

	err := uc.validateExperiment(ctx, experiment)
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

	reviews, err := uc.repo.ListExperimentReviews(ctx, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Reviews = reviews

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

	reviews, err := uc.repo.ListExperimentReviews(ctx, experiment.ID)
	if err != nil {
		return nil, err
	}
	experiment.Reviews = reviews

	return experiment, nil
}

func (uc ExperimentUseCase) Update(
	ctx context.Context,
	actor models.UserAuthData,
	experiment *models.Experiment,
) (*models.Experiment, error) {
	if !actor.Role.CanManageExperiments() {
		return nil, models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, experiment.ID)
	if err != nil {
		return nil, err
	}

	if actor.Role != models.UserRoleAdmin && actor.ID != dbExperiment.CreatedBy {
		return nil, models.ErrForbidden
	}

	if !dbExperiment.Status.EditingAllowed() {
		return nil, models.NewErrLocked(fmt.Sprintf(
			"Experiment with status %s cannot be edited",
			dbExperiment.Status,
		))
	}

	err = uc.validateExperiment(ctx, experiment)
	if err != nil {
		return nil, err
	}

	updExp, err := uc.repo.Update(ctx, experiment)
	if err != nil {
		return nil, err
	}

	err = uc.repo.DeleteVariantsByExperimentID(ctx, updExp.ID)
	if err != nil {
		return nil, err
	}
	err = uc.repo.CreateVariants(ctx, updExp.ID, experiment.Variants)
	if err != nil {
		return nil, err
	}
	variants, err := uc.repo.ListVariantsByExperimentID(ctx, updExp.ID)
	if err != nil {
		return nil, err
	}
	updExp.Variants = variants

	reviews, err := uc.repo.ListExperimentReviews(ctx, updExp.ID)
	if err != nil {
		return nil, err
	}
	updExp.Reviews = reviews

	return updExp, nil
}
