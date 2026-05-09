package main

import "encoding/json"

// ──────────────────── MQTT 主题常量 ────────────────────

const (
	topicPrefix    = "hydroponic"
	topicTelemetry = "telemetry"
	topicHeartbeat = "heartbeat"
	topicStatus    = "status"
	topicErrors    = "errors"
	topicAck       = "ack"
	topicCmdPrefix = "cmd"
)

// ──────────────────── 执行器类型常量 ────────────────────

const (
	ActuatorPUMP          = "PUMP"
	ActuatorAERATOR       = "AERATOR"
	ActuatorFAN           = "FAN"
	ActuatorVALVE         = "VALVE"
	ActuatorSHADE         = "SHADE"
	ActuatorLED           = "LED"
	ActuatorHEATER        = "HEATER"
	ActuatorCO2Gen        = "CO2_GEN"
	ActuatorFOGGER        = "FOGGER"
	ActuatorDOSING_PUMP   = "DOSING_PUMP"
	ActuatorCHILLER       = "CHILLER"
	ActuatorSTIRRER       = "STIRRER"
	ActuatorDEHUMIDIFIER  = "DEHUMIDIFIER"
	ActuatorDAMPER        = "DAMPER"
	ActuatorUV_STERILIZER = "UV_STERILIZER"
	ActuatorOZONE_GEN     = "OZONE_GENERATOR"
	ActuatorFILTER        = "FILTER"
	ActuatorRO_SYSTEM     = "RO_SYSTEM"
	ActuatorTOP_UP_VALVE  = "TOP_UP_VALVE"
	ActuatorALARM         = "ALARM"
	ActuatorCALIB_VALVE   = "CALIBRATION_VALVE"
)

// ──────────────────── API 请求/响应结构体 ────────────────────

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type loginData struct {
	Token string `json:"token"`
}

type registerChannelItem struct {
	ChannelCode         string   `json:"channel_code"`
	MetricCode          string   `json:"metric_code"`
	Unit                string   `json:"unit"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec uint     `json:"sampling_interval_sec"`
	ActuatorType        string   `json:"actuator_type"`
	RatedPowerWatt      *float64 `json:"rated_power_watt"`
}

type registerDeviceReq struct {
	DeviceCode      string                `json:"device_code"`
	Name            string                `json:"name"`
	Model           string                `json:"model"`
	FirmwareVersion string                `json:"firmware_version"`
	GreenhouseID    uint64                `json:"greenhouse_id"`
	Protocol        string                `json:"protocol"`
	DeviceType      string                `json:"device_type"`
	Channels        []registerChannelItem `json:"channels"`
}

type registerDeviceResp struct {
	DeviceID   uint64   `json:"device_id"`
	ChannelIDs []uint64 `json:"channel_ids"`
}

// ──────────────── 传感器设备自发现响应 ────────────────

type sensorDeviceSelfResp struct {
	DeviceType string                `json:"device_type"`
	Device     sensorDeviceDetail    `json:"device"`
	Channels   []sensorChannelDetail `json:"channels"`
}

type sensorDeviceDetail struct {
	ID         uint64 `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}

type sensorChannelDetail struct {
	ID         uint64   `json:"id"`
	MetricCode string   `json:"metric_code"`
	Unit       string   `json:"unit"`
	RangeMin   *float64 `json:"range_min"`
	RangeMax   *float64 `json:"range_max"`
}

// ──────────────── 执行器设备自发现响应 ────────────────

type actuatorDeviceSelfResp struct {
	DeviceType string                  `json:"device_type"`
	Device     actuatorDeviceDetail    `json:"device"`
	Channels   []actuatorChannelDetail `json:"channels"`
}

type actuatorDeviceDetail struct {
	ID         uint64 `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}

type actuatorChannelDetail struct {
	ID           uint64   `json:"id"`
	ChannelCode  string   `json:"channel_code"`
	ActuatorType string   `json:"actuator_type"`
	CurrentState string   `json:"current_state"`
	RatedPower   *float64 `json:"rated_power_watt"`
}

// ──────────────── MQTT 遥测 / 心跳 / ACK 结构体 ────────────────

type telemetryItem struct {
	SensorChannelID uint64   `json:"sensor_channel_id"`
	MetricCode      string   `json:"metric_code"`
	Value           float64  `json:"value"`
	RawValue        *float64 `json:"raw_value,omitempty"`
	QualityFlag     string   `json:"quality_flag,omitempty"`
	CollectedAt     string   `json:"collected_at"`
}

type heartbeatPayload struct {
	TS string `json:"ts"`
}

type statusPayload struct {
	Status string `json:"status"`
}

type ackPayload struct {
	CommandID  uint64 `json:"command_id"`
	AckCode    string `json:"ack_code"`
	AckMessage string `json:"ack_message"`
}

type incomingCmd struct {
	CommandType string          `json:"command_type"`
	Payload     json.RawMessage `json:"payload"`
	CommandID   uint64          `json:"command_id,omitempty"`

	// Parsed from wrapped payload (injected by backend dispatchMQTT)
	InternalCommandID   uint64 `json:"_command_id,omitempty"`
	InternalCommandType string `json:"_command_type,omitempty"`
}

// ──────────────────── 传感器度量配置 ────────────────────

type metricConfig struct {
	Code    string
	Unit    string
	Base    float64
	Range   float64
	Anomaly float64
}

var defaultMetrics = []metricConfig{
	{Code: "TEMP", Unit: "°C", Base: 25, Range: 5, Anomaly: 40},
	{Code: "HUMIDITY", Unit: "%", Base: 70, Range: 15, Anomaly: 20},
	{Code: "PH", Unit: "pH", Base: 6.0, Range: 0.8, Anomaly: 3.5},
	{Code: "EC", Unit: "mS/cm", Base: 2.0, Range: 0.6, Anomaly: 5.5},
	{Code: "CO2", Unit: "ppm", Base: 800, Range: 300, Anomaly: 150},
	{Code: "LIGHT", Unit: "lx", Base: 35000, Range: 35000, Anomaly: 500},
	{Code: "WATER_TEMP", Unit: "°C", Base: 22, Range: 3, Anomaly: 35},
	{Code: "DO", Unit: "mg/L", Base: 6.5, Range: 1.5, Anomaly: 2.0},
	{Code: "LEVEL", Unit: "cm", Base: 50, Range: 20, Anomaly: 40},
	{Code: "ORP", Unit: "mV", Base: 350, Range: 150, Anomaly: 800},
	{Code: "TDS", Unit: "ppm", Base: 800, Range: 400, Anomaly: 5000},
	{Code: "O3", Unit: "ppb", Base: 20, Range: 15, Anomaly: 200},
	{Code: "TURBIDITY", Unit: "NTU", Base: 5, Range: 10, Anomaly: 100},
	{Code: "FLOW_RATE", Unit: "L/min", Base: 10, Range: 5, Anomaly: 50},
}

// ──────────────── 命令 payload 中的 state/value 结构 ────────────────

type cmdPayloadState struct {
	State string  `json:"state"`
	Value float64 `json:"value"`
}
