package repository

import (
	"context"
	"encoding/json"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DecisionRepository interface {
	Create(ctx context.Context, decision *models.Decision) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Decision, error)
	ListDecisionIDsByExperimentVariantWindow(ctx context.Context, experimentID uuid.UUID, variantID *uuid.UUID, from, to time.Time) ([]uuid.UUID, error)
}

type SQLCDecisionRepository struct {
	db *dbgen.Queries
}

func NewSQLCDecisionRepository(db *dbgen.Queries) DecisionRepository {
	return &SQLCDecisionRepository{db: db}
}

func (r *SQLCDecisionRepository) Create(ctx context.Context, decision *models.Decision) (uuid.UUID, error) {
	contextBytes, err := json.Marshal(decision.Context)
	if err != nil {
		return uuid.Nil, err
	}

	jsonValue, err := json.Marshal(decision)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := r.db.CreateDecision(ctx, dbgen.CreateDecisionParams{
		ExperimentID: database.ToPgUUID(decision.ExperimentID),
		VariantID:    database.ToPgUUID(decision.VariantID),
		FlagKey:      decision.FlagKey,
		Value:        jsonValue,
		UserID:       decision.UserID,
		Context:      contextBytes,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *SQLCDecisionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Decision, error) {
	row, err := r.db.GetDecisionById(ctx, id)
	if err != nil {
		return nil, err
	}

	contextMap := make(map[string]interface{})
	err = json.Unmarshal(row.Context, &contextMap)
	if err != nil {
		return nil, err
	}

	return &models.Decision{
		ID:           row.ID,
		ExperimentID: database.FromPgUUID(row.ExperimentID),
		VariantID:    database.FromPgUUID(row.VariantID),
		FlagKey:      row.FlagKey,
		Value:        row.Value,
		UserID:       row.UserID,
		Context:      contextMap,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (r *SQLCDecisionRepository) ListDecisionIDsByExperimentVariantWindow(
	ctx context.Context,
	experimentID uuid.UUID,
	variantID *uuid.UUID,
	from, to time.Time,
) ([]uuid.UUID, error) {
	fromPg := pgtype.Timestamptz{Time: from, Valid: true}
	toPg := pgtype.Timestamptz{Time: to, Valid: true}

	return r.db.ListDecisionIDsByExperimentVariantWindow(ctx, dbgen.ListDecisionIDsByExperimentVariantWindowParams{
		ExperimentID: database.ToPgUUID(&experimentID),
		CreatedAt:    fromPg,
		CreatedAt_2:  toPg,
		VariantID:    database.ToPgUUID(variantID),
	})
}
