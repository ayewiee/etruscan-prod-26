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

-- name: UpdateExperimentStatus :exec
UPDATE experiments SET status = $2 WHERE id = $1;

-- name: GetExperimentByID :one
SELECT * FROM experiments WHERE id = $1;

-- name: GetActiveExperimentByFlagKey :one
SELECT e.*
FROM experiments e
JOIN flags f ON e.flag_id = f.id
WHERE f.key = $1 AND e.status IN ('LAUNCHED', 'PAUSED');

-- name: GetRunningExperimentByFlagKey :one
SELECT e.*
FROM experiments e
JOIN flags f ON e.flag_id = f.id
WHERE f.key = $1 AND e.status = 'LAUNCHED';

-- name: ListExperimentStatusChanges :many
SELECT * FROM experiment_history WHERE experiment_id = $1
ORDER BY created_at DESC;

-- name: CountExperiments :one
SELECT COUNT(*) FROM experiments
WHERE
    (created_by = sqlc.narg('created_by') OR sqlc.narg('created_by') IS NULL)
  AND (status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
  AND (outcome = sqlc.narg('outcome') OR sqlc.narg('outcome') IS NULL)
  AND (flag_id = sqlc.narg('flag_id') OR sqlc.narg('flag_id') IS NULL);

-- name: ListExperiments :many
SELECT * FROM experiments
WHERE
    (created_by = sqlc.narg('created_by') OR sqlc.narg('created_by') IS NULL)
    AND (status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
    AND (outcome = sqlc.narg('outcome') OR sqlc.narg('outcome') IS NULL)
    AND (flag_id = sqlc.narg('flag_id') OR sqlc.narg('flag_id') IS NULL)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateExperimentSnapshot :exec
INSERT INTO experiment_snapshots (experiment_id, version, data)
VALUES ($1, $2, $3);

-- name: GetExperimentSnapshots :many
SELECT * FROM experiment_snapshots
WHERE experiment_id = $1
ORDER BY version DESC;

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
SELECT * FROM variants WHERE experiment_id = $1 ORDER BY weight DESC;

-- name: DeleteVariantsByExperiment :exec
DELETE FROM variants WHERE experiment_id = $1;


-- name: CreateExperimentReview :exec
INSERT INTO experiment_reviews (experiment_id, approver_id, decision, comment)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: ListExperimentReviews :many
SELECT * FROM experiment_reviews WHERE experiment_id = $1
ORDER BY created_at DESC;

-- name: CountApprovals :one
SELECT COUNT(*) FROM experiment_reviews WHERE experiment_id = $1 AND decision = 'APPROVED';

-- name: ClearExperimentReviews :exec
DELETE FROM experiment_reviews WHERE experiment_id = $1;

-- name: LogExperimentStatusChange :exec
INSERT INTO experiment_history (experiment_id, actor_id, from_status, to_status, comment)
VALUES ($1, $2, $3, $4, $5);

-- name: GetExperimentStatusChangeHistory :many
SELECT * FROM experiment_history WHERE experiment_id = $1
ORDER BY created_at DESC;
