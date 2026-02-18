package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type ListFilters struct {
	FlagID    *uuid.UUID
	CreatedBy *uuid.UUID
	Status    *models.ExperimentStatus
	Outcome   *models.ExperimentOutcome
}

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error)
	Update(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.ExperimentStatus) (*models.Experiment, error)
	List(ctx context.Context, filters ListFilters) ([]*models.Experiment, error)

	CreateVariants(ctx context.Context, experimentID uuid.UUID, variants []*models.Variant) error
	ListVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) ([]*models.Variant, error)
	DeleteVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) error
}

type SQLCExperimentRepository struct {
	db *dbgen.Queries
}

func NewSQLCExperimentRepository(db *dbgen.Queries) ExperimentRepository {
	return &SQLCExperimentRepository{db: db}
}

func (r SQLCExperimentRepository) Create(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error) {
	row, err := r.db.CreateExperiment(ctx, dbgen.CreateExperimentParams{
		FlagID:        experiment.FlagID,
		Name:          experiment.Name,
		Description:   database.ToPgText(experiment.Description),
		CreatedBy:     experiment.CreatedBy,
		Status:        dbgen.ExperimentStatus(experiment.Status),
		AudiencePct:   int32(experiment.AudiencePercentage),
		TargetingRule: database.ToPgText(experiment.TargetingRule),
	})
	if err != nil {
		return nil, err
	}

	return experimentRowToDomain(row), nil
}

func (r SQLCExperimentRepository) Update(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error) {
	row, err := r.db.UpdateExperiment(ctx, dbgen.UpdateExperimentParams{
		ID:            experiment.ID,
		Name:          experiment.Name,
		Description:   database.ToPgText(experiment.Description),
		AudiencePct:   int32(experiment.AudiencePercentage),
		TargetingRule: database.ToPgText(experiment.TargetingRule),
	})
	if err != nil {
		return nil, err
	}

	return experimentRowToDomain(row), nil
}

func (r SQLCExperimentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error) {
	row, err := r.db.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return experimentRowToDomain(row), nil
}

func (r SQLCExperimentRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status models.ExperimentStatus,
) (*models.Experiment, error) {
	row, err := r.db.UpdateExperimentStatus(ctx, dbgen.UpdateExperimentStatusParams{
		ID:     id,
		Status: dbgen.ExperimentStatus(status),
	})
	if err != nil {
		return nil, err
	}

	return experimentRowToDomain(row), nil
}

func (r SQLCExperimentRepository) List(ctx context.Context, filters ListFilters) ([]*models.Experiment, error) {
	rows, err := r.db.ListExperiments(ctx, dbgen.ListExperimentsParams{
		CreatedBy: database.ToPgUUID(filters.CreatedBy),
		Status:    database.ToNullExperimentStatus(filters.Status),
		Outcome:   database.ToNullExperimentOutcome(filters.Outcome),
		FlagID:    database.ToPgUUID(filters.FlagID),
	})
	if err != nil {
		return nil, err
	}

	experiments := make([]*models.Experiment, len(rows))
	for i, row := range rows {
		experiments[i] = experimentRowToDomain(row)
	}

	return experiments, nil
}

func (r SQLCExperimentRepository) CreateVariants(
	ctx context.Context,
	experimentID uuid.UUID,
	variants []*models.Variant,
) error {
	paramsList := make([]dbgen.BatchCreateVariantsParams, len(variants))
	for i, variant := range variants {
		paramsList[i] = dbgen.BatchCreateVariantsParams{
			ExperimentID: experimentID,
			Name:         variant.Name,
			Value:        variant.Value,
			Weight:       int32(variant.Weight),
			IsControl:    variant.IsControl,
		}
	}

	rowsCreated, err := r.db.BatchCreateVariants(ctx, paramsList)
	if err != nil {
		return err
	}
	if int(rowsCreated) != len(variants) {
		return database.BatchOperationError
	}
	return nil
}

func (r SQLCExperimentRepository) ListVariantsByExperimentID(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]*models.Variant, error) {
	rows, err := r.db.ListVariantsByExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	return variantRowsToDomain(rows), nil
}

func (r SQLCExperimentRepository) DeleteVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) error {
	return r.db.DeleteVariantsByExperiment(ctx, experimentID)
}

func experimentRowToDomain(experiment dbgen.Experiment) *models.Experiment {
	return &models.Experiment{
		ID:                 experiment.ID,
		FlagID:             experiment.FlagID,
		Name:               experiment.Name,
		Description:        database.FromPgText(experiment.Description),
		CreatedBy:          experiment.CreatedBy,
		Status:             models.ExperimentStatus(experiment.Status),
		AudiencePercentage: int(experiment.AudiencePct),
		TargetingRule:      database.FromPgText(experiment.TargetingRule),
		Outcome:            database.FromNullExperimentOutcome(experiment.Outcome),
		OutcomeComment:     database.FromPgText(experiment.OutcomeComment),
		OutcomeSetAt:       database.FromPgTimestamptz(experiment.OutcomeSetAt), // 'cause it's nullable
		OutcomeSetBy:       database.FromPgUUID(experiment.OutcomeSetBy),
		CreatedAt:          experiment.CreatedAt.Time,
		UpdatedAt:          experiment.UpdatedAt.Time,
		Variants:           nil,
	}
}

func variantRowsToDomain(rows []dbgen.Variant) []*models.Variant {
	variants := make([]*models.Variant, len(rows))
	for i, row := range rows {
		variants[i] = &models.Variant{
			ID:           row.ID,
			ExperimentID: row.ExperimentID,
			Name:         row.Name,
			Value:        row.Value,
			Weight:       int(row.Weight),
			IsControl:    row.IsControl,
		}
	}
	return variants
}
