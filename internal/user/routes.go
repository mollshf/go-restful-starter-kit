package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/ums/internal/shared/queries"
	"github.com/mollshf/ums/internal/shared/web"
)

// This function wires the users module's handlers and constructs the shared
// user middleware. The returned web.UserMiddlewares is consumed by every
// other module to enforce authentication without importing this package.
func Routes(router *gin.RouterGroup, db *pgxpool.Pool, repo *queries.Queries, authMW web.AuthMiddlewares) {
	repository := NewRepository(db, repo)
	service := NewService(repository)
	handler := NewHandler(service)
	sessionMW := authMW.Session

	user := router.Group("/user")
	user.POST("/roles", sessionMW, web.Wrap(handler.CreateRole))
}
