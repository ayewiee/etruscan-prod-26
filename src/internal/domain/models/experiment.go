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
	ID     uuid.UUID
	FlagID uuid.UUID

	Name        string
	Description *string
	CreatedBy   uuid.UUID
	Status      ExperimentStatus

	AudiencePercentage int
	TargetingRule      *string

	Reviews []*ExperimentReview

	Variants []*Variant

	Outcome        *ExperimentOutcome
	OutcomeComment *string
	OutcomeSetAt   *time.Time
	OutcomeSetBy   *uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
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
