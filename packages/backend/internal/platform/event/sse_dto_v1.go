package event

type TelemetrySSEDataV1 struct {
	SchemaVersion   int     `json:"schema_version"`
	SensorChannelID uint64  `json:"sensor_channel_id"`
	MetricCode      string  `json:"metric_code"`
	Value           float64 `json:"value"`
	QualityFlag     string  `json:"quality_flag"`
	CollectedAt     string  `json:"collected_at"`
	DeviceCode      string  `json:"device_code"`
}

type DeviceStatusSSEDataV1 struct {
	SchemaVersion int    `json:"schema_version"`
	DeviceCode    string `json:"device_code"`
	Status        string `json:"status"`
	Reason        string `json:"reason,omitempty"`
	ReportedAt    string `json:"reported_at"`
}

type CommandDispatchedSSEDataV1 struct {
	SchemaVersion int    `json:"schema_version"`
	CommandID     uint64 `json:"command_id"`
	DeviceCode    string `json:"device_code"`
	Status        string `json:"status"`
	DispatchedAt  string `json:"dispatched_at"`
	SourceType    string `json:"source_type"`
	SourceID      uint64 `json:"source_id,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}
