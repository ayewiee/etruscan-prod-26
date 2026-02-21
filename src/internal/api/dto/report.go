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
	VariantID   uuid.UUID          `json:"variantId"`
	VariantName string             `json:"variantName"`
	Metrics     map[string]float64 `json:"metrics"`
}
