-- name: CreateFlag :one
INSERT INTO flags (key, description, default_value, value_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFlagByID :one
SELECT * FROM flags WHERE id = $1;

-- name: GetFlagByKey :one
SELECT * FROM flags WHERE key = $1;

-- name: ListFlags :many
SELECT * FROM flags;

-- name: UpdateFlag :one
UPDATE flags
SET
    key = $2,
    description = $3,
    default_value = $4,
    value_type = $5
WHERE id = $1
RETURNING *;

-- name: DeleteFlag :exec
DELETE FROM flags WHERE id = $1;