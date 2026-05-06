package nutrient

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	// NutrientTank routes
	tanks := r.Group("/nutrient-tanks")
	tanks.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListTanks)
	tanks.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateTank)
	tanks.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetTank)
	tanks.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateTank)
	tanks.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteTank)

	// SolutionChangeEvent routes
	solutions := r.Group("/solution-changes")
	solutions.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListSolutionChanges)
	solutions.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateSolutionChange)
	solutions.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetSolutionChange)

	// IonTestRecord routes
	ionTests := r.Group("/ion-tests")
	ionTests.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListIonTests)
	ionTests.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateIonTest)
	ionTests.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetIonTest)
	ionTests.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateIonTest)
	ionTests.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteIonTest)

	// NutrientConcentrateInventory routes
	inventory := r.Group("/concentrate-inventory")
	inventory.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListConcentrateInventory)
	inventory.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateConcentrateInventory)
	inventory.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetConcentrateInventory)
	inventory.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateConcentrateInventory)
	inventory.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteConcentrateInventory)

	// ConcentrateUsageLog routes
	usageLogs := r.Group("/concentrate-usage-logs")
	usageLogs.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListUsageLogs)
	usageLogs.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateUsageLog)
}
