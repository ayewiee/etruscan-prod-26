package dto

import (
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type DecideRequest struct {
	UserID  string                 `json:"userId" validate:"required"`
	FlagKey string                 `json:"flagKey" validate:"required"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type DecisionResponse struct {
	ID    uuid.UUID   `json:"decisionId"`
	Value interface{} `json:"value"`
}

func DecisionResponseFromDomain(d *models.Decision) *DecisionResponse {
	return &DecisionResponse{
		ID:    d.ID,
		Value: d.Value,
	}
}
