-- =====================================================================
-- USERS
-- =====================================================================

-- name: FindUser :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: FindUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: FindUsersPaginated :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE deleted_at IS NULL;

-- name: FindUserByUsername :one
SELECT *
FROM users
WHERE username = $1
  AND deleted_at IS NULL;

-- name: FindUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: ExistsUserByUsername :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE username = $1
      AND deleted_at IS NULL
);

-- name: ExistsUserByEmail :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1
      AND deleted_at IS NULL
);

-- name: InsertUser :one
INSERT INTO users
    (fullname, username, email, verified_email)
VALUES
    ($1, $2, $3, $4)
RETURNING id;

-- name: UpdateUser :exec
UPDATE users
SET fullname   = $2,
    username   = $3,
    email      = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL;

-- name: VerifyUserEmail :exec
UPDATE users
SET verified_email = TRUE,
    updated_at     = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL;

-- name: RestoreUser :exec
UPDATE users
SET deleted_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- =====================================================================
-- USER ROLES
-- =====================================================================

-- name: AssignRoleToUser :one
INSERT INTO user_roles
    (user_id, role_id)
VALUES
    ($1, $2)
RETURNING id;

-- name: RemoveRoleFromUser :exec
UPDATE user_roles
SET deleted_at = CURRENT_TIMESTAMP
WHERE user_id = $1
  AND role_id = $2
  AND deleted_at IS NULL;

-- name: RemoveAllRolesFromUser :exec
UPDATE user_roles
SET deleted_at = CURRENT_TIMESTAMP
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: FindRolesByUserID :many
SELECT r.*
FROM roles r
INNER JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
  AND ur.deleted_at IS NULL
  AND r.deleted_at IS NULL
ORDER BY r.role_name;

-- name: ExistsUserRole :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = $1
      AND role_id = $2
      AND deleted_at IS NULL
);

-- =====================================================================
-- USER PERMISSIONS (direct permissions attached to a user)
-- =====================================================================

-- name: AssignPermissionToUser :one
INSERT INTO user_permissions
    (user_id, permission_id)
VALUES
    ($1, $2)
RETURNING id;

-- name: RemovePermissionFromUser :exec
UPDATE user_permissions
SET deleted_at = CURRENT_TIMESTAMP
WHERE user_id = $1
  AND permission_id = $2
  AND deleted_at IS NULL;

-- name: FindDirectPermissionsByUserID :many
SELECT p.*
FROM permissions p
INNER JOIN user_permissions up ON up.permission_id = p.id
WHERE up.user_id = $1
  AND up.deleted_at IS NULL
  AND p.deleted_at IS NULL
ORDER BY p.module, p.act;

-- name: FindEffectivePermissionsByUserID :many
SELECT DISTINCT p.*
FROM permissions p
LEFT JOIN role_permissions rp ON rp.permission_id = p.id
LEFT JOIN user_roles ur       ON ur.role_id = rp.role_id
LEFT JOIN user_permissions up ON up.permission_id = p.id
WHERE p.deleted_at IS NULL
  AND (
        (ur.user_id = $1 AND ur.deleted_at IS NULL AND rp.deleted_at IS NULL)
     OR (up.user_id = $1 AND up.deleted_at IS NULL)
  )
ORDER BY p.module, p.act;

-- =====================================================================
-- ROLES
-- =====================================================================

-- name: InsertRole :one
INSERT INTO roles
    (role_name, role_code, role_category)
VALUES
    ($1, $2, $3)
RETURNING id;

-- name: SoftDeleteRole :exec
UPDATE roles
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL;


-- name: FindRoles :many
SELECT *
FROM roles
WHERE deleted_at IS NULL;

