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
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		} else if q := c.Query("token"); q != "" {
			// Support token via query param for SSE EventSource (browsers cannot set headers)
			tokenString = strings.TrimSpace(q)
		} else {
			response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}
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
