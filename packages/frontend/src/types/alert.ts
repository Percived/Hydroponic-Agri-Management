import type { PaginatedResponse } from './api'

export type AlertType = 'THRESHOLD' | 'DEVICE_OFFLINE' | 'SYSTEM'
export type AlertLevel = 'INFO' | 'WARN' | 'CRITICAL'
export type AlertStatus = 'OPEN' | 'ACKNOWLEDGED' | 'RESOLVED' | 'IGNORED'

export interface Alert {
  schema_version?: number
  id: number
  device_code?: string
  type: AlertType
  level: AlertLevel
  metric_code?: string
  sensor_channel_id?: number
  actuator_channel_id?: number
  batch_id?: number
  trigger_value?: number
  message: string
  status: AlertStatus
  triggered_at: string
  resolved_at?: string
  resolved_by?: number
  created_at: string
  updated_at: string
  timeline_count?: number
}

export interface AlertTimelineEvent {
  id: number
  alert_id: number
  event_type: string
  event_source: 'SYSTEM' | 'MANUAL'
  operator_id?: number
  comment?: string
  event_payload?: Record<string, unknown>
  event_time: string
  created_at: string
}

export interface CreateAlertRequest {
  type: AlertType
  level: AlertLevel
  metric_code?: string
  sensor_channel_id?: number
  actuator_channel_id?: number
  trigger_value?: number
  message: string
  triggered_at: string
}

export interface UpdateAlertStatusRequest {
  status: AlertStatus
  resolved_at?: string
  comment?: string
}

export interface CreateTimelineEventRequest {
  event_type: string
  event_source: 'SYSTEM' | 'MANUAL'
  operator_id?: number
  comment?: string
  event_payload?: Record<string, unknown>
  event_time: string
}

export interface AlertListResponse extends PaginatedResponse<Alert> {}

export interface AlertStats {
  open_count: number
  acknowledged_count: number
  resolved_count: number
  ignored_count: number
  critical_count: number
  warn_count: number
  info_count: number
}
