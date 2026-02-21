package models

import (
	"time"

	"github.com/google/uuid"
)

type MetricType string

const (
	MetricTypeBinomial   MetricType = "binomial"
	MetricTypeContinuous MetricType = "continuous"
)

type MetricAggregationType string

const (
	MetricAggregationCount MetricAggregationType = "count"
	MetricAggregationSum   MetricAggregationType = "sum"
	MetricAggregationAvg   MetricAggregationType = "avg"
	MetricAggregationP95   MetricAggregationType = "p95"
)

type Metric struct {
	ID              uuid.UUID
	Key             string
	Name            string
	Description     *string
	Type            MetricType
	EventTypeKey    string
	AggregationType MetricAggregationType
	IsGuardrail     bool
	CreatedAt       *time.Time
}

type ExperimentMetric struct {
	ExperimentID uuid.UUID
	MetricID     uuid.UUID
	MetricKey    string
	IsPrimary    bool
}

type ExperimentMetricRef struct {
	Metric    *Metric
	IsPrimary bool
}

type Guardrail struct {
	ID                 uuid.UUID
	ExperimentID       uuid.UUID
	MetricID           uuid.UUID
	MetricKey          string
	Threshold          float64
	ThresholdDirection string // "upper" or "lower"
	Action             string // "pause" or "rollback"
	WindowSeconds      int
}
