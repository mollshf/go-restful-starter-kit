package web

import "github.com/gin-gonic/gin"

// AuthMiddlewares is the cross-module bag of authentication middleware.
//
// The struct lives here (in shared/web) so that any module can accept it
// without importing internal/auth — only the auth module knows how to
// construct its fields. This keeps the modulith rule "modules never import
// each other" intact while still letting auth own the DB-backed validation
// logic.
//
// Add fields here as more shared auth middleware are introduced
// (e.g. Permission, RoleCheck). Existing callers keep compiling.
type AuthMiddlewares struct {
	Session gin.HandlerFunc
}
