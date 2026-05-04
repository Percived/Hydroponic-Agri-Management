package auth

import (
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.Config.Auth, deps.MySQL)

	authGroup := r.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.POST("/logout", AuthRequired(deps.Config.Auth, deps.MySQL), h.Logout)

	users := r.Group("/users")
	users.Use(AuthRequired(deps.Config.Auth, deps.MySQL, RoleAdmin))
	users.GET("", h.ListUsers)
	users.POST("", h.CreateUser)
	users.PUT("/:userId", h.UpdateUser)
	users.PATCH("/:userId/status", h.UpdateUserStatus)

	roles := r.Group("/roles")
	roles.Use(AuthRequired(deps.Config.Auth, deps.MySQL, RoleAdmin))
	roles.GET("", h.ListRoles)
	roles.POST("", h.CreateRole)
	roles.PUT("/:roleId", h.UpdateRole)
}
