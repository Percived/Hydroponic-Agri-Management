package review

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	reviews := r.Group("/reviews")

	// Write: Admin, Operator
	reviews.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateSnapshot)
	reviews.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateSnapshot)
	reviews.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteSnapshot)
	reviews.POST("/generate", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.GenerateReview)

	// Read: All roles
	reviews.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListSnapshots)
	reviews.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetSnapshot)

	// Batch-scoped snapshots
	reviews.GET("/batches/:batchId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetSnapshotsByBatch)
}
