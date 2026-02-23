-- +goose Up

CREATE TYPE notification_severity AS ENUM(
    'LOW',
    'HIGH'
);

CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    severity notification_severity NOT NULL,
    enable_telegram BOOLEAN NOT NULL,
    enable_email BOOLEAN NOT NULL,
    UNIQUE (user_id, experiment_id)
);

ALTER TABLE users
ADD COLUMN telegram_chat_id TEXT;

-- +goose Down
ALTER TABLE users
DROP COLUMN telegram_chat_id;

DROP TABLE notification_settings;
DROP TYPE notification_severity;
