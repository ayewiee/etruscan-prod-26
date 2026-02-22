-- name: CreateEventType :one
INSERT INTO event_types (key, name, description, requires)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEventTypeByKey :one
SELECT et.id, et.key, et.name, et.description, et.requires, et.created_at, et2.key AS requires_key
FROM event_types et
LEFT JOIN event_types et2 ON et.requires = et2.id
WHERE et.key = $1;

-- name: GetEventTypeByID :one
SELECT et.id, et.key, et.name, et.description, et.requires, et.created_at, et2.key AS requires_key
FROM event_types et
LEFT JOIN event_types et2 ON et.requires = et2.id
WHERE et.id = $1;

-- name: ListEventTypes :many
SELECT et.id, et.key, et.name, et.description, et.requires, et.created_at, et2.key AS requires_key
FROM event_types et
LEFT JOIN event_types et2 ON et.requires = et2.id
ORDER BY et.created_at;
