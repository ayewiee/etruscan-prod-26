-- name: CreateNotificationSettings :exec
INSERT INTO notification_settings (user_id, experiment_id, severity, enable_telegram, enable_email)
VALUES ($1, $2, $3, $4, $5);

-- name: GetNotificationSettingsByID :one
SELECT * FROM notification_settings WHERE id = $1;

-- name: ListNotificationSettingsForExperiment :many
SELECT * FROM notification_settings WHERE experiment_id = $1;

-- name: GetNotificationSettingsForExperimentAndUser :one
SELECT * FROM notification_settings WHERE experiment_id = $1 AND user_id = $2;

-- name: ListNotificationSettingsForUser :many
SELECT * FROM notification_settings WHERE user_id = $1;

-- name: DeleteNotificationSettingsByID :exec
DELETE FROM notification_settings WHERE id = $1;
