package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
)

type EventTypeUseCase struct {
	repo repository.EventTypeRepository
}

func NewEventTypeUseCase(repo repository.EventTypeRepository) *EventTypeUseCase {
	return &EventTypeUseCase{repo: repo}
}

func (uc *EventTypeUseCase) Create(
	ctx context.Context,
	actor models.UserAuthData,
	eventType *models.EventType,
) (*models.EventType, error) {
	if !actor.Role.CanManageEventTypes() {
		return nil, models.ErrForbidden
	}
	if eventType.RequiresKey != nil {
		requiredEventType, err := uc.repo.GetByKey(ctx, *eventType.RequiresKey)
		if err != nil {
			return nil, models.NewErrNotFound(
				"Event type that was specified as required not found",
				map[string]interface{}{"key": *eventType.RequiresKey},
				nil,
			)
		}
		eventType.RequiresID = requiredEventType.RequiresID
	}

	return uc.repo.Create(ctx, eventType)
}

func (uc *EventTypeUseCase) List(ctx context.Context) ([]*models.EventType, error) {
	return uc.repo.List(ctx)
}
