package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/google/uuid"
)

type NotificationSettingsUseCase struct {
	repo repository.NotificationSettingsRepository
}

func NewNotificationSettingsUseCase(repo repository.NotificationSettingsRepository) *NotificationSettingsUseCase {
	return &NotificationSettingsUseCase{repo: repo}
}

func (uc *NotificationSettingsUseCase) Create(ctx context.Context, settings *models.NotificationSettings) error {
	return uc.repo.Create(ctx, settings)
}

func (uc *NotificationSettingsUseCase) List(ctx context.Context, userID uuid.UUID) ([]*models.NotificationSettings, error) {
	return uc.repo.ListForUser(ctx, userID)
}

func (uc *NotificationSettingsUseCase) DeleteForExperimentAndUser(
	ctx context.Context,
	actor models.UserAuthData,
	expID uuid.UUID,
) error {
	settings, err := uc.repo.GetForExperimentAndUser(ctx, expID, actor.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.NewErrNotFound("Settings not found", nil, err)
		}
	}

	return uc.repo.DeleteByID(ctx, settings.ID)
}
