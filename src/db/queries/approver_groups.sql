-- name: CreateApproverGroup :one
INSERT INTO approver_groups (name, description) VALUES ($1, $2)
RETURNING *;

-- name: GetApproverGroup :one
SELECT * FROM approver_groups WHERE id = $1;

-- name: GetApproverGroupMembers :many
SELECT
    u.id,
    u.email,
    u.username,
    u.role
FROM approver_group_memberships agm
JOIN users u ON approver_id = u.id
WHERE approver_group_id = $1
ORDER BY u.username ASC;

-- name: ListApproverGroups :many
SELECT * FROM approver_groups;

-- name: AddApproversToApproverGroup :exec
INSERT INTO approver_group_memberships (approver_id, approver_group_id)
SELECT unnest($1::uuid[]), $2
ON CONFLICT DO NOTHING;

-- name: RemoveApproversFromApproverGroup :exec
DELETE FROM approver_group_memberships
WHERE approver_group_id = $1 and approver_id = ANY($2::uuid[]);

-- name: DeleteApproverGroup :exec
DELETE FROM approver_groups WHERE id = $1;