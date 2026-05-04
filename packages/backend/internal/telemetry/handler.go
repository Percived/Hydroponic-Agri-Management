package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	alertpkg "hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/control"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/config"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct {
	db     *gorm.DB
	influx influxdb2.Client
	mqtt   mqtt.Client
	cfg    config.InfluxConfig
	log    *slog.Logger
}

func NewHandler(db *gorm.DB, influx influxdb2.Client, mqttClient mqtt.Client, cfg config.InfluxConfig, log *slog.Logger) *Handler {
	return &Handler{db: db, influx: influx, mqtt: mqttClient, cfg: cfg, log: log}
}

func (h *Handler) Ingest(c *gin.Context) {
	var req IngestTelemetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var dev device.Device
	if err := h.db.Select("id", "device_code", "status").Where("device_code = ?", req.DeviceCode).First(&dev).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "device_not_found", nil)
		return
	}
	if dev.Status != device.DeviceStatusEnabled {
		response.Error(c, http.StatusConflict, platformErrors.CodeDeviceOffline, "device_disabled", nil)
		return
	}

	collectedAt := time.Now().UTC()
	if strings.TrimSpace(req.CollectedAt) != "" {
		t, err := time.Parse(time.RFC3339Nano, req.CollectedAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_collected_at", gin.H{"errors": []gin.H{{"field": "collected_at", "reason": "invalid_time"}}})
			return
		}
		collectedAt = t.UTC()
	}

	codes := make([]string, 0, len(req.Metrics))
	seen := map[string]struct{}{}
	for _, m := range req.Metrics {
		if _, ok := seen[m.Code]; ok {
			continue
		}
		seen[m.Code] = struct{}{}
		codes = append(codes, m.Code)
	}

	var metrics []Metric
	if err := h.db.Where("code IN ?", codes).Find(&metrics).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	metricByCode := map[string]Metric{}
	for _, m := range metrics {
		metricByCode[m.Code] = m
	}

	for _, code := range codes {
		if _, ok := metricByCode[code]; !ok {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_metric", gin.H{"errors": []gin.H{{"field": "metrics.code", "reason": fmt.Sprintf("%s_not_found", code)}}})
			return
		}
	}

	rows := make([]TelemetryData, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		meta := metricByCode[m.Code]
		raw := *m.Value
		quality := uint8(0)
		if (meta.MinValue != nil && raw < *meta.MinValue) || (meta.MaxValue != nil && raw > *meta.MaxValue) {
			quality = 1
		}
		value := raw
		rows = append(rows, TelemetryData{
			DeviceID:    dev.ID,
			MetricID:    meta.ID,
			Value:       value,
			RawValue:    &raw,
			Quality:     quality,
			CollectedAt: collectedAt,
		})
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		return tx.Model(&device.Device{}).Where("id = ?", dev.ID).Update("last_seen_at", collectedAt).Error
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "ingest_failed", nil)
		return
	}

	h.evaluateAndTrigger(dev, rows, metricByCode)
	h.writeInflux(c, dev.DeviceCode, rows, metricByCode)
	response.Success(c, gin.H{"accepted": len(rows)})
}

