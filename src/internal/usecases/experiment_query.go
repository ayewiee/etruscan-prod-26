package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/common/pagination"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

func (uc *ExperimentUseCase) Create(
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

func (uc *ExperimentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error) {
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

func (uc *ExperimentUseCase) Update(
	ctx context.Context,
	actor models.UserAuthData,
	experiment *models.Experiment,
) (*models.Experiment, error) {
	// validate the update as an action
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

	// save snapshot
	err = uc.saveSnapshot(ctx, dbExperiment)
	if err != nil {
		return nil, err
	}

	// actual update

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

	// retrieve additional properties after update

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

func (uc *ExperimentUseCase) saveSnapshot(ctx context.Context, exp *models.Experiment) error {
	variantSnapshots := make([]*models.VariantSnapshotData, len(exp.Variants))
	for i, variant := range exp.Variants {
		variantSnapshots[i] = &models.VariantSnapshotData{
			Name:      variant.Name,
			Value:     variant.Value,
			Weight:    variant.Weight,
			IsControl: variant.IsControl,
		}
	}

	snapshotData := models.ExperimentSnapshotData{
		ID:                 exp.ID,
		FlagID:             exp.FlagID,
		Name:               exp.Name,
		Description:        exp.Description,
		Status:             exp.Status,
		Version:            exp.Version,
		AudiencePercentage: exp.AudiencePercentage,
		TargetingRule:      exp.TargetingRule,
		Variants:           variantSnapshots,
	}

	return uc.repo.SaveExperimentSnapshot(ctx, &models.ExperimentSnapshot{
		ExperimentID: exp.ID,
		Version:      exp.Version,
		Data:         &snapshotData,
	})
}

type ExperimentListFilters struct {
	FlagID     *uuid.UUID
	CreatedBy  *uuid.UUID
	Status     *models.ExperimentStatus
	Outcome    *models.ExperimentOutcome
	Pagination pagination.Pagination
}

func (uc *ExperimentUseCase) List(ctx context.Context, actor models.UserAuthData, filters ExperimentListFilters) ([]*models.Experiment, int, error) {
	if actor.Role != models.UserRoleAdmin && filters.CreatedBy != nil {
		return nil, 0, models.NewErrForbidden("You can not filter by experiment creator")
	}

	return uc.repo.List(ctx, repository.ExperimentListFilters{
		FlagID:    filters.FlagID,
		CreatedBy: filters.CreatedBy,
		Status:    filters.Status,
		Outcome:   filters.Outcome,
		Limit:     filters.Pagination.Limit(),
		Offset:    filters.Pagination.Offset(),
	})
}

func (uc *ExperimentUseCase) ListStatusChanges(
	ctx context.Context,
	expId uuid.UUID,
) ([]*models.ExperimentStatusChange, error) {
	dbExperiment, err := uc.GetByID(ctx, expId)
	if err != nil {
		return nil, err
	}

	return uc.repo.ListStatusChanges(ctx, dbExperiment.ID)
}

func (uc *ExperimentUseCase) ListSnapshots(
	ctx context.Context,
	expId uuid.UUID,
) ([]*models.ExperimentSnapshot, error) {
	dbExperiment, err := uc.GetByID(ctx, expId)
	if err != nil {
		return nil, err
	}

	return uc.repo.ListExperimentSnapshots(ctx, dbExperiment.ID)
}
