package auth

import (
	"context"

	"github.com/mollshf/academic/internal/shared"
)

type UserService interface {
	CheckUserExists(ctx context.Context, email string) (bool, error)
}

type Options struct {
	*AuthRepository
	UserService
}

type AuthService struct {
	options *Options
}

func NewAuthService(opts *Options) (*AuthService, error) {
	if opts.UserService == nil {
		return nil, shared.NewInternalServerError("User service is required", "AUTH_UNKNOWN_ERROR")
	}
	return &AuthService{
		options: opts,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	exists, err := s.options.UserService.CheckUserExists(ctx, req.Email)

	if err != nil {
		return nil, shared.NewInternalServerError("Failed to check user exists", "AUTH_UNKNOWN_ERROR")
	}

	if !exists {
		return nil, shared.NewNotFoundError("User not found", "USER_NOT_FOUND")
	}

	return &LoginResponse{
		CsrToken: "csr_token",
		Roles:    []string{"user"},
	}, nil

}
