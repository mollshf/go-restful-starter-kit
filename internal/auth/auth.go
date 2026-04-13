package auth

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	IPAddress   string    `json:"ip_address"`
	TokenHashed string    `json:"token_hashed"`
	UserAgent   string    `json:"user_agent"`
	ActiveRole  string    `json:"active_role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
