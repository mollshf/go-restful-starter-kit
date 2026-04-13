package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mollshf/academic/internal/shared"
)

type AuthService struct {
	*AuthRepository
}

func NewAuthService(authRepository *AuthRepository) (*AuthService, error) {
	if authRepository == nil {
		return nil, shared.NewInternalServerError("User service is required", "AUTH_UNKNOWN_ERROR")
	}
	return &AuthService{
		authRepository,
	}, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req *CreateUserRequest) (*uuid.UUID, error) {

	id, err := s.AuthRepository.InsertUser(ctx, &User{
		Fullname:      req.Fullname,
		Username:      req.Username,
		Email:         req.Email,
		VerifiedEmail: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DeletedAt:     time.Time{},
	})

	if err != nil {
		return nil, shared.NewBadRequestError("Failed to insert user", "AUTH_UNKNOWN_ERROR")
	}

	return id, nil

}
