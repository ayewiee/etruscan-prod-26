-- name: GetEventTypeByKey :one
SELECT * FROM event_types WHERE key = $1;

-- name: GetEventTypeByID :one
SELECT * FROM event_types WHERE id = $1;

-- name: ListEventTypes :many
SELECT * FROM event_types ORDER BY key;
