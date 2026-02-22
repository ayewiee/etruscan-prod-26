-- name: CreateMetric :one
INSERT INTO metrics (key, name, description, type, event_type_key, aggregation_type, is_guardrail, numerator_metric_key, denominator_metric_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetMetricByKey :one
SELECT id, key, name, description, type, event_type_key, aggregation_type, is_guardrail, created_at, numerator_metric_key, denominator_metric_key FROM metrics WHERE key = $1;

-- name: GetMetricByID :one
SELECT id, key, name, description, type, event_type_key, aggregation_type, is_guardrail, created_at, numerator_metric_key, denominator_metric_key FROM metrics WHERE id = $1;

-- name: ListMetrics :many
SELECT id, key, name, description, type, event_type_key, aggregation_type, is_guardrail, created_at, numerator_metric_key, denominator_metric_key FROM metrics ORDER BY key;

-- name: DeleteExperimentMetrics :exec
DELETE FROM experiment_metrics WHERE experiment_id = $1;

-- name: AddExperimentMetric :exec
INSERT INTO experiment_metrics (experiment_id, metric_id, is_primary)
VALUES ($1, $2, $3)
ON CONFLICT (experiment_id, metric_id) DO UPDATE SET is_primary = $3;

-- name: ListExperimentMetricIDs :many
SELECT metric_id, is_primary FROM experiment_metrics WHERE experiment_id = $1;
