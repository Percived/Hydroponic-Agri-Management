package audit

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)
	logs := r.Group("/audit-logs")
	logs.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.List)
}
