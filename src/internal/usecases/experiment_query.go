package usecases

import (
	"context"
	"errors"
	"etruscan/internal/common/pagination"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	err = uc.setExperimentMetrics(ctx, exp.ID, experiment.MetricKeys, experiment.PrimaryMetricKey)
	if err != nil {
		return nil, err
	}
	exp.MetricKeys = experiment.MetricKeys
	exp.PrimaryMetricKey = experiment.PrimaryMetricKey

	guardrails, err := uc.setGuardrails(ctx, exp.ID, experiment.Guardrails)
	if err != nil {
		return nil, err
	}
	exp.Guardrails = guardrails

	return exp, nil
}

func (uc *ExperimentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error) {
	experiment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	em, err := uc.metricRepo.ListExperimentMetrics(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(em) > 0 {
		experiment.MetricKeys = make([]string, len(em))
		experiment.Metrics = make([]*models.ExperimentMetricRef, 0, len(em))
		for i, m := range em {
			experiment.MetricKeys[i] = m.MetricKey
		}
		for _, m := range em {
			if m.IsPrimary {
				experiment.PrimaryMetricKey = &m.MetricKey
				break
			}
		}
		for _, m := range em {
			metric, err := uc.metricRepo.GetByID(ctx, m.MetricID)
			if err != nil {
				continue
			}
			experiment.Metrics = append(experiment.Metrics, &models.ExperimentMetricRef{
				Metric:    metric,
				IsPrimary: m.IsPrimary,
			})
		}
	}

	guardrails, err := uc.guardrailRepo.ListByExperimentID(ctx, id)
	if err != nil {
		return nil, err
	}
	experiment.Guardrails = guardrails

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

	err = uc.setExperimentMetrics(ctx, updExp.ID, experiment.MetricKeys, experiment.PrimaryMetricKey)
	if err != nil {
		return nil, err
	}
	updExp.MetricKeys = experiment.MetricKeys
	updExp.PrimaryMetricKey = experiment.PrimaryMetricKey

	guardrails, err := uc.setGuardrails(ctx, updExp.ID, experiment.Guardrails)
	if err != nil {
		return nil, err
	}
	updExp.Guardrails = guardrails

	return updExp, nil
}

func (uc *ExperimentUseCase) setGuardrails(
	ctx context.Context,
	experimentID uuid.UUID,
	inputs []*models.Guardrail,
) ([]*models.Guardrail, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	if err := uc.guardrailRepo.DeleteByExperimentID(ctx, experimentID); err != nil {
		return nil, err
	}

	var guardrails []*models.Guardrail

	for _, g := range inputs {
		metric, err := uc.metricRepo.GetByKey(ctx, g.MetricKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, models.NewErrNotFound("Metric not found", map[string]interface{}{"key": g.MetricKey}, err)
			}
			return nil, err
		}

		createdGuardrail, err := uc.guardrailRepo.Create(
			ctx,
			experimentID,
			metric.ID,
			g.Threshold,
			g.ThresholdDirection,
			g.Action,
			g.WindowSeconds,
		)
		if err != nil {
			return nil, err
		}

		guardrails = append(guardrails, createdGuardrail)
	}
	return guardrails, nil
}

func (uc *ExperimentUseCase) setExperimentMetrics(
	ctx context.Context,
	experimentID uuid.UUID,
	metricKeys []string,
	primaryMetricKey *string,
) error {
	if len(metricKeys) == 0 {
		return nil
	}

	if err := uc.metricRepo.DeleteExperimentMetrics(ctx, experimentID); err != nil {
		return err
	}

	for _, key := range metricKeys {
		metric, err := uc.metricRepo.GetByKey(ctx, key)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.NewErrNotFound("Metric not found", map[string]interface{}{"key": key}, err)
			}
			return err
		}

		isPrimary := primaryMetricKey != nil && *primaryMetricKey == key

		if err := uc.metricRepo.AddExperimentMetric(ctx, experimentID, metric.ID, isPrimary); err != nil {
			return err
		}
	}
	return nil
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
