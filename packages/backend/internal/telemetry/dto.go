package telemetry

type IngestTelemetryRequest struct {
	DeviceCode  string                  `json:"device_code" binding:"required,min=1,max=64"`
	CollectedAt string                  `json:"collected_at" binding:"omitempty,max=64"`
	Metrics     []IngestTelemetryMetric `json:"metrics" binding:"required,min=1,max=50,dive"`
}

type IngestTelemetryMetric struct {
	Code  string   `json:"code" binding:"required,min=1,max=32"`
	Value *float64 `json:"value" binding:"required"`
	Unit  string   `json:"unit" binding:"omitempty,max=16"`
}
