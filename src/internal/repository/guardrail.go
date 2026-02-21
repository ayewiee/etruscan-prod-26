package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type GuardrailRepository interface {
	Create(
		ctx context.Context,
		experimentID, metricID uuid.UUID,
		threshold float64,
		direction, action string,
		windowSeconds int,
	) (*models.Guardrail, error)
	ListByExperimentID(ctx context.Context, experimentID uuid.UUID) ([]*models.Guardrail, error)
	DeleteByExperimentID(ctx context.Context, experimentID uuid.UUID) error
	CreateTrigger(
		ctx context.Context,
		guardrailID, experimentID uuid.UUID,
		metricValue float64,
		metricKey string,
		threshold float64,
		windowSeconds int,
		action string,
	) (uuid.UUID, error)
}

type SQLCGuardrailRepository struct {
	db     *dbgen.Queries
	metric MetricRepository
}

func NewSQLCGuardrailRepository(db *dbgen.Queries, metric MetricRepository) GuardrailRepository {
	return &SQLCGuardrailRepository{db: db, metric: metric}
}

func (r *SQLCGuardrailRepository) Create(
	ctx context.Context,
	experimentID, metricID uuid.UUID,
	threshold float64,
	direction, action string,
	windowSeconds int,
) (*models.Guardrail, error) {
	row, err := r.db.CreateGuardrail(ctx, dbgen.CreateGuardrailParams{
		ExperimentID:       experimentID,
		MetricID:           metricID,
		Threshold:          threshold,
		ThresholdDirection: direction,
		Action:             action,
		WindowSeconds:      int32(windowSeconds),
	})
	if err != nil {
		return nil, err
	}

	m, _ := r.metric.GetByID(ctx, row.MetricID)
	key := ""
	if m != nil {
		key = m.Key
	}

	return &models.Guardrail{
		ID:                 row.ID,
		ExperimentID:       row.ExperimentID,
		MetricID:           row.MetricID,
		MetricKey:          key,
		Threshold:          row.Threshold,
		ThresholdDirection: row.ThresholdDirection,
		Action:             row.Action,
		WindowSeconds:      int(row.WindowSeconds),
	}, nil
}

func (r *SQLCGuardrailRepository) ListByExperimentID(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]*models.Guardrail, error) {
	rows, err := r.db.ListGuardrailsByExperimentID(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.Guardrail, len(rows))
	for i := range rows {
		m, _ := r.metric.GetByID(ctx, rows[i].MetricID)
		key := ""
		if m != nil {
			key = m.Key
		}

		out[i] = &models.Guardrail{
			ID:                 rows[i].ID,
			ExperimentID:       rows[i].ExperimentID,
			MetricID:           rows[i].MetricID,
			MetricKey:          key,
			Threshold:          rows[i].Threshold,
			ThresholdDirection: rows[i].ThresholdDirection,
			Action:             rows[i].Action,
			WindowSeconds:      int(rows[i].WindowSeconds),
		}
	}

	return out, nil
}

func (r *SQLCGuardrailRepository) DeleteByExperimentID(ctx context.Context, experimentID uuid.UUID) error {
	return r.db.DeleteGuardrailsByExperimentID(ctx, experimentID)
}

func (r *SQLCGuardrailRepository) CreateTrigger(
	ctx context.Context,
	guardrailID, experimentID uuid.UUID,
	metricValue float64,
	metricKey string,
	threshold float64,
	windowSeconds int,
	action string,
) (uuid.UUID, error) {
	return r.db.CreateGuardrailTrigger(ctx, dbgen.CreateGuardrailTriggerParams{
		GuardrailID:    guardrailID,
		ExperimentID:   experimentID,
		MetricValue:    metricValue,
		MetricKey:      database.ToPgText(&metricKey),
		ThresholdValue: pgtype.Float8{Float64: threshold, Valid: true},
		WindowSeconds:  pgtype.Int4{Int32: int32(windowSeconds), Valid: true},
		Action:         database.ToPgText(&action),
	})
}
