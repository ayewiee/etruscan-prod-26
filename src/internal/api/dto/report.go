package dto

import (
	"time"

	"github.com/google/uuid"
)

type ExperimentReportQuery struct {
	From time.Time `query:"from" validate:"required"`
	To   time.Time `query:"to" validate:"required"`
}

type ExperimentReportResponse struct {
	ExperimentID uuid.UUID             `json:"experimentId"`
	From         string                `json:"from"`
	To           string                `json:"to"`
	Variants     []VariantMetricValues `json:"variants"`
}

type VariantMetricValues struct {
	ID      uuid.UUID          `json:"id"`
	Name    string             `json:"name"`
	Value   interface{}        `json:"value"`
	Metrics map[string]float64 `json:"metrics"`
}
