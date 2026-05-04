package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/starter-kit/internal/shared/queries"
)

var (
	ErrRoleNameDuplicate = errors.New("role name already exists")
	ErrRoleCodeDuplicate = errors.New("role code already exists")
)

type Repository struct {
	db   *pgxpool.Pool
	repo *queries.Queries
}

func NewRepository(db *pgxpool.Pool, repo *queries.Queries) *Repository {
	return &Repository{db: db, repo: repo}
}

type RolePayload struct {
	RoleName     string
	RoleCode     string
	RoleCategory queries.RoleType
}

func (r *Repository) InsertRole(ctx context.Context, role *RolePayload) (*uuid.UUID, error) {
	roleID, err := r.repo.InsertRole(ctx, queries.InsertRoleParams{
		RoleName:     role.RoleName,
		RoleCode:     role.RoleCode,
		RoleCategory: role.RoleCategory,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "roles_role_name_key":
				return nil, ErrRoleNameDuplicate
			case "roles_role_code_key":
				return nil, ErrRoleCodeDuplicate
			}
		}
		return nil, fmt.Errorf("insert role: %w", err)
	}

	return &roleID, nil
}
