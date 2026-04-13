package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionRepository memeriksa permission lewat role (tabel user_role, role_permission, permission).
// Kolom user_role.user_id harus bertipe UUID dan merujuk ke users(id).
type PermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// HasPermission untuk middleware.RequirePermission.
func (r *PermissionRepository) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_role ur
			INNER JOIN role_permission rp ON rp.role_id = ur.role_id
			INNER JOIN permission p ON p.id = rp.permission_id
			WHERE ur.user_id = $1 AND p.permission_name = $2
		)
	`, userID, permission).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
