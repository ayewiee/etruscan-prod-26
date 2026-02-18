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
		return to == ExperimentStatusApproved || to == ExperimentStatusDeclined
	case ExperimentStatusDeclined:
		return to == ExperimentStatusDraft
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

func (s ExperimentStatus) CanBeEdited(to ExperimentStatus) bool {
	// TASK.md 2.3.2
	return to == ExperimentStatusDraft || to == ExperimentStatusOnReview || to == ExperimentStatusApproved || to == ExperimentStatusDeclined
}

type ExperimentOutcome string

const (
	ExperimentOutcomeRollout  ExperimentOutcome = "ROLLOUT"
	ExperimentOutcomeRollback ExperimentOutcome = "ROLLBACK"
	ExperimentOutcomeNoEffect ExperimentOutcome = "NO_EFFECT"
)

type Experiment struct {
	ID     uuid.UUID
	FlagID uuid.UUID

	Name        string
	Description *string
	CreatedBy   uuid.UUID
	Status      ExperimentStatus

	AudiencePercentage int
	TargetingRule      *string

	Outcome        *ExperimentOutcome
	OutcomeComment *string
	OutcomeSetAt   *time.Time
	OutcomeSetBy   *uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	Variants []*Variant
}
