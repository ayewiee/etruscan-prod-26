-- +goose Up
-- metrics: metric_key for guardrails
ALTER TABLE metrics ADD COLUMN IF NOT EXISTS key TEXT;
UPDATE metrics SET key = 'm_' || id::text WHERE key IS NULL;
ALTER TABLE metrics ALTER COLUMN key SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS metrics_key_key ON metrics(key);

-- events: add client_event_id
ALTER TABLE events ADD COLUMN IF NOT EXISTS client_event_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_events_dedup ON events(event_type_key, client_event_id)
WHERE client_event_id IS NOT NULL;

-- guardrail_triggers: save fields for audit
ALTER TABLE guardrail_triggers ADD COLUMN IF NOT EXISTS metric_key TEXT;
ALTER TABLE guardrail_triggers ADD COLUMN IF NOT EXISTS threshold_value DOUBLE PRECISION;
ALTER TABLE guardrail_triggers ADD COLUMN IF NOT EXISTS window_seconds INT;
ALTER TABLE guardrail_triggers ADD COLUMN IF NOT EXISTS action TEXT;

-- +goose Down
ALTER TABLE guardrail_triggers DROP COLUMN IF EXISTS metric_key;
ALTER TABLE guardrail_triggers DROP COLUMN IF EXISTS threshold_value;
ALTER TABLE guardrail_triggers DROP COLUMN IF EXISTS window_seconds;
ALTER TABLE guardrail_triggers DROP COLUMN IF EXISTS action;

DROP INDEX IF EXISTS idx_events_dedup;
ALTER TABLE events DROP COLUMN IF EXISTS client_event_id;

DROP INDEX IF EXISTS metrics_key_key;
ALTER TABLE metrics DROP COLUMN IF EXISTS key;
