-- name: CreateDecision :one
INSERT INTO decisions (experiment_id, variant_id, flag_key, value, user_id, context)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetDecisionById :one
SELECT * FROM decisions d WHERE d.id = $1;

-- name: ListDecisionIDsByExperimentVariantWindow :many
SELECT id FROM decisions
WHERE experiment_id = $1
  AND (sqlc.narg('variant_id')::uuid IS NULL OR variant_id = sqlc.narg('variant_id'))
  AND created_at >= $2
  AND created_at < $3;
