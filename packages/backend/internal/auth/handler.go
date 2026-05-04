package auth

import (
	"net/http"
	"strconv"
	"strings"

	"hydroponic-backend/internal/platform/config"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	cfg config.AuthConfig
	db  *gorm.DB
}

func NewHandler(cfg config.AuthConfig, db *gorm.DB) *Handler {
	return &Handler{cfg: cfg, db: db}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var user User
	if err := h.db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "invalid_credentials", nil)
		return
	}
	if user.Status == UserStatusDisabled {
		response.Error(c, http.StatusForbidden, platformErrors.CodeForbidden, "user_disabled", nil)
		return
	}
	if !CheckPassword(user.PasswordHash, req.Password) {
		response.Error(c, http.StatusUnauthorized, platformErrors.CodeUnauthorized, "invalid_credentials", nil)
		return
	}

	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	token, err := GenerateToken(h.cfg.JWTSecret, user, roles, h.cfg.TokenExpireSecs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "token_error", nil)
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_in": h.cfg.TokenExpireSecs,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"roles":    roles,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, gin.H{})
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	keyword := strings.TrimSpace(c.Query("keyword"))

	query := h.db.Model(&User{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var users []User
	if total > 0 {
		if err := query.Preload("Roles").Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&users).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		roles := make([]string, 0, len(u.Roles))
		for _, r := range u.Roles {
			roles = append(roles, r.Name)
		}
		items = append(items, gin.H{
			"id":       u.ID,
			"username": u.Username,
			"roles":    roles,
			"status":   u.Status,
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var existing int64
	if err := h.db.Model(&User{}).Where("username = ?", req.Username).Count(&existing).Error; err == nil && existing > 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "username_exists", nil)
		return
	}

	roles, err := h.loadRoles(req.Roles)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_roles", gin.H{"errors": []gin.H{{"field": "roles", "reason": "not_found"}}})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "hash_failed", nil)
		return
	}

	user := User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Status:       UserStatusEnabled,
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Model(&user).Association("Roles").Replace(&roles)
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": user.ID})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID, err := parseID(c.Param("userId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}

	if len(updates) > 0 {
		if err := h.db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
	}

	if req.Roles != nil {
		if len(*req.Roles) == 0 {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_roles", gin.H{"errors": []gin.H{{"field": "roles", "reason": "empty"}}})
			return
		}
		roles, err := h.loadRoles(*req.Roles)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_roles", gin.H{"errors": []gin.H{{"field": "roles", "reason": "not_found"}}})
			return
		}
		var user User
		if err := h.db.First(&user, userID).Error; err != nil {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		if err := h.db.Model(&user).Association("Roles").Replace(&roles); err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
	}

	response.Success(c, gin.H{})
}

func (h *Handler) UpdateUserStatus(c *gin.Context) {
	userID, err := parseID(c.Param("userId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.db.Model(&User{}).Where("id = ?", userID).Update("status", req.Status).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ListRoles(c *gin.Context) {
	var roles []Role
	if err := h.db.Order("id asc").Find(&roles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	items := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		items = append(items, gin.H{"id": r.ID, "name": r.Name, "description": r.Description})
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	role := Role{Name: req.Name, Description: req.Description}
	if err := h.db.Create(&role).Error; err != nil {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "role_exists", nil)
		return
	}
	response.Success(c, gin.H{"id": role.ID})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	roleID, err := parseID(c.Param("roleId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.db.Model(&Role{}).Where("id = ?", roleID).Update("description", req.Description).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) loadRoles(names []string) ([]Role, error) {
	nameSet := map[string]struct{}{}
	uniq := make([]string, 0, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		if _, ok := nameSet[n]; ok {
			continue
		}
		nameSet[n] = struct{}{}
		uniq = append(uniq, n)
	}

	var roles []Role
	if err := h.db.Where("name IN ?", uniq).Find(&roles).Error; err != nil {
		return nil, err
	}
	if len(roles) != len(uniq) {
		return nil, gorm.ErrRecordNotFound
	}
	return roles, nil
}

func parsePage(c *gin.Context) (int, int) {
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}
	return page, pageSize
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

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}
