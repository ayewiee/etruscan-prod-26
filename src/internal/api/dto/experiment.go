package dto

import (
	"encoding/json"
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type VariantRequest struct {
	Name      string          `json:"name" validate:"required"`
	Value     json.RawMessage `json:"value" validate:"required"`
	Weight    int             `json:"weight" validate:"required,gt=0,lte=100"`
	IsControl *bool           `json:"isControl"`
}

type CreateUpdateExperimentRequest struct {
	FlagID             uuid.UUID         `json:"flagId" validate:"required,uuid"`
	Name               string            `json:"name" validate:"required,min=5"`
	Description        *string           `json:"description"`
	AudiencePercentage int               `json:"audiencePercentage" validate:"required,gt=0,lte=100"`
	TargetingRule      *string           `json:"targetingRule"`
	Variants           []*VariantRequest `json:"variants" validate:"required,min=1,dive"`
}

type VariantResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Value     json.RawMessage `json:"value"`
	Weight    int             `json:"weight"`
	IsControl bool            `json:"isControl"`
}

type ExperimentResponse struct {
	ID                 uuid.UUID          `json:"id"`
	FlagID             uuid.UUID          `json:"flagId"`
	Name               string             `json:"name"`
	Description        *string            `json:"description"`
	CreatedBy          uuid.UUID          `json:"createdBy"`
	Status             string             `json:"status"`
	AudiencePercentage int                `json:"audiencePercentage"`
	TargetingRule      *string            `json:"targetingRule"`
	CreatedAt          string             `json:"createdAt"`
	UpdatedAt          string             `json:"updatedAt"`
	Variants           []*VariantResponse `json:"variants"`
}

func ExperimentResponseFromDomain(e *models.Experiment) *ExperimentResponse {
	variants := make([]*VariantResponse, len(e.Variants))
	for i, variant := range e.Variants {
		variants[i] = variantResponseFromDomain(variant)
	}
	return &ExperimentResponse{
		ID:                 e.ID,
		FlagID:             e.FlagID,
		Name:               e.Name,
		Description:        e.Description,
		CreatedBy:          e.CreatedBy,
		Status:             string(e.Status),
		AudiencePercentage: e.AudiencePercentage,
		TargetingRule:      e.TargetingRule,
		CreatedAt:          e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          e.UpdatedAt.Format(time.RFC3339),
		Variants:           variants,
	}
}

func variantResponseFromDomain(v *models.Variant) *VariantResponse {
	return &VariantResponse{
		ID:        v.ID,
		Name:      v.Name,
		Value:     v.Value,
		Weight:    v.Weight,
		IsControl: v.IsControl,
	}
}
