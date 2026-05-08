package http

import (
	"fmt"
	"net/http"
	"strconv"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 200
)

func ParseID(c *gin.Context, key string) (uint64, error) {
	v := c.Param(key)
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", key, v)
	}
	return id, nil
}

func ParsePage(c *gin.Context) (int, int) {
	page := parseInt(c.Query("page"), DefaultPage)
	pageSize := parseInt(c.Query("page_size"), DefaultPageSize)
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if page < 1 {
		page = DefaultPage
	}
	return page, pageSize
}

func CurrentUserID(c *gin.Context) uint64 {
	v, ok := c.Get("user_id")
	if !ok {
		return 0
	}
	id, ok := v.(uint64)
	if !ok {
		return 0
	}
	return id
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func EnsureOneRowAffected(result *gorm.DB, resource string) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s not found", resource)
	}
	return nil
}

func CheckOneRowAffected(c *gin.Context, result *gorm.DB, resource string) bool {
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, resource+"_failed", nil)
		return false
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, resource+"_not_found", nil)
		return false
	}
	return true
}
