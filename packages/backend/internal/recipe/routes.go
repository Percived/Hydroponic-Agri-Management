package recipe

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	// NutrientRecipe routes
	recipes := r.Group("/recipes")
	recipes.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListRecipes)
	recipes.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateRecipe)
	recipes.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetRecipe)
	recipes.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateRecipe)
	recipes.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteRecipe)
	recipes.POST("/:id/publish", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.PublishRecipe)

	// Recipe targets (stage + ion)
	recipes.GET("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetRecipeTargets)
	recipes.PUT("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateRecipeTargets)

	// Recipe-to-batch binding
	recipes.POST("/:id/bind", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.BindRecipeToBatch)

	// BatchRecipeBinding routes
	bindings := r.Group("/recipe-bindings")
	bindings.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListBindings)
	bindings.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetBinding)
	bindings.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateBinding)
	bindings.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteBinding)
}
