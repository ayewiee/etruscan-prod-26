package models

import (
	"time"

	"github.com/google/uuid"
)

type Decision struct {
	ID           uuid.UUID
	ExperimentID *uuid.UUID
	VariantID    *uuid.UUID // Variant *Variant
	FlagKey      string

	Value   interface{}
	UserID  string
	Context map[string]interface{}

	CreatedAt time.Time
}
