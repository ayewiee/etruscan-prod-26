-- +goose Up
INSERT INTO event_types (id, key, name, description, requires)
VALUES
    ('a0000000-0000-0000-0000-000000000001'::uuid, 'exposure', 'Exposure', 'Fact of showing a variant to the user', NULL),
    ('a0000000-0000-0000-0000-000000000002'::uuid, 'click', 'Click', 'User click', 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('a0000000-0000-0000-0000-000000000003'::uuid, 'conversion', 'Conversion', 'Target action (e.g. purchase)', 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('a0000000-0000-0000-0000-000000000004'::uuid, 'error', 'Error', 'Error event', 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('a0000000-0000-0000-0000-000000000005'::uuid, 'latency', 'Latency', 'Response latency (ms)', 'a0000000-0000-0000-0000-000000000001'::uuid)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM event_types WHERE key IN ('exposure', 'click', 'conversion', 'error', 'latency');
