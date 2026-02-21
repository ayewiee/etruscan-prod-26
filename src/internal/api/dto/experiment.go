package dto

import (
	"encoding/json"
	"etruscan/internal/domain/models"
	"fmt"
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

type ExperimentListFiltersQuery struct {
	FlagID    *uuid.UUID `query:"flagId" validate:"omitempty,uuid"`
	CreatedBy *uuid.UUID `query:"createdBy" validate:"omitempty,uuid"`
	Status    *string    `query:"status" validate:"omitempty,oneof=DRAFT ON_REVIEW APPROVED DECLINED LAUNCHED PAUSED FINISHED ARCHIVED"`
	Outcome   *string    `query:"outcome" validate:"omitempty,oneof=ROLLOUT ROLLBACK NO_EFFECT"`
	Page      *int       `query:"page" validate:"omitempty,gte=0"`
	Size      *int       `query:"size" validate:"omitempty,gte=0,lte=100"`
}

type ExperimentReviewRequest struct {
	Comment string `json:"comment" validate:"required"`
}

type ExperimentStatusChangeResponse struct {
	ID        uuid.UUID  `json:"id"`
	ActorID   *uuid.UUID `json:"actorId"`
	From      *string    `json:"from"`
	To        string     `json:"to"`
	Comment   *string    `json:"comment"`
	CreatedAt string     `json:"createdAt"`
}

type ExperimentSnapshotResponse struct {
	ID           uuid.UUID       `json:"id"`
	ExperimentID uuid.UUID       `json:"experimentId"`
	Version      int             `json:"version"`
	Data         json.RawMessage `json:"data"`
	CreatedAt    string          `json:"createdAt"`
}

type VariantResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Value     json.RawMessage `json:"value"`
	Weight    int             `json:"weight"`
	IsControl bool            `json:"isControl"`
}

type ReviewResponse struct {
	ID         uuid.UUID `json:"id"`
	ApproverID uuid.UUID `json:"approverId"`
	Decision   string    `json:"decision"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  string    `json:"createdAt"`
}

type ExperimentResponse struct {
	ID                 uuid.UUID          `json:"id"`
	FlagID             uuid.UUID          `json:"flagId"`
	Name               string             `json:"name"`
	Description        *string            `json:"description"`
	CreatedBy          uuid.UUID          `json:"createdBy"`
	Status             string             `json:"status"`
	Version            int                `json:"version"`
	Reviews            []*ReviewResponse  `json:"reviews"`
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
	reviews := make([]*ReviewResponse, len(e.Reviews))
	for i, review := range e.Reviews {
		reviews[i] = reviewResponseFromDomain(review)
	}

	return &ExperimentResponse{
		ID:                 e.ID,
		FlagID:             e.FlagID,
		Name:               e.Name,
		Description:        e.Description,
		CreatedBy:          e.CreatedBy,
		Status:             string(e.Status),
		Version:            e.Version,
		AudiencePercentage: e.AudiencePercentage,
		TargetingRule:      e.TargetingRule,
		CreatedAt:          e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          e.UpdatedAt.Format(time.RFC3339),
		Variants:           variants,
		Reviews:            reviews,
	}
}

func ExperimentResponseListFromDomain(experiments []*models.Experiment) []*ExperimentResponse {
	responses := make([]*ExperimentResponse, len(experiments))
	for i, experiment := range experiments {
		responses[i] = ExperimentResponseFromDomain(experiment)
	}
	return responses
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

func reviewResponseFromDomain(r *models.ExperimentReview) *ReviewResponse {
	return &ReviewResponse{
		ID:         r.ID,
		ApproverID: r.ApproverID,
		Decision:   string(r.Decision),
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.Format(time.RFC3339),
	}
}

func ExperimentStatusChangeResponseListFromDomain(changes []*models.ExperimentStatusChange) []*ExperimentStatusChangeResponse {
	responses := make([]*ExperimentStatusChangeResponse, len(changes))
	for i, change := range changes {
		responses[i] = experimentStatusChangeResponseFromDomain(change)
	}
	return responses
}

func experimentStatusChangeResponseFromDomain(c *models.ExperimentStatusChange) *ExperimentStatusChangeResponse {
	return &ExperimentStatusChangeResponse{
		ID:        c.ID,
		ActorID:   c.ActorID,
		From:      (*string)(c.From),
		To:        string(c.To),
		Comment:   c.Comment,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func ExperimentSnapshotResponseListFromDomain(snapshots []*models.ExperimentSnapshot) []*ExperimentSnapshotResponse {
	responses := make([]*ExperimentSnapshotResponse, len(snapshots))
	for i, snapshot := range snapshots {
		data, err := snapshot.Data.ToJSON()
		if err != nil {
			data = []byte(fmt.Sprintf("error converting snapshot to JSON: %s", err))
		}
		responses[i] = &ExperimentSnapshotResponse{
			ID:           snapshot.ID,
			ExperimentID: snapshot.ExperimentID,
			Version:      snapshot.Version,
			Data:         data,
			CreatedAt:    snapshot.CreatedAt.Format(time.RFC3339),
		}
	}

	return responses
}
