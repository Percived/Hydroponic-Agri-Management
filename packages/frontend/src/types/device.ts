import type { PaginatedResponse } from './api'

// ── Sensor Devices ──
export interface SensorDevice {
  id: number
  greenhouse_id: number
  growing_zone_id?: number
  device_code: string
  name: string
  model?: string
  firmware_version?: string
  status: 'ONLINE' | 'OFFLINE' | 'FAULT'
  last_seen_at?: string
  protocol: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateSensorDeviceRequest {
  greenhouse_id: number
  growing_zone_id?: number
  device_code: string
  name: string
  model?: string
  firmware_version?: string
  protocol?: string
  metadata?: Record<string, unknown>
}

export interface UpdateSensorDeviceRequest {
  name?: string
  model?: string
  firmware_version?: string
  status?: string
  growing_zone_id?: number
  metadata?: Record<string, unknown>
}

export interface SensorDeviceListResponse extends PaginatedResponse<SensorDevice> {}

// ── Sensor Channels ──
export interface SensorChannel {
  id: number
  sensor_device_id: number
  channel_code: string
  metric_code: string
  unit: string
  precision_digits: number
  range_min?: number
  range_max?: number
  sampling_interval_sec: number
  enabled: number
  last_reported_at?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateSensorChannelRequest {
  sensor_device_id: number
  channel_code: string
  metric_code: string
  unit: string
  precision_digits?: number
  range_min?: number
  range_max?: number
  sampling_interval_sec?: number
  metadata?: Record<string, unknown>
}

export interface UpdateSensorChannelRequest {
  channel_code?: string
  metric_code?: string
  unit?: string
  precision_digits?: number
  range_min?: number
  range_max?: number
  sampling_interval_sec?: number
  enabled?: number
  metadata?: Record<string, unknown>
}

export interface SensorChannelListResponse extends PaginatedResponse<SensorChannel> {}

// ── Actuator Devices ──
export interface ActuatorDevice {
  id: number
  greenhouse_id: number
  growing_zone_id?: number
  device_code: string
  name: string
  model?: string
  firmware_version?: string
  status: 'ONLINE' | 'OFFLINE' | 'FAULT'
  last_seen_at?: string
  protocol: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateActuatorDeviceRequest {
  greenhouse_id: number
  growing_zone_id?: number
  device_code: string
  name: string
  model?: string
  firmware_version?: string
  protocol?: string
  metadata?: Record<string, unknown>
}

export interface UpdateActuatorDeviceRequest {
  name?: string
  model?: string
  firmware_version?: string
  status?: string
  growing_zone_id?: number
  metadata?: Record<string, unknown>
}

export interface ActuatorDeviceListResponse extends PaginatedResponse<ActuatorDevice> {}

// ── Actuator Channels ──
export type ActuatorType = 'PUMP' | 'AERATOR' | 'FAN' | 'VALVE' | 'SHADE' | 'LED' | 'HEATER' | 'CO2_GEN' | 'FOGGER'

export interface ActuatorChannel {
  id: number
  actuator_device_id: number
  channel_code: string
  actuator_type: ActuatorType
  current_state: string
  rated_power_watt?: number
  enabled: number
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateActuatorChannelRequest {
  actuator_device_id: number
  channel_code: string
  actuator_type: ActuatorType
  rated_power_watt?: number
  metadata?: Record<string, unknown>
}

export interface UpdateActuatorChannelRequest {
  channel_code?: string
  actuator_type?: ActuatorType
  rated_power_watt?: number
  enabled?: number
  current_state?: string
  metadata?: Record<string, unknown>
}

export interface ActuatorChannelListResponse extends PaginatedResponse<ActuatorChannel> {}
