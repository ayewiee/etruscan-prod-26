package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	var eventTypeKey pgtype.Text
	var aggType dbgen.NullMetricAggregationType
	var numKey, denKey pgtype.Text
	if m.IsDerived() {
		numKey = database.ToPgText(m.NumeratorMetricKey)
		denKey = database.ToPgText(m.DenominatorMetricKey)
	} else {
		eventTypeKey = pgtype.Text{String: m.EventTypeKey, Valid: true}
		aggType = dbgen.NullMetricAggregationType{MetricAggregationType: dbgen.MetricAggregationType(m.AggregationType), Valid: true}
	}
	row, err := r.db.CreateMetric(ctx, dbgen.CreateMetricParams{
		Key:                  m.Key,
		Name:                 m.Name,
		Description:          database.ToPgText(m.Description),
		Type:                 dbgen.MetricType(m.Type),
		EventTypeKey:         eventTypeKey,
		AggregationType:      aggType,
		IsGuardrail:          m.IsGuardrail,
		NumeratorMetricKey:   numKey,
		DenominatorMetricKey: denKey,
	})
	if err != nil {
		return nil, err
	}
	return metricRowToDomain(row.ID, row.Key, row.Name, row.Description, row.Type, row.EventTypeKey, row.AggregationType, row.IsGuardrail, row.CreatedAt, row.NumeratorMetricKey, row.DenominatorMetricKey), nil
}

func (r *SQLCMetricRepository) GetByKey(ctx context.Context, key string) (*models.Metric, error) {
	row, err := r.db.GetMetricByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return metricRowToDomain(row.ID, row.Key, row.Name, row.Description, row.Type, row.EventTypeKey, row.AggregationType, row.IsGuardrail, row.CreatedAt, row.NumeratorMetricKey, row.DenominatorMetricKey), nil
}

func (r *SQLCMetricRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Metric, error) {
	row, err := r.db.GetMetricByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return metricRowToDomain(row.ID, row.Key, row.Name, row.Description, row.Type, row.EventTypeKey, row.AggregationType, row.IsGuardrail, row.CreatedAt, row.NumeratorMetricKey, row.DenominatorMetricKey), nil
}

func (r *SQLCMetricRepository) List(ctx context.Context) ([]*models.Metric, error) {
	rows, err := r.db.ListMetrics(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*models.Metric, len(rows))
	for i := range rows {
		out[i] = metricRowToDomain(rows[i].ID, rows[i].Key, rows[i].Name, rows[i].Description, rows[i].Type, rows[i].EventTypeKey, rows[i].AggregationType, rows[i].IsGuardrail, rows[i].CreatedAt, rows[i].NumeratorMetricKey, rows[i].DenominatorMetricKey)
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

func metricRowToDomain(id uuid.UUID, key, name string, desc pgtype.Text, t dbgen.MetricType, eventKey pgtype.Text, agg dbgen.NullMetricAggregationType, isGuardrail bool, createdAt pgtype.Timestamptz, numKey, denKey pgtype.Text) *models.Metric {
	var eventTypeKey string
	if eventKey.Valid {
		eventTypeKey = eventKey.String
	}
	var aggType models.MetricAggregationType
	if agg.Valid {
		aggType = models.MetricAggregationType(agg.MetricAggregationType)
	}
	return &models.Metric{
		ID:                   id,
		Key:                  key,
		Name:                 name,
		Type:                 models.MetricType(t),
		EventTypeKey:         eventTypeKey,
		AggregationType:      aggType,
		NumeratorMetricKey:   database.FromPgText(numKey),
		DenominatorMetricKey: database.FromPgText(denKey),
		IsGuardrail:          isGuardrail,
		Description:          database.FromPgText(desc),
		CreatedAt:            database.FromPgTimestamptz(createdAt),
	}
}
