package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/domain"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"etruscan/internal/repository/cache"
	"fmt"

	"github.com/labstack/echo/v4"
)

type ExperimentUseCase struct {
	repo                repository.ExperimentRepository
	flagRepo            repository.FlagRepository
	userRepo            repository.UserRepository
	approverGroupRepo   repository.ApproverGroupRepository
	metricRepo          repository.MetricRepository
	guardrailRepo       repository.GuardrailRepository
	runningExpCache     *cache.RunningExperimentCache
	defaultMinApprovals int
	notifications       *NotificationRouter
}

func NewExperimentUseCase(
	repo repository.ExperimentRepository,
	flagRepo repository.FlagRepository,
	userRepo repository.UserRepository,
	approverGroupRepo repository.ApproverGroupRepository,
	metricRepo repository.MetricRepository,
	guardrailRepo repository.GuardrailRepository,
	runningExpCache *cache.RunningExperimentCache,
	defaultMinApprovals int,
	notifications *NotificationRouter,
) *ExperimentUseCase {
	return &ExperimentUseCase{
		repo:                repo,
		flagRepo:            flagRepo,
		userRepo:            userRepo,
		approverGroupRepo:   approverGroupRepo,
		metricRepo:          metricRepo,
		guardrailRepo:       guardrailRepo,
		runningExpCache:     runningExpCache,
		defaultMinApprovals: defaultMinApprovals,
		notifications:       notifications,
	}
}

func (uc *ExperimentUseCase) validateExperiment(ctx context.Context, experiment *models.Experiment) error {
	// check that flag exists
	flag, err := uc.flagRepo.GetByID(ctx, experiment.FlagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.NewErrNotFound("Flag not found", echo.Map{"flagId": experiment.FlagID}, err)
		}
		return err
	}

	err = validateVariants(flag, experiment.Variants)
	if err != nil {
		return err
	}

	return nil
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

// updateStatus well, updates status, logs it and notifies about it.
// status transition is not being validated!!!
func (uc *ExperimentUseCase) updateStatus(
	ctx context.Context,
	experiment *models.Experiment,
	statusChange *models.ExperimentStatusChange,
) (err error) {
	err = uc.repo.UpdateStatus(ctx, statusChange.ExperimentID, statusChange.To)
	if err != nil {
		return
	}
	err = uc.repo.LogExperimentStatusChange(ctx, statusChange)

	if statusChange.ActorID != nil { // user updated status
		NotifyExperimentStatusChangedUser(
			ctx,
			uc.notifications,
			experiment,
			statusChange.From,
			statusChange.To,
			nil,
		)
	} else { // system status update
		NotifyExperimentStatusChangedSystem(
			ctx,
			uc.notifications,
			experiment,
			statusChange.From,
			statusChange.To,
			nil,
		)
	}

	return
}
