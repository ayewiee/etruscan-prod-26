package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/domain/models"
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (uc *ExperimentUseCase) Launch(ctx context.Context, actor models.UserAuthData, expId uuid.UUID) error {
	if !actor.Role.CanManageExperiments() {
		return models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, expId)
	if err != nil {
		return err
	}

	if actor.Role != models.UserRoleAdmin && actor.ID != dbExperiment.CreatedBy {
		return models.ErrForbidden
	}

	if !dbExperiment.Status.CanTransitionTo(models.ExperimentStatusLaunched) {
		return models.NewErrForbidden(fmt.Sprintf(
			"Experiment with status %s cannot be launched",
			dbExperiment.Status,
		))
	}

	flag, err := uc.flagRepo.GetByID(ctx, dbExperiment.FlagID)
	if err != nil {
		return err
	}

	activeExperiment, err := uc.repo.GetRunningExperimentByFlagKey(ctx, flag.Key)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if activeExperiment != nil && activeExperiment.ID != dbExperiment.ID {
		return models.NewApiError(
			models.ErrCodeForbidden,
			"There's already an active experiment for this flag",
			echo.Map{"activeExperimentId": activeExperiment.ID},
			nil, nil,
		)
	}

	err = uc.updateStatus(ctx, &models.ExperimentStatusChange{
		ExperimentID: dbExperiment.ID,
		ActorID:      &actor.ID,
		From:         &dbExperiment.Status,
		To:           models.ExperimentStatusLaunched,
		Comment:      nil,
	})
	if err != nil {
		return err
	}

	// cache launched experiment
	_ = uc.runningExpCache.Set(ctx, flag.Key, dbExperiment)
	return nil
}

func (uc *ExperimentUseCase) Pause(ctx context.Context, actor models.UserAuthData, expId uuid.UUID) error {
	if !actor.Role.CanManageExperiments() {
		return models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, expId)
	if err != nil {
		return err
	}

	if actor.Role != models.UserRoleAdmin && actor.ID != dbExperiment.CreatedBy {
		return models.ErrForbidden
	}

	if !dbExperiment.Status.CanTransitionTo(models.ExperimentStatusPaused) {
		return models.NewErrForbidden(fmt.Sprintf(
			"Experiment with status %s cannot be paused",
			dbExperiment.Status,
		))
	}

	flag, err := uc.flagRepo.GetByID(ctx, dbExperiment.FlagID)
	if err != nil {
		return err
	}

	err = uc.updateStatus(ctx, &models.ExperimentStatusChange{
		ExperimentID: dbExperiment.ID,
		ActorID:      &actor.ID,
		From:         &dbExperiment.Status,
		To:           models.ExperimentStatusPaused,
		Comment:      nil,
	})
	if err != nil {
		return err
	}

	// remove experiment from cache
	_ = uc.runningExpCache.Set(ctx, flag.Key, dbExperiment)
	return nil
}

func (uc *ExperimentUseCase) Finish(ctx context.Context, actor models.UserAuthData, expID uuid.UUID, outcome models.ExperimentOutcome, comment string) (*models.Experiment, error) {
	if !actor.Role.CanManageExperiments() {
		return nil, models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, expID)
	if err != nil {
		return nil, err
	}

	if actor.Role != models.UserRoleAdmin && actor.ID != dbExperiment.CreatedBy {
		return nil, models.ErrForbidden
	}

	if !dbExperiment.Status.CanTransitionTo(models.ExperimentStatusFinished) {
		return nil, models.NewErrForbidden(fmt.Sprintf(
			"Experiment with status %s cannot be finished",
			dbExperiment.Status,
		))
	}

	exp, err := uc.repo.Finish(ctx, expID, outcome, comment, actor.ID)
	if err != nil {
		return nil, err
	}

	err = uc.repo.LogExperimentStatusChange(ctx, &models.ExperimentStatusChange{
		ExperimentID: expID,
		ActorID:      &actor.ID,
		From:         &dbExperiment.Status,
		To:           models.ExperimentStatusFinished,
		Comment:      &comment,
	})
	if err != nil {
		return nil, err
	}

	return exp, nil
}
