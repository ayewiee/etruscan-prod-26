package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type ExperimentListFilters struct {
	FlagID    *uuid.UUID
	CreatedBy *uuid.UUID
	Status    *models.ExperimentStatus
	Outcome   *models.ExperimentOutcome
	Limit     int
	Offset    int
}

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error)
	Update(ctx context.Context, experiment *models.Experiment) (*models.Experiment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Experiment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.ExperimentStatus) error
	List(ctx context.Context, filters ExperimentListFilters) ([]*models.Experiment, int, error)
	ListStatusChanges(ctx context.Context, experimentID uuid.UUID) ([]*models.ExperimentStatusChange, error)

	SaveExperimentSnapshot(ctx context.Context, snapshot *models.ExperimentSnapshot) error
	ListExperimentSnapshots(ctx context.Context, experimentID uuid.UUID) ([]*models.ExperimentSnapshot, error)

	GetActiveExperimentByFlagKey(ctx context.Context, flagID string) (*models.Experiment, error)
	GetRunningExperimentByFlagKey(ctx context.Context, flagID string) (*models.Experiment, error)

	CreateVariants(ctx context.Context, experimentID uuid.UUID, variants []*models.Variant) error
	ListVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) ([]*models.Variant, error)
	DeleteVariantsByExperimentID(ctx context.Context, experimentID uuid.UUID) error

	CreateExperimentReview(ctx context.Context, review *models.ExperimentReview) error
	ListExperimentReviews(ctx context.Context, id uuid.UUID) ([]*models.ExperimentReview, error)
	CountApprovals(ctx context.Context, experimentID uuid.UUID) (int, error)
	ClearExperimentReviews(ctx context.Context, experimentID uuid.UUID) error

	LogExperimentStatusChange(ctx context.Context, expStatusChange *models.ExperimentStatusChange) error

	Finish(ctx context.Context, id uuid.UUID, outcome models.ExperimentOutcome, outcomeComment string, setBy uuid.UUID) (*models.Experiment, error)
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
) error {
	return r.db.UpdateExperimentStatus(ctx, dbgen.UpdateExperimentStatusParams{
		ID:     id,
		Status: dbgen.ExperimentStatus(status),
	})
}

