-- name: CreateEvent :one
INSERT INTO events (event_type_key, decision_id, user_id, properties, timestamp, client_event_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: ExistsEventByTypeAndClientID :one
SELECT EXISTS(
    SELECT 1 FROM events
    WHERE event_type_key = $1 AND client_event_id = $2
);

-- name: ListEventsByDecisionIDsAndWindow :many
SELECT e.* FROM events e
WHERE e.decision_id = ANY($1::uuid[])
  AND e.timestamp >= $2
  AND e.timestamp < $3
ORDER BY e.timestamp;
