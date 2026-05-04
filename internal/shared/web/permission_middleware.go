package web

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mollshf/starter-kit/internal/shared/utility"
)

const ContextUserIDKey = "auth_user_id"

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

func SetUserID(c *gin.Context, id uuid.UUID) {
	c.Set(ContextUserIDKey, id)
}

// UserID mengembalikan (uuid, nil) jika sudah login.
func UserID(c *gin.Context) (uuid.UUID, *utility.APIError) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return uuid.Nil, utility.NewUnauthorizedError("Autentikasi diperlukan", "UNAUTHORIZED")
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, utility.NewUnauthorizedError("Sesi tidak valid", "INVALID_SESSION")
	}
	return id, nil
}

// RequirePermission memastikan user punya semua permission yang disebut (AND).
// Tanpa argumen permission = cukup cek sudah login.
func RequirePermission(checker PermissionChecker, names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, apiErr := UserID(c)
		if apiErr != nil {
			Failed(c, apiErr)
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
				Failed(c, utility.NewInternalServerError("Gagal memeriksa permission", "PERMISSION_CHECK_FAILED"))
				c.Abort()
				return
			}
			if !ok {
				Failed(c, utility.NewForbiddenError("Akses ditolak", "FORBIDDEN"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
