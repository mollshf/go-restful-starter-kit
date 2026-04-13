package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mollshf/academic/internal/shared"
)

const ContextUserIDKey = "auth_user_id"

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

func SetUserID(c *gin.Context, id uuid.UUID) {
	c.Set(ContextUserIDKey, id)
}

// UserID mengembalikan (uuid, nil) jika sudah login.
func UserID(c *gin.Context) (uuid.UUID, *shared.APIError) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return uuid.Nil, shared.NewUnauthorizedError("Autentikasi diperlukan", "UNAUTHORIZED")
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, shared.NewUnauthorizedError("Sesi tidak valid", "INVALID_SESSION")
	}
	return id, nil
}

// RequirePermission memastikan user punya semua permission yang disebut (AND).
// Tanpa argumen permission = cukup cek sudah login.
func RequirePermission(checker PermissionChecker, names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, apiErr := UserID(c)
		if apiErr != nil {
			shared.Failed(c, apiErr)
			c.Abort()
			return
		}
		if len(names) == 0 {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		for _, name := range names {
			ok, err := checker.HasPermission(ctx, uid, name)
			if err != nil {
				shared.Failed(c, shared.NewInternalServerError("Gagal memeriksa permission", "PERMISSION_CHECK_FAILED"))
				c.Abort()
				return
			}
			if !ok {
				shared.Failed(c, shared.NewForbiddenError("Akses ditolak", "FORBIDDEN"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
