package metric

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	mg := r.Group("/metrics")
	mg.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListMetrics)
	mg.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetMetric)
}
