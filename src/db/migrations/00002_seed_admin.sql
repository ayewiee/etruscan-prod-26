-- +goose Up
INSERT INTO users (id, email, username, password_hash, role, min_approvals, approver_group)
VALUES (
        '00000000-0000-0000-0000-000000000001',
        'admin@etruscan.com',
        'admin',
        '$2a$10$uhUTPCyyLPRsuek7OdFyaOy1KhYZ2KdLWRUZG.MDAsfBJxQTO4zAG',
        'ADMIN',
        null,
        null
) ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';
