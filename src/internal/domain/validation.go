package domain

import (
	"encoding/json"
	"errors"
	"etruscan/internal/domain/models"
)

func NewTypeMismatchFieldError(value any, expected string) *models.FieldError {
	return &models.FieldError{
		Field:         "value",
		Issue:         "type mismatch: expected " + expected,
		RejectedValue: value,
	}
}

func ValidateValueMatchesType(value json.RawMessage, valueType models.FlagValueType) error {
	var temp interface{}
	if err := json.Unmarshal(value, &temp); err != nil {
		return NewTypeMismatchFieldError(value, string(valueType))
	}

	switch valueType {
	case models.FlagValueTypeString:
		if _, ok := temp.(string); !ok {
			return NewTypeMismatchFieldError(temp, string(valueType))
		}
	case models.FlagValueTypeNumber:
		if _, ok := temp.(float64); !ok {
			return NewTypeMismatchFieldError(temp, string(valueType))
		}
	case models.FlagValueTypeBool:
		if _, ok := temp.(bool); !ok {
			return NewTypeMismatchFieldError(temp, string(valueType))
		}
	case models.FlagValueTypeJSON:
		return nil
	default:
		return errors.New("value type validation not implemented")
	}
	return nil
}
