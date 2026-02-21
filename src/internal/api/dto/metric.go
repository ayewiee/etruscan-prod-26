package dto

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CreateMetricRequest struct {
	Key             string  `json:"key" validate:"required"`
	Name            string  `json:"name" validate:"required"`
	Description     *string `json:"description"`
	Type            string  `json:"type" validate:"required,oneof=binomial continuous"`
	EventTypeKey    string  `json:"eventTypeKey" validate:"required"`
	AggregationType string  `json:"aggregationType" validate:"required,oneof=count sum avg p95"`
	IsGuardrail     bool    `json:"isGuardrail"`
}

type MetricResponse struct {
	ID              uuid.UUID `json:"id"`
	Key             string    `json:"key"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	Type            string    `json:"type"`
	EventTypeKey    string    `json:"eventTypeKey"`
	AggregationType string    `json:"aggregationType"`
	IsGuardrail     bool      `json:"isGuardrail"`
	CreatedAt       string    `json:"createdAt"`
}

func MetricResponseFromDomain(m *models.Metric) *MetricResponse {
	var createdAt string
	if m.CreatedAt != nil {
		createdAt = m.CreatedAt.Format(time.RFC3339)
	}
	return &MetricResponse{
		ID:              m.ID,
		Key:             m.Key,
		Name:            m.Name,
		Description:     m.Description,
		Type:            string(m.Type),
		EventTypeKey:    m.EventTypeKey,
		AggregationType: string(m.AggregationType),
		IsGuardrail:     m.IsGuardrail,
		CreatedAt:       createdAt,
	}
}
