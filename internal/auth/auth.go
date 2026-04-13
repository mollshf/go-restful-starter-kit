package auth

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID                uuid.UUID `db:"id" json:"id"`
	UserId            uuid.UUID `json:"user_id"`
	PasswordHash      string    `json:"password_hash"`
	Provider          string    `json:"provider"`
	ProviderAccountID string    `json:"provider_account_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Session struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ActiveRole uuid.UUID `json:"active_role"`
	TokenHash  string    `json:"token_hash"`
	CsrfToken  string    `json:"csrf_token"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiredAt  time.Time `json:"expired_at"`
}

type User struct {
	ID            uuid.UUID `json:"id"`
	Fullname      string    `json:"fullname"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	VerifiedEmail bool      `json:"verified_email"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     time.Time `json:"deleted_at"`
}
