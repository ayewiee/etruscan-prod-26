-- name: CreateGuardrail :one
INSERT INTO guardrails (experiment_id, metric_id, threshold, threshold_direction, action, window_seconds)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListGuardrailsByExperimentID :many
SELECT * FROM guardrails WHERE experiment_id = $1;

-- name: DeleteGuardrailsByExperimentID :exec
DELETE FROM guardrails WHERE experiment_id = $1;

-- name: CreateGuardrailTrigger :one
INSERT INTO guardrail_triggers (guardrail_id, experiment_id, metric_value, metric_key, threshold_value, window_seconds, action)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;
