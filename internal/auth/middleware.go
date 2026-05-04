package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mollshf/ums/internal/shared/utility"
	"github.com/mollshf/ums/internal/shared/web"
)

const (
	ContextUserIDKey     = "user_id"
	ContextActiveRoleKey = "active_role"
	ContextSessionIDKey  = "session_id"
)

// SessionMiddleware validates the auth cookie against user_sessions:
//   - reads the cookie
//   - hashes the raw token
//   - looks up the session row
//   - rejects expired sessions
//   - sets user_id and active_role on the context for downstream handlers
func SessionMiddleware(repo *AuthRepository) gin.HandlerFunc {
	return web.Wrap(func(c *gin.Context) error {
		rawToken, err := c.Cookie(web.AuthCookieName)
		if err != nil || rawToken == "" {
			c.Abort()
			return utility.NewUnauthorizedError("Sesi tidak ditemukan", "AUTH_SESSION_NOT_FOUND")
		}

		hashed := utility.HashToken(rawToken)
		session, err := repo.FindSessionByTokenHash(c.Request.Context(), hashed)
		if err != nil {
			c.Abort()
			return err
		}
		if session == nil {
			c.Abort()
			return utility.NewUnauthorizedError("Sesi tidak valid", "AUTH_SESSION_INVALID")
		}

		if session.ExpiresAt.Before(time.Now()) {
			c.Abort()
			return utility.NewUnauthorizedError("Sesi telah berakhir", "AUTH_SESSION_EXPIRED")
		}

		c.Set(ContextUserIDKey, session.UserID)
		c.Set(ContextActiveRoleKey, session.ActiveRole)
		c.Set(ContextSessionIDKey, session.ID)
		return nil
	})
}
