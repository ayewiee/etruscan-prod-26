-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TYPE user_role AS ENUM(
    'ADMIN',
    'EXPERIMENTER',
    'APPROVER',
    'VIEWER'
);

CREATE TABLE approver_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    role user_role NOT NULL,
    min_approvals INT,
    approver_group UUID REFERENCES approver_groups(id) ON DELETE SET NULL,

    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE approver_group_memberships (
    approver_id UUID REFERENCES users(id) ON DELETE CASCADE,
    approver_group_id UUID NOT NULL REFERENCES approver_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (approver_id, approver_group_id)
);


CREATE TABLE flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL UNIQUE,
    description TEXT,

    default_value JSONB NOT NULL,
    value_type TEXT NOT NULL CHECK (value_type IN ('string', 'number', 'bool', 'json')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_flags_key ON flags(key);


CREATE TYPE experiment_status AS ENUM(
    'DRAFT',
    'ON_REVIEW',
    'APPROVED',
    'LAUNCHED',
    'PAUSED',
    'FINISHED',
    'ARCHIVED',
    'DECLINED'
);

CREATE TYPE experiment_outcome AS ENUM(
    'ROLLOUT',
    'ROLLBACK',
    'NO_EFFECT'
);

CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id UUID NOT NULL REFERENCES flags(id),

    name TEXT NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),

    status experiment_status NOT NULL,
    version INT NOT NULL DEFAULT 1,

    audience_pct INT NOT NULL DEFAULT 100 CHECK (audience_pct >= 0 AND audience_pct <= 100),
    targeting_rule TEXT,

    outcome experiment_outcome,
    outcome_comment TEXT,
    outcome_set_at TIMESTAMPTZ,
    outcome_set_by UUID REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE UNIQUE INDEX idx_one_active_experiment_per_flag ON experiments(flag_id)
WHERE status IN ('LAUNCHED', 'PAUSED');


CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id),

    name TEXT NOT NULL,
    value JSONB NOT NULL,
    weight INT NOT NULL CHECK (weight >= 0 AND weight <= 100),
    is_control BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT unique_variant_name_per_exp UNIQUE (experiment_id, name)
);

CREATE TABLE experiment_approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL REFERENCES users(id),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (experiment_id, approver_id)
);

CREATE TABLE experiment_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id),
    from_status experiment_status,
    to_status experiment_status NOT NULL,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE event_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,

    requires UUID REFERENCES event_types(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type_key TEXT NOT NULL REFERENCES event_types(key),

    decision_id UUID,

    user_id TEXT NOT NULL,
    properties JSONB DEFAULT '{}'::jsonb,

    timestamp TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_decision_id ON events (decision_id);
CREATE INDEX idx_events_user_id ON events (user_id);


CREATE TYPE metric_type AS ENUM(
    'binomial',
    'continuous'
);

CREATE TYPE metric_aggregation_type AS ENUM(
    'count',
    'sum',
    'avg',
    'p95'
);

CREATE TABLE metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    type metric_type NOT NULL,

    event_type_key TEXT NOT NULL REFERENCES event_types(key),

    aggregation_type metric_aggregation_type NOT NULL,

    is_guardrail BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE experiment_metrics (
    experiment_id UUID NOT NULL REFERENCES experiments(id),
    metric_id UUID NOT NULL REFERENCES metrics(id),
    is_primary BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (experiment_id, metric_id)
);


CREATE TABLE guardrails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id),
    metric_id UUID NOT NULL REFERENCES metrics(id),

    threshold DOUBLE PRECISION NOT NULL,
    threshold_direction TEXT NOT NULL CHECK (threshold_direction IN ('upper', 'lower')),

    action TEXT NOT NULL CHECK (action IN ('pause', 'rollback')),

    window_seconds INT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE guardrail_triggers(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    guardrail_id UUID NOT NULL REFERENCES guardrails(id),
    experiment_id UUID NOT NULL REFERENCES experiments(id),

    metric_value DOUBLE PRECISION NOT NULL,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE decisions(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id),
    variant_id UUID NOT NULL REFERENCES variants(id),

    user_id TEXT NOT NULL,
    context JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_decisions_exp_variant ON decisions (experiment_id, variant_id);
CREATE INDEX idx_decisions_created_at ON decisions (created_at);


CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER set_timestamp_flags BEFORE UPDATE ON flags
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER set_timestamp_experiments BEFORE UPDATE ON experiments
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- +goose Down
DROP TRIGGER set_timestamp_users ON users;
DROP TRIGGER set_timestamp_flags ON flags;
DROP TRIGGER set_timestamp_experiments ON experiments;

DROP INDEX idx_one_active_experiment_per_flag;
DROP INDEX idx_events_decision_id, idx_events_user_id;
DROP INDEX idx_decisions_exp_variant, idx_decisions_created_at;

DROP TABLE users, approver_groups, approver_group_memberships;
DROP TABLE flags;
DROP TABLE experiments, variants, experiment_approvals, experiment_history;
DROP TABLE metrics, experiment_metrics;
DROP TABLE guardrails, guardrail_triggers;

DROP TYPE user_role;
DROP TYPE experiment_status;
DROP TYPE metric_type, metric_aggregation_type;