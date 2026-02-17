package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FlagValueType string

var (
	FlagValueTypeString FlagValueType = "string"
	FlagValueTypeNumber FlagValueType = "number"
	FlagValueTypeBool   FlagValueType = "bool"
	FlagValueTypeJSON   FlagValueType = "json"
)

type Flag struct {
	ID           uuid.UUID
	Key          string
	Description  *string
	DefaultValue json.RawMessage
	ValueType    FlagValueType
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