func (h *Handler) Latest(c *gin.Context) {
	deviceID, err := parseUint64Query(c, "device_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_device_id", nil)
		return
	}

	filterCodes := splitCSV(c.Query("metrics"))
	q := h.db.Table("telemetry_data t").
		Select("t.device_id, m.code AS metric_code, t.value, t.raw_value, t.quality, t.collected_at").
		Joins("JOIN metrics m ON m.id = t.metric_id").
		Joins("JOIN (SELECT metric_id, MAX(collected_at) max_collected_at FROM telemetry_data WHERE device_id = ? GROUP BY metric_id) latest ON latest.metric_id = t.metric_id AND latest.max_collected_at = t.collected_at", deviceID).
		Where("t.device_id = ?", deviceID)
	if len(filterCodes) > 0 {
		q = q.Where("m.code IN ?", filterCodes)
	}

	type row struct {
		DeviceID    uint64    `json:"device_id"`
		MetricCode  string    `json:"metric_code"`
		Value       float64   `json:"value"`
		RawValue    *float64  `json:"raw_value"`
		Quality     uint8     `json:"quality"`
		CollectedAt time.Time `json:"collected_at"`
	}
	rows := []row{}
	if err := q.Order("m.code ASC").Find(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		items = append(items, gin.H{
			"device_id":    r.DeviceID,
			"metric_code":  r.MetricCode,
			"value":        r.Value,
			"raw_value":    r.RawValue,
			"quality":      r.Quality,
			"collected_at": r.CollectedAt,
		})
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) History(c *gin.Context) {
	deviceID, err := parseUint64Query(c, "device_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_device_id", nil)
		return
	}
	metricCode := strings.TrimSpace(c.Query("metric_code"))
	if metricCode == "" {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "metric_code_required", nil)
		return
	}

	startTime, err := parseRFC3339(c.Query("start_time"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
		return
	}
	endTime, err := parseRFC3339(c.Query("end_time"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
		return
	}

	includeRaw := strings.EqualFold(c.DefaultQuery("include_raw", "false"), "true")
	page, pageSize := parsePage(c, 100, 2000)

	query := h.db.Table("telemetry_data t").
		Select("t.device_id, m.code AS metric_code, t.value, t.raw_value, t.quality, t.collected_at").
		Joins("JOIN metrics m ON m.id = t.metric_id").
		Where("t.device_id = ? AND m.code = ? AND t.collected_at >= ? AND t.collected_at <= ?", deviceID, metricCode, startTime, endTime)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	type row struct {
		DeviceID    uint64    `json:"device_id"`
		MetricCode  string    `json:"metric_code"`
		Value       float64   `json:"value"`
		RawValue    *float64  `json:"raw_value"`
		Quality     uint8     `json:"quality"`
		CollectedAt time.Time `json:"collected_at"`
	}
	rows := []row{}
	if total > 0 {
		if err := query.Order("t.collected_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		item := gin.H{
			"device_id":    r.DeviceID,
			"metric_code":  r.MetricCode,
			"value":        r.Value,
			"quality":      r.Quality,
			"collected_at": r.CollectedAt,
		}
		if includeRaw {
			item["raw_value"] = r.RawValue
		}
		items = append(items, item)
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) Stats(c *gin.Context) {
	deviceID, err := parseUint64Query(c, "device_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_device_id", nil)
		return
	}
	metricCode := strings.TrimSpace(c.Query("metric_code"))
	if metricCode == "" {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "metric_code_required", nil)
		return
	}
	startTime, err := parseRFC3339(c.Query("start_time"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
		return
	}
	endTime, err := parseRFC3339(c.Query("end_time"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
		return
	}

	type agg struct {
		Avg *float64 `json:"avg"`
		Max *float64 `json:"max"`
		Min *float64 `json:"min"`
	}
	res := agg{}
	err = h.db.Table("telemetry_data t").
		Select("AVG(t.value) AS avg, MAX(t.value) AS max, MIN(t.value) AS min").
		Joins("JOIN metrics m ON m.id = t.metric_id").
		Where("t.device_id = ? AND m.code = ? AND t.collected_at >= ? AND t.collected_at <= ? AND t.quality = 0", deviceID, metricCode, startTime, endTime).
		Scan(&res).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, gin.H{"avg": res.Avg, "max": res.Max, "min": res.Min})
}

func (h *Handler) SetRetention(c *gin.Context) {
	var req struct {
		KeepDays uint `json:"keep_days" binding:"required,min=7,max=3650"`
		Archive  bool `json:"archive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	entries := []SystemConfig{
		{ConfigKey: "retention_days", ConfigValue: strconv.FormatUint(uint64(req.KeepDays), 10), Description: "Telemetry retention days"},
		{ConfigKey: "retention_archive", ConfigValue: strconv.FormatBool(req.Archive), Description: "Archive expired telemetry"},
	}
	if err := h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "description", "updated_at"}),
	}).Create(&entries).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) writeInflux(c *gin.Context, deviceCode string, rows []TelemetryData, metricByCode map[string]Metric) {
	if h.influx == nil || h.cfg.Org == "" || h.cfg.Bucket == "" {
		return
	}

	metricByID := make(map[uint64]Metric, len(metricByCode))
	for _, m := range metricByCode {
		metricByID[m.ID] = m
	}

	points := make([]*write.Point, 0, len(rows))
	for _, r := range rows {
		meta, ok := metricByID[r.MetricID]
		if !ok {
			continue
		}
		fields := map[string]interface{}{
			"value":   r.Value,
			"quality": int(r.Quality),
		}
		if r.RawValue != nil {
			fields["raw_value"] = *r.RawValue
		}
		points = append(points, influxdb2.NewPoint(
			"telemetry",
			map[string]string{"device_code": deviceCode, "metric_code": meta.Code},
			fields,
			r.CollectedAt,
		))
	}
	if len(points) == 0 {
		return
	}

	writeCtx := context.Background()
	if c != nil && c.Request != nil {
		writeCtx = c.Request.Context()
	}
	if err := h.influx.WriteAPIBlocking(h.cfg.Org, h.cfg.Bucket).WritePoint(writeCtx, points...); err != nil {
		h.log.Warn("influx write failed", "error", err)
	}
}

func (h *Handler) evaluateAndTrigger(sourceDevice device.Device, rows []TelemetryData, metricByCode map[string]Metric) {
	if len(rows) == 0 {
		return
	}

	metricByID := map[uint64]Metric{}
	metricIDs := make([]uint64, 0, len(metricByCode))
	seenMetric := map[uint64]struct{}{}
	for _, m := range metricByCode {
		metricByID[m.ID] = m
		if _, ok := seenMetric[m.ID]; !ok {
			seenMetric[m.ID] = struct{}{}
			metricIDs = append(metricIDs, m.ID)
		}
	}

	var rules []control.ControlRule
	if err := h.db.Where("enabled = ? AND metric_id IN ?", true, metricIDs).Find(&rules).Error; err != nil {
		h.log.Warn("rule query failed", "error", err)
		return
	}
	if len(rules) == 0 {
		return
	}

	targetIDs := make([]uint64, 0, len(rules))
	seenTarget := map[uint64]struct{}{}
	for _, r := range rules {
		if _, ok := seenTarget[r.TargetDeviceID]; ok {
			continue
		}
		seenTarget[r.TargetDeviceID] = struct{}{}
		targetIDs = append(targetIDs, r.TargetDeviceID)
	}

	var targetDevices []device.Device
	if err := h.db.Select("id", "device_code", "status").Where("id IN ?", targetIDs).Find(&targetDevices).Error; err != nil {
		h.log.Warn("target device query failed", "error", err)
		return
	}
	targetByID := map[uint64]device.Device{}
	for _, d := range targetDevices {
		targetByID[d.ID] = d
	}

	type actionPayload struct {
		CommandType string                 `json:"command_type"`
		Payload     map[string]interface{} `json:"payload"`
	}

	for _, row := range rows {
		for _, rule := range rules {
			if rule.MetricID != row.MetricID {
				continue
			}
			if !compare(row.Value, rule.Operator, rule.Threshold) {
				continue
			}

			target, ok := targetByID[rule.TargetDeviceID]
			if !ok {
				continue
			}

			var action actionPayload
			if err := json.Unmarshal(rule.Action, &action); err != nil {
				h.log.Warn("rule action parse failed", "rule_id", rule.ID, "error", err)
				continue
			}
			if action.CommandType == "" {
				action.CommandType = "AUTO"
			}
			if action.Payload == nil {
				action.Payload = map[string]interface{}{}
			}
			actionBytes, _ := json.Marshal(action.Payload)

			now := time.Now().UTC()
			cmd := control.ControlCommand{
				DeviceID:    target.ID,
				CommandType: action.CommandType,
				Payload:     actionBytes,
				Status:      control.CommandStatusPending,
				CreatedBy:   1,
			}
			if err := h.db.Create(&cmd).Error; err != nil {
				h.log.Warn("auto command create failed", "rule_id", rule.ID, "error", err)
				continue
			}

			if target.Status == device.DeviceStatusEnabled && h.publishAutoCommand(target.DeviceCode, cmd.ID, action.CommandType, action.Payload) {
				_ = h.db.Model(&cmd).Updates(map[string]interface{}{"status": control.CommandStatusSent, "sent_at": now}).Error
			} else {
				_ = h.db.Model(&cmd).Update("status", control.CommandStatusFailed).Error
			}

			metricMeta := metricByID[row.MetricID]
			message := fmt.Sprintf("Rule %s triggered on %s: %.4f %s %.4f", rule.Name, metricMeta.Code, row.Value, rule.Operator, rule.Threshold)
			metricID := row.MetricID
			alertValue := row.Value
			alert := alertpkg.Alert{
				Type:        "THRESHOLD",
				Level:       "WARN",
				MetricID:    &metricID,
				DeviceID:    sourceDevice.ID,
				Value:       &alertValue,
				Message:     message,
				Status:      alertpkg.StatusOpen,
				TriggeredAt: now,
			}
			if err := h.db.Create(&alert).Error; err != nil {
				h.log.Warn("auto alert create failed", "rule_id", rule.ID, "error", err)
			}
		}
	}
}

func (h *Handler) publishAutoCommand(deviceCode string, commandID uint64, commandType string, payload map[string]interface{}) bool {
	if h.mqtt == nil || !h.mqtt.IsConnectionOpen() {
		return false
	}
	topic := fmt.Sprintf("hydroponic/v1/command/%s", deviceCode)
	msg := map[string]interface{}{
		"command_id":   commandID,
		"command_type": commandType,
		"payload":      payload,
		"created_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(msg)
	token := h.mqtt.Publish(topic, 1, false, b)
	if !token.WaitTimeout(3 * time.Second) {
		return false
	}
	return token.Error() == nil
}

func compare(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

func parseUint64Query(c *gin.Context, key string) (uint64, error) {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return 0, fmt.Errorf("%s required", key)
	}
	return strconv.ParseUint(v, 10, 64)
}

func parseRFC3339(v string) (time.Time, error) {
	if strings.TrimSpace(v) == "" {
		return time.Time{}, fmt.Errorf("time required")
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func parsePage(c *gin.Context, defaultSize int, maxSize int) (int, int) {
	page := 1
	if v := c.Query("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			page = i
		}
	}
	pageSize := defaultSize
	if v := c.Query("page_size"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			pageSize = i
		}
	}
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return page, pageSize
}

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
