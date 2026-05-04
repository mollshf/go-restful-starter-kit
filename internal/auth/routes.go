package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/ums/internal/shared/queries"
	"github.com/mollshf/ums/internal/shared/web"
)

// AuthRoutes wires the auth module's handlers and constructs the shared
// auth middleware. The returned web.AuthMiddlewares is consumed by every
// other module to enforce authentication without importing this package.
func AuthRoutes(router *gin.RouterGroup, db *pgxpool.Pool, repo *queries.Queries) web.AuthMiddlewares {
	authRepository := NewAuthRepository(db, repo)
	authService, _ := NewAuthService(authRepository)
	authHandler := NewAuthHandler(authService)

	sessionMW := SessionMiddleware(authRepository)

	auth := router.Group("/auth")
	auth.POST("/register", web.Wrap(authHandler.RegisterUser))
	auth.POST("/login", web.Wrap(authHandler.Login))
	auth.GET("/me", sessionMW, web.Wrap(authHandler.Me))
	auth.GET("/sessions", sessionMW, web.Wrap(authHandler.ListActiveSessions))
	auth.DELETE("/logout", web.Wrap(authHandler.Logout))
	auth.DELETE("/logout-all", sessionMW, web.Wrap(authHandler.LogoutAll))

	return web.AuthMiddlewares{
		Session: sessionMW,
	}
}
