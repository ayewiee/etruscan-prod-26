-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, role, min_approvals, approver_group, telegram_chat_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: ValidateApproversExistAndRole :one
WITH input_ids AS (SELECT UNNEST($1::uuid[]) AS id)
SELECT NOT EXISTS (
    SELECT 1
    FROM input_ids i
    LEFT JOIN users u ON u.id = i.id AND u.role IN ('APPROVER', 'ADMIN')
    WHERE u.id IS NULL
) AS all_valid;

-- name: ListUsers :many
SELECT * FROM users WHERE is_active = true LIMIT $1 OFFSET $2;

-- name: AdminUpdateUser :one
UPDATE users SET
    email = $2,
    username = $3,
    password_hash = $4,
    role = $5,
    min_approvals = $6,
    approver_group = $7,
    telegram_chat_id = $8
WHERE id = $1
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    username = $2,
    password_hash = $3,
    telegram_chat_id = $4
WHERE id = $1
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE users SET is_active = false WHERE id = $1;

