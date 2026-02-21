-- name: CreateDecision :one
INSERT INTO decisions (experiment_id, variant_id, flag_key, value, user_id, context)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetDecisionById :one
SELECT * FROM decisions d WHERE d.id = $1;
