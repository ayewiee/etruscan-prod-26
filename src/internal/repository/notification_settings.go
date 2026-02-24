package repository

import (
	"context"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type NotificationSettingsRepository interface {
	Create(ctx context.Context, settings *models.NotificationSettings) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.NotificationSettings, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]*models.NotificationSettings, error)
	ListForExperiment(ctx context.Context, experimentID uuid.UUID) ([]*models.NotificationSettings, error)
	GetForExperimentAndUser(ctx context.Context, experimentID, userID uuid.UUID) (*models.NotificationSettings, error)
	DeleteByID(ctx context.Context, settingsID uuid.UUID) error
}

type SQLCNotificationSettingsRepository struct {
	db *dbgen.Queries
}

func NewSQLCNotificationSettingsRepository(db *dbgen.Queries) *SQLCNotificationSettingsRepository {
	return &SQLCNotificationSettingsRepository{db: db}
}

func (r *SQLCNotificationSettingsRepository) Create(ctx context.Context, settings *models.NotificationSettings) error {
	return r.db.CreateNotificationSettings(ctx, dbgen.CreateNotificationSettingsParams{
		UserID:         settings.UserID,
		ExperimentID:   settings.ExperimentID,
		Severity:       dbgen.NotificationSeverity(settings.Severity),
		EnableTelegram: settings.EnableTelegram,
		EnableEmail:    settings.EnableEmail,
	})
}

func (r *SQLCNotificationSettingsRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.NotificationSettings, error) {
	row, err := r.db.GetNotificationSettingsByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return notificationSettingsRowToDomain(row), nil
}

func (r *SQLCNotificationSettingsRepository) ListForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*models.NotificationSettings, error) {
	rows, err := r.db.ListNotificationSettingsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	notificationSettings := make([]*models.NotificationSettings, len(rows))
	for i, row := range rows {
		notificationSettings[i] = notificationSettingsRowToDomain(row)
	}

	return notificationSettings, nil
}

func (r *SQLCNotificationSettingsRepository) ListForExperiment(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]*models.NotificationSettings, error) {
	rows, err := r.db.ListNotificationSettingsForExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	notificationSettings := make([]*models.NotificationSettings, len(rows))
	for i, row := range rows {
		notificationSettings[i] = notificationSettingsRowToDomain(row)
	}

	return notificationSettings, nil
}

func (r *SQLCNotificationSettingsRepository) GetForExperimentAndUser(
	ctx context.Context,
	experimentID, userID uuid.UUID,
) (*models.NotificationSettings, error) {
	row, err := r.db.GetNotificationSettingsForExperimentAndUser(
		ctx,
		dbgen.GetNotificationSettingsForExperimentAndUserParams{
			ExperimentID: experimentID,
			UserID:       userID,
		},
	)
	if err != nil {
		return nil, err
	}

	return notificationSettingsRowToDomain(row), nil
}

func (r *SQLCNotificationSettingsRepository) DeleteByID(ctx context.Context, settingsID uuid.UUID) error {
	return r.db.DeleteNotificationSettingsByID(ctx, settingsID)
}

func notificationSettingsRowToDomain(row dbgen.NotificationSetting) *models.NotificationSettings {
	return &models.NotificationSettings{
		ID:             row.ID,
		UserID:         row.UserID,
		ExperimentID:   row.ExperimentID,
		Severity:       models.NotificationSeverity(row.Severity),
		EnableTelegram: row.EnableTelegram,
		EnableEmail:    row.EnableEmail,
	}
}
