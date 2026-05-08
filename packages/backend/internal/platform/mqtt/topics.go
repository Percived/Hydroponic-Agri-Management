package mqtt

const (
	// TopicPrefix is the base prefix for all MQTT topics
	TopicPrefix = "hydroponic"

	// Inbound topics (device → backend)
	TopicTelemetry   = "telemetry"
	TopicStatus      = "status"
	TopicHeartbeat   = "heartbeat"
	TopicErrors      = "errors"
	TopicDiagnostics = "diagnostics"
	TopicAck         = "ack"

	// Outbound topics (backend → device)
	TopicCmdPrefix = "cmd"
)
