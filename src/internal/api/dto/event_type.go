package dto

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CreateEventTypeRequest struct {
	Key         string  `json:"key" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty,required"`
	Requires    *string `json:"requires" validate:"omitempty,required"`
}

type EventTypeResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Requires    string    `json:"requires,omitempty"`
	CreatedAt   string    `json:"createdAt"`
}

func EventTypeResponseFromDomain(et *models.EventType) *EventTypeResponse {
	var requires string
	if et.RequiresKey != nil {
		requires = *et.RequiresKey
	}
	return &EventTypeResponse{
		ID:          et.ID,
		Key:         et.Key,
		Name:        et.Name,
		Description: et.Description,
		Requires:    requires,
		CreatedAt:   et.CreatedAt.Format(time.RFC3339),
	}
}

func EventTypeResponseListFromDomain(ets []*models.EventType) []*EventTypeResponse {
	responses := make([]*EventTypeResponse, len(ets))
	for i, et := range ets {
		responses[i] = EventTypeResponseFromDomain(et)
	}
	return responses
}