func (r SQLCExperimentRepository) List(
	ctx context.Context,
	filters ExperimentListFilters,
) ([]*models.Experiment, int, error) {
	total, err := r.db.CountExperiments(ctx, dbgen.CountExperimentsParams{
		CreatedBy: database.ToPgUUID(filters.CreatedBy),
		Status:    database.ToNullExperimentStatus(filters.Status),
		Outcome:   database.ToNullExperimentOutcome(filters.Outcome),
		FlagID:    database.ToPgUUID(filters.FlagID),
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.ListExperiments(ctx, dbgen.ListExperimentsParams{
		CreatedBy: database.ToPgUUID(filters.CreatedBy),
		Status:    database.ToNullExperimentStatus(filters.Status),
		Outcome:   database.ToNullExperimentOutcome(filters.Outcome),
		FlagID:    database.ToPgUUID(filters.FlagID),
		Limit:     int32(filters.Limit),
		Offset:    int32(filters.Offset),
	})
	if err != nil {
		return nil, 0, err
	}

	experiments := make([]*models.Experiment, len(rows))
	for i, row := range rows {
		experiments[i] = experimentRowToDomain(row)
	}

	return experiments, int(total), nil
}

func (r SQLCExperimentRepository) SaveExperimentSnapshot(
	ctx context.Context,
	snapshot *models.ExperimentSnapshot,
) error {
	jsonData, err := snapshot.Data.ToJSON()
	if err != nil {
		return err
	}

	err = r.db.CreateExperimentSnapshot(ctx, dbgen.CreateExperimentSnapshotParams{
		ExperimentID: snapshot.ExperimentID,
		Version:      int32(snapshot.Version),
		Data:         jsonData,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r SQLCExperimentRepository) ListExperimentSnapshots(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]*models.ExperimentSnapshot, error) {
	rows, err := r.db.GetExperimentSnapshots(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	snapshots := make([]*models.ExperimentSnapshot, len(rows))
	for i, row := range rows {
		data := models.ExperimentSnapshotData{}
		err = data.LoadFromJSON(row.Data)
		if err != nil {
			return nil, err
		}
		snapshots[i] = &models.ExperimentSnapshot{
			ID:           row.ID,
			ExperimentID: row.ExperimentID,
			Version:      int(row.Version),
			Data:         &data,
			CreatedAt:    row.CreatedAt.Time,
		}
	}

	return snapshots, nil
}

func (r SQLCExperimentRepository) ListStatusChanges(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]*models.ExperimentStatusChange, error) {
	rows, err := r.db.ListExperimentStatusChanges(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	experimentStatusChanges := make([]*models.ExperimentStatusChange, len(rows))
	for i, row := range rows {
		experimentStatusChanges[i] = &models.ExperimentStatusChange{
			ID:           row.ID,
			ExperimentID: row.ExperimentID,
			ActorID:      database.FromPgUUID(row.ActorID),
			From:         database.FromNullExperimentStatus(row.FromStatus),
			To:           models.ExperimentStatus(row.ToStatus),
			Comment:      database.FromPgText(row.Comment),
			CreatedAt:    row.CreatedAt.Time,
		}
	}

	return experimentStatusChanges, nil
}

func (r SQLCExperimentRepository) GetActiveExperimentByFlagKey(
	ctx context.Context,
	flagKey string,
) (*models.Experiment, error) {
	exp, err := r.db.GetActiveExperimentByFlagKey(ctx, flagKey)
	if err != nil {
		return nil, err
	}
	return experimentRowToDomain(exp), nil
}

func (r SQLCExperimentRepository) GetRunningExperimentByFlagKey(
	ctx context.Context,
	flagKey string,
) (*models.Experiment, error) {
	exp, err := r.db.GetRunningExperimentByFlagKey(ctx, flagKey)
	if err != nil {
		return nil, err
	}
	return experimentRowToDomain(exp), nil
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
		Version:            int(experiment.Version),
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

func (r SQLCExperimentRepository) CreateExperimentReview(
	ctx context.Context,
	review *models.ExperimentReview,
) error {
	return r.db.CreateExperimentReview(ctx, dbgen.CreateExperimentReviewParams{
		ExperimentID: review.ExperimentID,
		ApproverID:   review.ApproverID,
		Decision:     dbgen.ExperimentReviewDecision(review.Decision),
		Comment:      database.ToPgText(review.Comment),
	})
}

func (r SQLCExperimentRepository) ListExperimentReviews(
	ctx context.Context,
	id uuid.UUID,
) ([]*models.ExperimentReview, error) {
	rows, err := r.db.ListExperimentReviews(ctx, id)
	if err != nil {
		return nil, err
	}

	reviews := make([]*models.ExperimentReview, len(rows))
	for i, row := range rows {
		reviews[i] = &models.ExperimentReview{
			ID:           row.ID,
			ExperimentID: row.ExperimentID,
			ApproverID:   row.ApproverID,
			Decision:     models.ExperimentReviewDecision(row.Decision),
			Comment:      database.FromPgText(row.Comment),
			CreatedAt:    row.CreatedAt.Time,
		}
	}

	return reviews, nil
}

func (r SQLCExperimentRepository) CountApprovals(ctx context.Context, experimentID uuid.UUID) (int, error) {
	count, err := r.db.CountApprovals(ctx, experimentID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r SQLCExperimentRepository) ClearExperimentReviews(ctx context.Context, experimentID uuid.UUID) error {
	return r.db.ClearExperimentReviews(ctx, experimentID)
}

func (r SQLCExperimentRepository) LogExperimentStatusChange(
	ctx context.Context,
	expStatusChange *models.ExperimentStatusChange,
) error {
	return r.db.LogExperimentStatusChange(ctx, dbgen.LogExperimentStatusChangeParams{
		ExperimentID: expStatusChange.ExperimentID,
		ActorID:      database.ToPgUUID(expStatusChange.ActorID),
		FromStatus:   database.ToNullExperimentStatus(expStatusChange.From),
		ToStatus:     dbgen.ExperimentStatus(expStatusChange.To),
		Comment:      database.ToPgText(expStatusChange.Comment),
	})
}

func (r SQLCExperimentRepository) Finish(ctx context.Context, id uuid.UUID, outcome models.ExperimentOutcome, outcomeComment string, setBy uuid.UUID) (*models.Experiment, error) {
	row, err := r.db.FinishExperiment(ctx, dbgen.FinishExperimentParams{
		ID:             id,
		Status:         dbgen.ExperimentStatusFINISHED,
		Outcome:        database.ToNullExperimentOutcome(&outcome),
		OutcomeComment: database.ToPgText(&outcomeComment),
		OutcomeSetBy:   database.ToPgUUID(&setBy),
	})
	if err != nil {
		return nil, err
	}

	return experimentRowToDomain(row), nil
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
