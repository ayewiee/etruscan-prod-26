-- +goose Up
-- add derived metrics functionality

ALTER TABLE metrics
    ADD COLUMN numerator_metric_key TEXT REFERENCES metrics(key) ON DELETE RESTRICT,
    ADD COLUMN denominator_metric_key TEXT REFERENCES metrics(key) ON DELETE RESTRICT;

ALTER TABLE metrics
    ALTER COLUMN event_type_key DROP NOT NULL,
    ALTER COLUMN aggregation_type DROP NOT NULL;

ALTER TABLE metrics
    ADD CONSTRAINT metrics_primitive_or_derived CHECK (
        (event_type_key IS NOT NULL AND aggregation_type IS NOT NULL AND numerator_metric_key IS NULL AND denominator_metric_key IS NULL)
        OR
        (event_type_key IS NULL AND aggregation_type IS NULL AND numerator_metric_key IS NOT NULL AND denominator_metric_key IS NOT NULL)
    );

-- +goose Down
ALTER TABLE metrics DROP CONSTRAINT IF EXISTS metrics_primitive_or_derived;

ALTER TABLE metrics
    ALTER COLUMN event_type_key SET NOT NULL,
    ALTER COLUMN aggregation_type SET NOT NULL;

ALTER TABLE metrics
    DROP COLUMN IF EXISTS numerator_metric_key,
    DROP COLUMN IF EXISTS denominator_metric_key;
