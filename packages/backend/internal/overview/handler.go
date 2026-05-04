package overview

import (
	"net/http"

	alertpkg "hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/device"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Dashboard(c *gin.Context) {
	var devicesOnline int64
	if err := h.db.Model(&device.Device{}).
		Where("status = ?", device.DeviceStatusEnabled).
		Where("last_seen_at IS NOT NULL").
		Where("TIMESTAMPDIFF(SECOND, last_seen_at, UTC_TIMESTAMP()) <= sampling_interval_sec * 3").
		Count(&devicesOnline).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var totalDevices int64
	if err := h.db.Model(&device.Device{}).Count(&totalDevices).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var alertsOpen int64
	if err := h.db.Model(&alertpkg.Alert{}).Where("status = ?", alertpkg.StatusOpen).Count(&alertsOpen).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	devicesOffline := totalDevices - devicesOnline
	if devicesOffline < 0 {
		devicesOffline = 0
	}

	response.Success(c, gin.H{
		"devices_online":  devicesOnline,
		"devices_offline": devicesOffline,
		"alerts_open":     alertsOpen,
	})
}
