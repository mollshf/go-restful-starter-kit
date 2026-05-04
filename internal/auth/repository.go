package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/starter-kit/internal/shared/queries"
)

var (
	ErrEmailDuplicate    = errors.New("email already exists")
	ErrUsernameDuplicate = errors.New("username already exists")
	ErrAccountNotFound   = errors.New("account not found")
)

type AuthRepository struct {
	db   *pgxpool.Pool
	repo *queries.Queries
}

func NewAuthRepository(db *pgxpool.Pool, repo *queries.Queries) *AuthRepository {
	return &AuthRepository{db: db, repo: repo}
}

// FindAccountByUsername is a function to find an account by username
func (pg *AuthRepository) FindAccountByUsername(ctx context.Context, username string) (*queries.FindAccountByUsernameRow, error) {
	account, err := pg.repo.FindAccountByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account by username: %w", err)
	}
	return &account, nil
}

// InsertUserCredential is a function to insert a user credential using transaction
type CreateUserCredentialPayload struct {
	DtoCreateUser  *CreateUserRequest
	HashedPassword string
	Provider       Provider
}

func (pg *AuthRepository) CreateUserCredential(ctx context.Context, payload *CreateUserCredentialPayload) (*uuid.UUID, error) {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.repo.WithTx(tx)

	// Tx step 1: insert into users — create the new user identity.
	userID, err := qtx.InsertUser(ctx, queries.InsertUserParams{
		Fullname:      payload.DtoCreateUser.Fullname,
		Username:      payload.DtoCreateUser.Username,
		Email:         payload.DtoCreateUser.Email,
		VerifiedEmail: false,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return nil, ErrEmailDuplicate
			case "users_username_key":
				return nil, ErrUsernameDuplicate
			}
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	// Tx step 2: insert into user_accounts — attach the password credential to the user.
	_, err = qtx.InsertAccount(ctx, queries.InsertAccountParams{
		UserID:            userID,
		PasswordHash:      &payload.HashedPassword,
		Provider:          string(payload.Provider),
		ProviderAccountID: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("insert account: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &userID, nil
}

// DeleteSessionByTokenHash removes a single session row by its token hash.
// Used by logout — silent success if no row matches (idempotent).
func (pg *AuthRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	if err := pg.repo.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("AUTH_REPOSITORY_DELETE_SESSION_BY_TOKEN_HASH_ERROR: %w", err)
	}
	return nil
}

// DeleteAllSessionsByUserID removes every session row owned by the user.
// Used by logout-all and on password change to forcibly invalidate all
// devices currently signed in.
func (pg *AuthRepository) DeleteAllSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := pg.repo.DeleteAllSessionsByUserID(ctx, userID); err != nil {
		return fmt.Errorf("AUTH_REPOSITORY_DELETE_ALL_SESSIONS_BY_USER_ID_ERROR: %w", err)
	}
	return nil
}

// FindSessionByTokenHash returns the session row for a token hash, or nil if not found.
// Caller is responsible for checking expired_at.
func (pg *AuthRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*queries.FindSessionByTokenHashRow, error) {
	session, err := pg.repo.FindSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("AUTH_REPOSITORY_FIND_SESSION_BY_TOKEN_HASH_ERROR: %w", err)
	}
	return &session, nil
}

// FindActiveSessionsByUserID returns all non-expired sessions of a user,
// ordered newest first. Empty slice if none.
func (pg *AuthRepository) FindActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]queries.FindActiveSessionsByUserIDRow, error) {
	sessions, err := pg.repo.FindActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finc active session by user id: %w", err)
	}
	return sessions, nil
}

// FindUserByID returns the user row by id, or nil if soft-deleted / not found.
func (pg *AuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*queries.FindUserByIDRow, error) {
	user, err := pg.repo.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("AUTH_REPOSITORY_FIND_USER_BY_ID_ERROR: %w", err)
	}
	return &user, nil
}

// CreateSession inserts a new session and enforces the max active sessions
// per user in a single transaction.
type CreateSessionPayload struct {
	UserID     uuid.UUID
	TokenHash  string
	UserAgent  string
	IpAddress  netip.Addr
	ActiveRole *uuid.UUID
	MaxActive  int32
	ExpiresAt  time.Time
}

func (pg *AuthRepository) CreateSession(ctx context.Context, payload *CreateSessionPayload) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := pg.repo.WithTx(tx)

	// Tx step 1: insert a new row into user_sessions — token_hash, ip, user_agent, active_role.
	if err := qtx.InsertSession(ctx, queries.InsertSessionParams{
		UserID:     payload.UserID,
		TokenHash:  payload.TokenHash,
		UserAgent:  payload.UserAgent,
		IpAddress:  payload.IpAddress,
		ActiveRole: payload.ActiveRole,
		ExpiresAt:  payload.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	// Tx step 2: delete the user's oldest sessions when the total exceeds MaxActive — enforce the concurrent session cap per user.
	if err := qtx.DeleteSessionsBeyondLimit(ctx, queries.DeleteSessionsBeyondLimitParams{
		UserID: payload.UserID,
		Offset: payload.MaxActive,
	}); err != nil {
		return fmt.Errorf("delete sessions beyond limit: %w", err)
	}

	// Tx step 3: update last_login_at on users — record when the user last successfully authenticated.
	if err := qtx.UpdateLastLoginAt(ctx, payload.UserID); err != nil {
		return fmt.Errorf("update last login at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
