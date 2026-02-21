package usecases

import (
	"context"
	"encoding/json"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"sort"
	"time"

	"github.com/google/uuid"
)

// MetricComputer computes a single metric value over a time window for an experiment (optionally per variant).
type MetricComputer struct {
	decisionRepo  repository.DecisionRepository
	eventRepo     repository.EventRepository
	eventTypeRepo repository.EventTypeRepository
}

func NewMetricComputer(
	decisionRepo repository.DecisionRepository,
	eventRepo repository.EventRepository,
	eventTypeRepo repository.EventTypeRepository,
) *MetricComputer {
	return &MetricComputer{
		decisionRepo:  decisionRepo,
		eventRepo:     eventRepo,
		eventTypeRepo: eventTypeRepo,
	}
}

// Compute returns the metric value for the given experiment (and optional variant) in [from, to).
// variantID nil = all variants (experiment-level).
func (c *MetricComputer) Compute(
	ctx context.Context,
	experimentID uuid.UUID,
	variantID *uuid.UUID,
	metric *models.Metric,
	from, to time.Time,
) (float64, error) {
	decisionIDs, err := c.decisionRepo.ListDecisionIDsByExperimentVariantWindow(ctx, experimentID, variantID, from, to)
	if err != nil {
		return 0, err
	}
	if len(decisionIDs) == 0 {
		return 0, nil
	}

	events, err := c.eventRepo.ListByDecisionIDsAndWindow(ctx, decisionIDs, from, to)
	if err != nil {
		return 0, err
	}

	eventType, err := c.eventTypeRepo.GetByKey(ctx, metric.EventTypeKey)
	if err != nil {
		return 0, err
	}

	// if event type requires another (e.g. exposure), only count events whose decision_id has that type
	var exposureDecisionIDs map[uuid.UUID]struct{}
	if eventType.Requires != nil {
		requiredType, err := c.eventTypeRepo.GetByID(ctx, *eventType.Requires)
		if err != nil {
			return 0, err
		}

		requiredKey := requiredType.Key

		exposureDecisionIDs = make(map[uuid.UUID]struct{})
		for _, e := range events {
			if e.EventTypeKey == requiredKey && e.DecisionID != nil {
				exposureDecisionIDs[*e.DecisionID] = struct{}{}
			}
		}
	}

	var filtered []*models.Event
	for _, e := range events {
		if e.EventTypeKey != metric.EventTypeKey {
			continue
		}
		if exposureDecisionIDs != nil {
			if e.DecisionID == nil {
				continue
			}
			if _, ok := exposureDecisionIDs[*e.DecisionID]; !ok {
				continue
			}
		}
		filtered = append(filtered, e)
	}

	return aggregate(metric.AggregationType, filtered)
}

func aggregate(agg models.MetricAggregationType, events []*models.Event) (float64, error) {
	switch agg {
	case models.MetricAggregationCount:
		return float64(len(events)), nil
	case models.MetricAggregationSum, models.MetricAggregationAvg, models.MetricAggregationP95:
		values := extractNumericValues(events)
		if len(values) == 0 {
			return 0, nil
		}
		switch agg {
		case models.MetricAggregationSum:
			var s float64
			for _, v := range values {
				s += v
			}
			return s, nil
		case models.MetricAggregationAvg:
			var s float64
			for _, v := range values {
				s += v
			}
			return s / float64(len(values)), nil
		case models.MetricAggregationP95:
			sort.Float64s(values)

			idx := int(0.95 * float64(len(values)))
			if idx >= len(values) {
				idx = len(values) - 1
			}
			if idx < 0 {
				idx = 0
			}
			return values[idx], nil
		}
	}
	return 0, nil
}

func extractNumericValues(events []*models.Event) []float64 {
	var out []float64
	for _, e := range events {
		if len(e.Properties) == 0 {
			continue
		}

		var m map[string]interface{}
		if err := json.Unmarshal(e.Properties, &m); err != nil {
			continue
		}

		if v, ok := m["value"]; ok {
			switch n := v.(type) {
			case float64:
				out = append(out, n)
			case int:
				out = append(out, float64(n))
			case int64:
				out = append(out, float64(n))
			}
		}
	}
	return out
}
