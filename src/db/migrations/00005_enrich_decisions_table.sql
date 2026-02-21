-- +goose Up
DROP INDEX IF EXISTS idx_decisions_exp_variant;
DROP INDEX IF EXISTS idx_decisions_created_at;

DROP TABLE decisions;

CREATE TABLE decisions(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID REFERENCES experiments(id),
    variant_id UUID REFERENCES variants(id),

    flag_key TEXT NOT NULL REFERENCES flags(key),
    value JSONB NOT NULL,

    user_id TEXT NOT NULL,
    context JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_decisions_exp_variant ON decisions (experiment_id, variant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_decisions_exp_variant;

DROP TABLE decisions;

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
