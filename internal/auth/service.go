package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/mollshf/starter-kit/internal/shared/queries"
	"github.com/mollshf/starter-kit/internal/shared/utility"
)

const (
	MaxSessionsPerUser = 3
	SessionDuration    = 7 * 24 * time.Hour // 7 days
)

type Provider string

var (
	ProviderCredentials Provider = "credentials"
)

type AuthService struct {
	repo *AuthRepository
}

func NewAuthService(authRepository *AuthRepository) (*AuthService, error) {
	if authRepository == nil {
		return nil, utility.NewInternalServerError("User service is required", "AUTH_UNKNOWN_ERROR")
	}
	return &AuthService{
		repo: authRepository,
	}, nil
}

type LoginInput struct {
	Username  string
	Password  string
	UserAgent string
	IpAddress netip.Addr
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
}

func (s *AuthService) Login(ctx context.Context, in *LoginInput) (LoginResult, error) {
	account, err := s.repo.FindAccountByUsername(ctx, in.Username)
	if err != nil {
		// check if account is not found
		if errors.Is(err, ErrAccountNotFound) {
			return LoginResult{}, utility.NewNotFoundError("Akun tidak ditemukan", "AUTH_USER_NOT_FOUND")
		}
		return LoginResult{}, err
	}

	// verify password
	if err := utility.VerifyPassword(in.Password, *account.PasswordHash); err != nil {
		return LoginResult{}, err
	}

	// generate token
	token, err := utility.GenerateSecureToken()
	if err != nil {
		return LoginResult{}, err
	}

	// hash token
	hashedToken := utility.HashToken(token)
	expiredTime := time.Now().Add(SessionDuration)

	// insert session + enforce max active sessions per user
	err = s.repo.CreateSession(ctx, &CreateSessionPayload{
		UserID:     account.UserID,
		ActiveRole: nil,
		TokenHash:  hashedToken,
		UserAgent:  in.UserAgent,
		IpAddress:  in.IpAddress,
		MaxActive:  MaxSessionsPerUser,
		ExpiresAt:  expiredTime,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		Token:     token,
		ExpiresAt: expiredTime,
	}, nil
}

func (s *AuthService) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]queries.FindActiveSessionsByUserIDRow, error) {
	sessions, err := s.repo.FindActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
	}

	hashedToken := utility.HashToken(rawToken)
	if err := s.repo.DeleteSessionByTokenHash(ctx, hashedToken); err != nil {
		slog.Error("Failed to delete session", "error", err)
		return err
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.DeleteAllSessionsByUserID(ctx, userID); err != nil {
		slog.Error("Failed to delete all sessions", "error", err)
		return err
	}
	return nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req *CreateUserRequest) error {
	hashedPassword, err := utility.HashPassword(req.Password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return err
	}

	err = s.repo.CreateUserCredential(ctx, &CreateUserCredentialPayload{
		DtoCreateUser:  req,
		HashedPassword: hashedPassword,
		Provider:       ProviderCredentials,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailDuplicate):
			return utility.NewConflictError("email sudah terdaftar", "AUTH_EMAIL_TAKEN")
		case errors.Is(err, ErrUsernameDuplicate):
			return utility.NewConflictError("username sudah digunakan", "AUTH_USERNAME_TAKEN")
		}
		slog.Error("Failed to insert user credential", "error", err)
		return err
	}

	return nil
}

func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*queries.FindUserByIDRow, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, utility.NewNotFoundError(ErrAccountNotFound.Error(), "AUTH_USER_NOT_FOUND")
	}
	return user, nil
}
