package auth

import (
	"net/http"
	"strings"

	"hydroponic-backend/internal/platform/config"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	CtxUserID   = "user_id"
	CtxUsername = "username"
	CtxRoles    = "roles"
)

func AuthRequired(cfg config.AuthConfig, db *gorm.DB, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := ParseToken(cfg.JWTSecret, tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		if claims.UserID == 0 {
			response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		var user User
		if err := db.Select("id", "status").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}
		if user.Status == UserStatusDisabled {
			response.Error(c, http.StatusForbidden, platformErrors.CodeForbidden, "user_disabled", nil)
			c.Abort()
			return
		}

		if len(roles) > 0 && !hasAnyRole(claims.Roles, roles) {
			response.Error(c, http.StatusForbidden, platformErrors.CodeForbidden, "forbidden", nil)
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxRoles, claims.Roles)
		c.Next()
	}
}

func hasAnyRole(userRoles []string, required []string) bool {
	roleSet := map[string]struct{}{}
	for _, r := range userRoles {
		roleSet[r] = struct{}{}
	}
	for _, r := range required {
		if _, ok := roleSet[r]; ok {
			return true
		}
	}
	return false
}
