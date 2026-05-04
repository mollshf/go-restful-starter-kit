package rbac

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/starter-kit/internal/shared/queries"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool, repo *queries.Queries) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertRole(ctx context.Context) (uuid.UUID, error) {
	return uuid.Nil, nil
}
