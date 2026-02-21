package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Variant struct {
	ID           uuid.UUID
	ExperimentID uuid.UUID

	Name      string
	Value     json.RawMessage
	Weight    int // 0-100
	IsControl bool
}

type ExperimentStatus string

const (
	ExperimentStatusDraft    ExperimentStatus = "DRAFT"
	ExperimentStatusOnReview ExperimentStatus = "ON_REVIEW"
	ExperimentStatusApproved ExperimentStatus = "APPROVED"
	ExperimentStatusDeclined ExperimentStatus = "DECLINED"
	ExperimentStatusLaunched ExperimentStatus = "LAUNCHED"
	ExperimentStatusPaused   ExperimentStatus = "PAUSED"
	ExperimentStatusFinished ExperimentStatus = "FINISHED"
	ExperimentStatusArchived ExperimentStatus = "ARCHIVED"
)

func (s ExperimentStatus) CanTransitionTo(to ExperimentStatus) bool {
	// TASK.md 2.5
	switch s {
	case ExperimentStatusDraft:
		return to == ExperimentStatusOnReview
	case ExperimentStatusOnReview:
		return to == ExperimentStatusApproved || to == ExperimentStatusDraft || to == ExperimentStatusDeclined
	case ExperimentStatusApproved:
		return to == ExperimentStatusLaunched
	case ExperimentStatusLaunched:
		return to == ExperimentStatusPaused || to == ExperimentStatusFinished
	case ExperimentStatusPaused:
		return to == ExperimentStatusLaunched || to == ExperimentStatusFinished
	case ExperimentStatusFinished:
		return to == ExperimentStatusArchived
	default:
		return false
	}
}

func (s ExperimentStatus) EditingAllowed() bool {
	// TASK.md 2.3.2
	return s == ExperimentStatusDraft
}
func (s ExperimentStatus) ApprovalAllowed() bool {
	// TASK.md 2.3.2
	return s == ExperimentStatusOnReview
}

type ExperimentOutcome string

const (
	ExperimentOutcomeRollout  ExperimentOutcome = "ROLLOUT"
	ExperimentOutcomeRollback ExperimentOutcome = "ROLLBACK"
	ExperimentOutcomeNoEffect ExperimentOutcome = "NO_EFFECT"
)

type ExperimentReviewAction string

const (
	ExperimentReviewActionApprove        ExperimentReviewAction = "APPROVE"
	ExperimentReviewActionRequestChanges ExperimentReviewAction = "REQUEST_CHANGES"
	ExperimentReviewActionDecline        ExperimentReviewAction = "DECLINE"
)

type ExperimentReviewDecision string

const (
	ExperimentReviewDecisionApproved         ExperimentReviewDecision = "APPROVED"
	ExperimentReviewDecisionChangesRequested ExperimentReviewDecision = "CHANGES_REQUESTED"
	ExperimentReviewDecisionDeclined         ExperimentReviewDecision = "DECLINED"
)

type ExperimentReview struct {
	ID           uuid.UUID
	ExperimentID uuid.UUID
	ApproverID   uuid.UUID
	Decision     ExperimentReviewDecision
	Comment      *string
	CreatedAt    time.Time
}

type Experiment struct {
	ID      uuid.UUID
	FlagID  uuid.UUID
	FlagKey string

	Name        string
	Description *string
	CreatedBy   uuid.UUID
	Status      ExperimentStatus
	Version     int

	AudiencePercentage int
	TargetingRule      *string

	Reviews []*ExperimentReview

	Variants []*Variant

	Outcome        *ExperimentOutcome
	OutcomeComment *string
	OutcomeSetAt   *time.Time
	OutcomeSetBy   *uuid.UUID

	MetricKeys       []string
	PrimaryMetricKey *string
	Metrics          []*ExperimentMetricRef
	Guardrails       []*Guardrail

	CreatedAt time.Time
	UpdatedAt time.Time
}

type VariantSnapshotData struct {
	Name      string          `json:"name"`
	Value     json.RawMessage `json:"value"`
	Weight    int             `json:"weight"`
	IsControl bool            `json:"isControl"`
}

type ExperimentSnapshotData struct {
	ID     uuid.UUID `json:"id"`
	FlagID uuid.UUID `json:"flagId"`

	Name        string           `json:"name"`
	Description *string          `json:"description"`
	Status      ExperimentStatus `json:"status"`
	Version     int              `json:"version"`

	AudiencePercentage int     `json:"audiencePercentage"`
	TargetingRule      *string `json:"targetingRule"`

	Variants []*VariantSnapshotData `json:"variants"`
}

func (e *ExperimentSnapshotData) LoadFromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *ExperimentSnapshotData) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

type ExperimentSnapshot struct {
	ID           uuid.UUID
	ExperimentID uuid.UUID
	Version      int
	Data         *ExperimentSnapshotData
	CreatedAt    time.Time
}

type ExperimentStatusChange struct {
	ID           uuid.UUID
	ExperimentID uuid.UUID
	ActorID      *uuid.UUID
	From         *ExperimentStatus
	To           ExperimentStatus
	Comment      *string
	CreatedAt    time.Time
}
