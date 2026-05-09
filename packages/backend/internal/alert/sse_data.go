package alert

func BuildAlertSSEDataV1(a Alert, deviceCode string, timelineCount int64) map[string]interface{} {
	out := map[string]interface{}{
		"schema_version":      1,
		"id":                  a.ID,
		"type":                a.Type,
		"level":               a.Level,
		"metric_code":         a.MetricCode,
		"sensor_channel_id":   a.SensorChannelID,
		"actuator_channel_id": a.ActuatorChannelID,
		"batch_id":            a.BatchID,
		"trigger_value":       a.TriggerValue,
		"message":             a.Message,
		"status":              a.Status,
		"triggered_at":        timeToStr(a.TriggeredAt),
		"resolved_at":         timePtrToStr(a.ResolvedAt),
		"resolved_by":         a.ResolvedBy,
		"timeline_count":      timelineCount,
		"created_at":          timeToStr(a.CreatedAt),
		"updated_at":          timeToStr(a.UpdatedAt),
	}
	if deviceCode != "" {
		out["device_code"] = deviceCode
	}
	return out
}
