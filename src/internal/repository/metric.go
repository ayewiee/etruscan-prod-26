package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type MetricRepository interface {
	Create(ctx context.Context, m *models.Metric) (*models.Metric, error)
	GetByKey(ctx context.Context, key string) (*models.Metric, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Metric, error)
	List(ctx context.Context) ([]*models.Metric, error)
	DeleteExperimentMetrics(ctx context.Context, experimentID uuid.UUID) error
	AddExperimentMetric(ctx context.Context, experimentID, metricID uuid.UUID, isPrimary bool) error
	ListExperimentMetrics(ctx context.Context, experimentID uuid.UUID) ([]*models.ExperimentMetric, error)
}

type SQLCMetricRepository struct {
	db *dbgen.Queries
}

func NewSQLCMetricRepository(db *dbgen.Queries) MetricRepository {
	return &SQLCMetricRepository{db: db}
}

func (r *SQLCMetricRepository) Create(ctx context.Context, m *models.Metric) (*models.Metric, error) {
	row, err := r.db.CreateMetric(ctx, dbgen.CreateMetricParams{
		Key:             m.Key,
		Name:            m.Name,
		Description:     database.ToPgText(m.Description),
		Type:            dbgen.MetricType(m.Type),
		EventTypeKey:    m.EventTypeKey,
		AggregationType: dbgen.MetricAggregationType(m.AggregationType),
		IsGuardrail:     m.IsGuardrail,
	})
	if err != nil {
		return nil, err
	}

	return metricFromDB(row), nil
}

func (r *SQLCMetricRepository) GetByKey(ctx context.Context, key string) (*models.Metric, error) {
	row, err := r.db.GetMetricByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return metricFromDB(row), nil
}

func (r *SQLCMetricRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Metric, error) {
	row, err := r.db.GetMetricByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return metricFromDB(row), nil
}

func (r *SQLCMetricRepository) List(ctx context.Context) ([]*models.Metric, error) {
	rows, err := r.db.ListMetrics(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*models.Metric, len(rows))
	for i := range rows {
		out[i] = metricFromDB(rows[i])
	}

	return out, nil
}

func (r *SQLCMetricRepository) DeleteExperimentMetrics(ctx context.Context, experimentID uuid.UUID) error {
	return r.db.DeleteExperimentMetrics(ctx, experimentID)
}

func (r *SQLCMetricRepository) AddExperimentMetric(ctx context.Context, experimentID, metricID uuid.UUID, isPrimary bool) error {
	return r.db.AddExperimentMetric(ctx, dbgen.AddExperimentMetricParams{
		ExperimentID: experimentID,
		MetricID:     metricID,
		IsPrimary:    isPrimary,
	})
}

func (r *SQLCMetricRepository) ListExperimentMetrics(ctx context.Context, experimentID uuid.UUID) ([]*models.ExperimentMetric, error) {
	rows, err := r.db.ListExperimentMetricIDs(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.ExperimentMetric, len(rows))
	for i := range rows {
		metric, err := r.db.GetMetricByID(ctx, rows[i].MetricID)
		if err != nil {
			return nil, err
		}

		out[i] = &models.ExperimentMetric{
			ExperimentID: experimentID,
			MetricID:     rows[i].MetricID,
			MetricKey:    metric.Key,
			IsPrimary:    rows[i].IsPrimary,
		}
	}

	return out, nil
}

func metricFromDB(row dbgen.Metric) *models.Metric {
	return &models.Metric{
		ID:              row.ID,
		Key:             row.Key,
		Name:            row.Name,
		Type:            models.MetricType(row.Type),
		EventTypeKey:    row.EventTypeKey,
		AggregationType: models.MetricAggregationType(row.AggregationType),
		IsGuardrail:     row.IsGuardrail,
		Description:     database.FromPgText(row.Description),
		CreatedAt:       database.FromPgTimestamptz(row.CreatedAt),
	}
}
