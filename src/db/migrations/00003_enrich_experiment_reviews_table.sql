-- +goose Up
CREATE TYPE experiment_review_decision AS ENUM(
    'APPROVED',
    'CHANGES_REQUESTED',
    'DECLINED'
);

ALTER TABLE experiment_approvals
RENAME TO experiment_reviews;

ALTER TABLE experiment_reviews
ADD COLUMN decision experiment_review_decision NOT NULL;

-- +goose Down
ALTER TABLE experiment_reviews
DROP COLUMN decision;

ALTER TABLE experiment_reviews
RENAME TO experiment_approvals;
