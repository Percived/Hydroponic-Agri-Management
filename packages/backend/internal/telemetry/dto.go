package telemetry

// TelemetryRecordResponse is the JSON response for a telemetry record.
type TelemetryRecordResponse struct {
	ID              uint64   `json:"id"`
	SensorChannelID uint64   `json:"sensor_channel_id"`
	MetricCode      string   `json:"metric_code"`
	Value           float64  `json:"value"`
	RawValue        *float64 `json:"raw_value"`
	QualityFlag     string   `json:"quality_flag"`
	CollectedAt     string   `json:"collected_at"`
	IngestedAt      string   `json:"ingested_at"`
	BatchID         *uint64  `json:"batch_id"`
	CreatedAt       string   `json:"created_at"`
}

// IngestTelemetryRequest is the request for ingesting a telemetry record.
type IngestTelemetryRequest struct {
	SensorChannelID uint64   `json:"sensor_channel_id" binding:"required"`
	MetricCode      string   `json:"metric_code" binding:"required,min=1,max=32"`
	Value           float64  `json:"value" binding:"required"`
	RawValue        *float64 `json:"raw_value"`
	QualityFlag     string   `json:"quality_flag"`
	CollectedAt     string   `json:"collected_at" binding:"required"`
	BatchID         *uint64  `json:"batch_id"`
}

// IngestTelemetryBatchRequest supports single or batch ingestion.
type IngestTelemetryBatchRequest struct {
	Items []IngestTelemetryRequest `json:"items" binding:"required,min=1,max=200,dive"`
}

// TelemetryQuery holds query parameters for listing telemetry records.
type TelemetryQuery struct {
	SensorChannelID *uint64 `json:"sensor_channel_id" form:"sensor_channel_id"`
	MetricCode      string  `json:"metric_code" form:"metric_code"`
	StartTime       string  `json:"start_time" form:"start_time"`
	EndTime         string  `json:"end_time" form:"end_time"`
	BatchID         *uint64 `json:"batch_id" form:"batch_id"`
	QualityFlag     string  `json:"quality_flag" form:"quality_flag"`
	Page            int     `json:"page" form:"page"`
	PageSize        int     `json:"page_size" form:"page_size"`
}

// TelemetryListResponse is the paginated list response.
type TelemetryListResponse struct {
	Items    []TelemetryRecordResponse `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// LatestRecordResponse is the per-channel item in a batch latest response.
type LatestRecordResponse struct {
	SensorChannelID uint64  `json:"sensor_channel_id"`
	MetricCode      string  `json:"metric_code"`
	Value           float64 `json:"value"`
	QualityFlag     string  `json:"quality_flag"`
	CollectedAt     string  `json:"collected_at"`
}

// LatestBatchResponse wraps a batch latest query response.
type LatestBatchResponse struct {
	Items []LatestRecordResponse `json:"items"`
}
