package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type MetricUseCase struct {
	repo repository.MetricRepository
}

func NewMetricUseCase(repo repository.MetricRepository) *MetricUseCase {
	return &MetricUseCase{repo: repo}
}

func (uc *MetricUseCase) Create(ctx context.Context, m *models.Metric) (*models.Metric, error) {
	if m.IsDerived() {
		if err := uc.validateDerivedMetric(ctx, m); err != nil {
			return nil, err
		}
	}
	return uc.repo.Create(ctx, m)
}

// validateDerivedMetric checks that numerator and denominator metrics exist and that adding m would not create a cycle.
func (uc *MetricUseCase) validateDerivedMetric(ctx context.Context, m *models.Metric) error {
	numKey, denKey := *m.NumeratorMetricKey, *m.DenominatorMetricKey
	var selfErrs []models.FieldError
	if m.Key == numKey {
		selfErrs = append(selfErrs, models.FieldError{Field: "numeratorMetricKey", Issue: "must differ from metric key", RejectedValue: numKey})
	}
	if m.Key == denKey {
		selfErrs = append(selfErrs, models.FieldError{Field: "denominatorMetricKey", Issue: "must differ from metric key", RejectedValue: denKey})
	}
	if len(selfErrs) > 0 {
		return models.NewApiError(models.ErrCodeValidationFailed, "derived metric cannot use itself as numerator or denominator", nil, selfErrs, nil)
	}
	numMetric, err := uc.repo.GetByKey(ctx, numKey)
	if err != nil {
		return models.NewApiError(models.ErrCodeNotFound, fmt.Sprintf("numerator metric %q not found", numKey), nil, nil, err)
	}
	denMetric, err := uc.repo.GetByKey(ctx, denKey)
	if err != nil {
		return models.NewApiError(models.ErrCodeNotFound, fmt.Sprintf("denominator metric %q not found", denKey), nil, nil, err)
	}
	depsNum, err := uc.dependsOn(ctx, numMetric, map[string]struct{}{})
	if err != nil {
		return err
	}
	if _, ok := depsNum[m.Key]; ok {
		return models.NewApiError(
			models.ErrCodeValidationFailed,
			"cycle detected: numerator metric (or its dependencies) would depend on this metric",
			nil,
			[]models.FieldError{{Field: "numeratorMetricKey", Issue: "would create a cycle", RejectedValue: numKey}},
			nil,
		)
	}
	depsDen, err := uc.dependsOn(ctx, denMetric, map[string]struct{}{})
	if err != nil {
		return err
	}
	if _, ok := depsDen[m.Key]; ok {
		return models.NewApiError(
			models.ErrCodeValidationFailed,
			"cycle detected: denominator metric (or its dependencies) would depend on this metric",
			nil,
			[]models.FieldError{{Field: "denominatorMetricKey", Issue: "would create a cycle", RejectedValue: denKey}},
			nil,
		)
	}
	return nil
}

// dependsOn returns the set of metric keys that the given metric depends on (transitively). visited is used to detect cycles.
func (uc *MetricUseCase) dependsOn(ctx context.Context, metric *models.Metric, visited map[string]struct{}) (map[string]struct{}, error) {
	out := map[string]struct{}{metric.Key: {}}
	if !metric.IsDerived() {
		return out, nil
	}
	if _, ok := visited[metric.Key]; ok {
		return nil, models.NewApiError(models.ErrCodeValidationFailed, "cycle in metric dependencies", nil, nil, nil)
	}
	visited[metric.Key] = struct{}{}
	defer delete(visited, metric.Key)
	num, err := uc.repo.GetByKey(ctx, *metric.NumeratorMetricKey)
	if err != nil {
		return nil, err
	}
	den, err := uc.repo.GetByKey(ctx, *metric.DenominatorMetricKey)
	if err != nil {
		return nil, err
	}
	numDeps, err := uc.dependsOn(ctx, num, visited)
	if err != nil {
		return nil, err
	}
	denDeps, err := uc.dependsOn(ctx, den, visited)
	if err != nil {
		return nil, err
	}
	for k := range numDeps {
		out[k] = struct{}{}
	}
	for k := range denDeps {
		out[k] = struct{}{}
	}
	return out, nil
}

func (uc *MetricUseCase) GetByKey(ctx context.Context, key string) (*models.Metric, error) {
	return uc.repo.GetByKey(ctx, key)
}

func (uc *MetricUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Metric, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *MetricUseCase) List(ctx context.Context) ([]*models.Metric, error) {
	return uc.repo.List(ctx)
}
