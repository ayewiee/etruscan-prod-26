-- +goose Up
-- Derived (ratio) metrics for 5.4.3: conversion rate, error rate, click-through rate.
INSERT INTO metrics (key, name, description, type, event_type_key, aggregation_type, is_guardrail, numerator_metric_key, denominator_metric_key)
VALUES
    ('conversion_rate', 'Conversion rate', 'Conversions / exposures', 'continuous', NULL, NULL, false, 'conversion_count', 'exposure_count'),
    ('error_rate', 'Error rate', 'Errors / exposures (guardrail)', 'continuous', NULL, NULL, true, 'error_count', 'exposure_count'),
    ('click_through_rate', 'Click-through rate', 'Clicks / exposures', 'continuous', NULL, NULL, false, 'click_count', 'exposure_count')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM metrics WHERE key IN ('conversion_rate', 'error_rate', 'click_through_rate');
