-- name: InsertSession :exec
INSERT INTO user_sessions
(user_id, active_role, token_hash, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertAccount :one
INSERT INTO user_accounts
(user_id, password_hash, provider, provider_account_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: FindAccountByUsername :one
SELECT u.id AS user_id, u.username, u.email, u.verified_email, a.password_hash
FROM users u
INNER JOIN user_accounts a ON a.user_id = u.id
WHERE u.username = $1
  AND u.deleted_at IS NULL;

-- name: DeleteSessionsBeyondLimit :exec
DELETE FROM user_sessions
WHERE id IN (
    SELECT s.id
    FROM user_sessions s
    WHERE s.user_id = $1
    ORDER BY s.created_at DESC
    OFFSET $2
);

-- name: DeleteExpiredSessionsByUser :exec
DELETE FROM user_sessions
WHERE user_id = $1 AND expires_at < NOW();

-- name: DeleteSessionByTokenHash :exec
DELETE FROM user_sessions
WHERE token_hash = $1;

-- name: DeleteAllSessionsByUserID :exec
DELETE FROM user_sessions
WHERE user_id = $1;

-- name: FindSessionByTokenHash :one
SELECT id, user_id, expires_at, active_role
FROM user_sessions
WHERE token_hash = $1;

-- name: FindActiveSessionsByUserID :many
SELECT id, ip_address, user_agent, active_role, created_at, expires_at
FROM user_sessions
WHERE user_id = $1
  AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY created_at DESC;

-- name: FindUserByID :one
SELECT id, fullname, username, email, verified_email
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateLastLoginAt :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;