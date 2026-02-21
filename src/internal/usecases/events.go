package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventsUseCase struct {
	eventRepo     repository.EventRepository
	eventTypeRepo repository.EventTypeRepository
}

func NewEventsUseCase(eventRepo repository.EventRepository, eventTypeRepo repository.EventTypeRepository) *EventsUseCase {
	return &EventsUseCase{eventRepo: eventRepo, eventTypeRepo: eventTypeRepo}
}

func (uc *EventsUseCase) Ingest(ctx context.Context, items []models.BatchEventItem) (*models.BatchEventsResult, error) {
	result := &models.BatchEventsResult{Errors: make([]models.BatchEventErr, 0)}

	for i, item := range items {
		// B4-2: required fields
		if item.EventID == "" {
			result.Rejected++
			result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "eventId is required"})
			continue
		}
		if item.EventTypeKey == "" {
			result.Rejected++
			result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "eventTypeKey is required"})
			continue
		}
		if item.UserID == "" {
			result.Rejected++
			result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "userId is required"})
			continue
		}
		if item.Timestamp.IsZero() {
			result.Rejected++
			result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "timestamp is required"})
			continue
		}

		// event type must exist
		_, err := uc.eventTypeRepo.GetByKey(ctx, item.EventTypeKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result.Rejected++
				result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "unknown eventTypeKey: " + item.EventTypeKey})
				continue
			}
			return nil, err
		}

		// dedup by (event_type_key, client_event_id)
		exists, err := uc.eventRepo.ExistsByTypeAndClientID(ctx, item.EventTypeKey, item.EventID)
		if err != nil {
			return nil, err
		}
		if exists {
			result.Duplicates++
			continue
		}

		props := item.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		propsBytes, err := json.Marshal(props)
		if err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, models.BatchEventErr{Index: i, Message: "invalid properties: " + err.Error()})
			continue
		}

		ts := pgtype.Timestamptz{Time: item.Timestamp, Valid: true}

		_, err = uc.eventRepo.Create(ctx, repository.CreateEventParams{
			EventTypeKey:  item.EventTypeKey,
			DecisionID:    &item.DecisionID,
			UserID:        item.UserID,
			Properties:    propsBytes,
			Timestamp:     ts,
			ClientEventID: item.EventID,
		})
		if err != nil {
			return nil, err
		}
		result.Accepted++
	}

	return result, nil
}
