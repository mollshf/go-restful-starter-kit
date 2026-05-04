package auth

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type SessionResponse struct {
	ID         uuid.UUID  `json:"id"`
	IPAddress  netip.Addr `json:"ip_address,omitempty"`
	UserAgent  string     `json:"user_agent"`
	ActiveRole *uuid.UUID `json:"active_role,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at,omitempty"`
	IsCurrent  bool       `json:"is_current"`
}

type CreateUserRequest struct {
	Fullname string `json:"fullname" binding:"required,min=3,max=64"`
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type MessageOnlyResponse struct {
	Message string `json:"message"`
}

type CreateRoleRequest struct {
	RoleName string `json:"role_name" binding:"required"`
	RoleCode string `json:"role_code" binding:"required"`
}

type TstQueryStringQuery struct {
	Skip   string `json:"skip"`
	Limit  string `json:"limit"`
	Search string `json:"search"`
	Name   string `json:"name"`
}
