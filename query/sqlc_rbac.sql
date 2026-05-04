-- name: RBACInsertRole :one
INSERT INTO roles
    (role_name, role_code)
VALUES
    ($1, $2)
RETURNING id;

