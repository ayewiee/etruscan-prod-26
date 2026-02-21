package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ReportUseCase struct {
	experimentRepo repository.ExperimentRepository
	metricRepo     repository.MetricRepository
	computer       *MetricComputer
}

func NewReportUseCase(
	experimentRepo repository.ExperimentRepository,
	metricRepo repository.MetricRepository,
	computer *MetricComputer,
) *ReportUseCase {
	return &ReportUseCase{
		experimentRepo: experimentRepo,
		metricRepo:     metricRepo,
		computer:       computer,
	}
}

type ExperimentReport struct {
	ExperimentID uuid.UUID
	From         time.Time
	To           time.Time
	Variants     []VariantMetricValues
}

type VariantMetricValues struct {
	VariantID   uuid.UUID
	VariantName string
	Metrics     map[string]float64
}

func (uc *ReportUseCase) GetExperimentReport(ctx context.Context, experimentID uuid.UUID, from, to time.Time) (*ExperimentReport, error) {
	exp, err := uc.experimentRepo.GetByID(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	variants, err := uc.experimentRepo.ListVariantsByExperimentID(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants

	em, err := uc.metricRepo.ListExperimentMetrics(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	if len(em) == 0 {
		return &ExperimentReport{
			ExperimentID: experimentID,
			From:         from,
			To:           to,
			Variants:     buildEmptyVariants(variants),
		}, nil
	}

	metrics := make([]*models.Metric, len(em))
	for i := range em {
		m, err := uc.metricRepo.GetByID(ctx, em[i].MetricID)
		if err != nil {
			return nil, err
		}
		metrics[i] = m
	}

	variantValues := make([]VariantMetricValues, len(variants))
	for i, v := range variants {
		vm := VariantMetricValues{
			VariantID:   v.ID,
			VariantName: v.Name,
			Metrics:     make(map[string]float64),
		}
		for _, metric := range metrics {
			val, err := uc.computer.Compute(ctx, experimentID, &v.ID, metric, from, to)
			if err != nil {
				return nil, err
			}
			vm.Metrics[metric.Key] = val
		}
		variantValues[i] = vm
	}

	return &ExperimentReport{
		ExperimentID: experimentID,
		From:         from,
		To:           to,
		Variants:     variantValues,
	}, nil
}

func buildEmptyVariants(variants []*models.Variant) []VariantMetricValues {
	out := make([]VariantMetricValues, len(variants))
	for i, v := range variants {
		out[i] = VariantMetricValues{
			VariantID:   v.ID,
			VariantName: v.Name,
			Metrics:     make(map[string]float64),
		}
	}
	return out
}
