package dto

import (
	"encoding/json"
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CreateUpdateFlagRequest struct {
	Key          string          `json:"key" validate:"required"`
	Description  *string         `json:"description" validate:"omitempty,required"`
	DefaultValue json.RawMessage `json:"defaultValue"`
	ValueType    string          `json:"valueType" validate:"required,oneof=string number bool json"`
}

type FlagResponse struct {
	ID           uuid.UUID       `json:"id"`
	Key          string          `json:"key"`
	Description  *string         `json:"description"`
	DefaultValue json.RawMessage `json:"defaultValue"`
	ValueType    string          `json:"valueType"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

func FlagResponseFromDomain(flag *models.Flag) FlagResponse {
	return FlagResponse{
		ID:           flag.ID,
		Key:          flag.Key,
		Description:  flag.Description,
		DefaultValue: flag.DefaultValue,
		ValueType:    string(flag.ValueType),
		CreatedAt:    flag.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    flag.UpdatedAt.Format(time.RFC3339),
	}
}

func FlagResponseListFromDomain(flags []*models.Flag) []FlagResponse {
	flagResponses := make([]FlagResponse, len(flags))
	for i, flag := range flags {
		flagResponses[i] = FlagResponseFromDomain(flag)
	}
	return flagResponses
}
