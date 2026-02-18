-- name: CreateExperiment :one
INSERT INTO experiments(flag_id, name, description, created_by, status, audience_pct, targeting_rule)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateExperiment :one
UPDATE experiments
SET
    name = $2,
    description = $3,
    audience_pct = $4,
    targeting_rule = $5,
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: UpdateExperimentStatus :one
UPDATE experiments SET status = $2 WHERE id = $1 RETURNING *;

-- name: GetExperimentByID :one
SELECT * FROM experiments WHERE id = $1;

-- name: ListExperiments :many
SELECT * FROM experiments
WHERE
    (created_by = sqlc.narg('created_by') OR sqlc.narg('created_by') IS NULL)
    AND (status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
    AND (outcome = sqlc.narg('outcome') OR sqlc.narg('outcome') IS NULL)
    AND (flag_id = sqlc.narg('flag_id') OR sqlc.narg('flag_id') IS NULL);

-- name: FinishExperiment :one
UPDATE experiments
SET
    status = $2,
    outcome = $3,
    outcome_comment = $4,
    outcome_set_by = $5,
    outcome_set_at = now()
WHERE id = $1
RETURNING *;

-- name: BatchCreateVariants :copyfrom
INSERT INTO variants (experiment_id, name, value, weight, is_control)
VALUES ($1, $2, $3, $4, $5);

-- name: ListVariantsByExperiment :many
SELECT * FROM variants WHERE experiment_id = $1;

-- name: DeleteVariantsByExperiment :exec
DELETE FROM variants WHERE experiment_id = $1;