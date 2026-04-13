package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/academic/internal/shared"
)

var (
	ErrorInsertUser = shared.NewBadRequestError("Failed to insert user", "AUTH_UNKNOWN_ERROR")
	ErrorSqlError   = shared.NewInternalServerError("Failed to execute query", "AUHT_INTERNAL_ERROR")
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (pg *AuthRepository) InsertUser(ctx context.Context, user *User) (*uuid.UUID, error) {

	var id uuid.UUID
	err := pg.db.QueryRow(ctx, "INSERT INTO users (name) VALUES ($1) RETURNING id", user.Name).Scan(&id)
	fmt.Println(err, "INI ADALAH ERRORNYA CUY")
	if err != nil {
		return nil, ErrorSqlError
	}

	return &id, nil
}

func (pg *AuthRepository) GetUsers(ctx context.Context) ([]*User, error) {

	var users []*User
	err := pgxscan.Select(ctx, pg.db, &users, "SELECT * FROM users")
	fmt.Println(err, "INI ADALAH ERRORNYA CUY")
	if err != nil {
		return nil, ErrorSqlError
	}

	return users, nil
}

func (pg *AuthRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	res, err := pg.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if res.RowsAffected() == 0 {
		return shared.NewNotFoundError("User tidak ditemukan", "USER_NOT_FOUND")
	}
	if err != nil {
		slog.Error("Failed to delete user", "error", err)
		return ErrorSqlError
	}

	return nil
}
