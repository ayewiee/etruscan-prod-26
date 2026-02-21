-- +goose Up
INSERT INTO metrics (key, name, description, type, event_type_key, aggregation_type, is_guardrail)
VALUES
    ('exposure_count', 'Exposure count', 'Number of exposures', 'binomial', 'exposure', 'count', false),
    ('conversion_count', 'Conversion count', 'Number of conversions (requires exposure)', 'binomial', 'conversion', 'count', false),
    ('click_count', 'Click count', 'Number of clicks (requires exposure)', 'binomial', 'click', 'count', false),
    ('error_count', 'Error count', 'Number of errors (guardrail)', 'binomial', 'error', 'count', true),
    ('latency_p95', 'Latency P95 (ms)', '95th percentile latency (guardrail)', 'continuous', 'latency', 'p95', true)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM metrics WHERE key IN ('exposure_count', 'conversion_count', 'click_count', 'error_count', 'latency_p95');
