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
	query := "INSERT INTO users (fullname, username, email, verified_email, created_at, updated_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	err := pg.db.QueryRow(ctx, query, user.Fullname, user.Username, user.Email, user.VerifiedEmail, user.CreatedAt, user.UpdatedAt, user.DeletedAt).Scan(&id)
	fmt.Println(err, "INI ADALAH ERRORNYA CUY")
	if err != nil {
		return nil, ErrorSqlError
	}

	return &id, nil
}

func (pg *AuthRepository) InsertAccount(ctx context.Context, acc *Account) (bool, error) {
	var id uuid.UUID
	query := "INSERT INTO accounts (user_id, password_hash, provider, provider_account_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err := pg.db.QueryRow(ctx, query, acc.UserId, acc.PasswordHash, acc.Provider, acc.ProviderAccountID, acc.CreatedAt, acc.UpdatedAt).Scan(&id)
	fmt.Println(err, "INI ADALAH ERRORNYA CUY")
	if err != nil {
		return false, ErrorSqlError
	}
	return true, nil
}

func (pg *AuthRepository) InsertSession(ctx context.Context, session *Session) (bool, error) {

	var id uuid.UUID
	query := "INSERT INTO sessions (user_id, active_role, token_hash, csrf_token, ip_address, user_agent, created_at, expired_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	err := pg.db.QueryRow(ctx, query, session.UserID, session.ActiveRole, session.TokenHash, session.CsrfToken, session.IPAddress, session.UserAgent, session.CreatedAt, session.ExpiredAt).Scan(&id)
	fmt.Println(err, "INI ADALAH ERRORNYA CUY")
	if err != nil {
		return false, ErrorSqlError
	}

	return true, nil
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
