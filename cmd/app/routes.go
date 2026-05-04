package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mollshf/starter-kit/internal/auth"
	"github.com/mollshf/starter-kit/internal/shared/queries"
)

func ApiRoutes(router *gin.Engine, db *pgxpool.Pool, repo *queries.Queries) {
	api := router.Group("/api")

	// auth must be wired first — its returned middleware bag (session,
	// permission, ...) is consumed by every other module to enforce auth.
	authMW := auth.AuthRoutes(api, db, repo)
	_ = authMW // pass to other module routes once they exist:
	// akademik.AkademikRoutes(api, db, repo, authMW)
	// keuangan.KeuanganRoutes(api, db, repo, authMW)
}
