package dto

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type BatchEventsRequest struct {
	Events []BatchEventItemRequest `json:"events" validate:"required,min=1,dive"`
}

type BatchEventItemRequest struct {
	EventID      string                 `json:"eventId" validate:"required"`
	EventTypeKey string                 `json:"eventTypeKey" validate:"required"`
	DecisionID   uuid.UUID              `json:"decisionId" validate:"required"`
	UserID       string                 `json:"userId" validate:"required"`
	Timestamp    time.Time              `json:"timestamp" validate:"required"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

func (r BatchEventItemRequest) ToDomain() models.BatchEventItem {
	return models.BatchEventItem{
		EventID:      r.EventID,
		EventTypeKey: r.EventTypeKey,
		DecisionID:   r.DecisionID,
		UserID:       r.UserID,
		Timestamp:    r.Timestamp,
		Properties:   r.Properties,
	}
}

type BatchEventsResponse struct {
	Accepted   int                 `json:"accepted"`
	Duplicates int                 `json:"duplicates"`
	Rejected   int                 `json:"rejected"`
	Errors     []BatchEventErrResp `json:"errors,omitempty"`
}

type BatchEventErrResp struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func BatchEventsResponseFromDomain(res *models.BatchEventsResult) *BatchEventsResponse {
	errs := make([]BatchEventErrResp, len(res.Errors))
	for i, e := range res.Errors {
		errs[i] = BatchEventErrResp{Index: e.Index, Message: e.Message}
	}
	return &BatchEventsResponse{
		Accepted:   res.Accepted,
		Duplicates: res.Duplicates,
		Rejected:   res.Rejected,
		Errors:     errs,
	}
}
