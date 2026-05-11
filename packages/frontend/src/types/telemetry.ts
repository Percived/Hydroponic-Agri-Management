import type { PaginatedResponse } from './api'

export type TelemetryQualityFlag = 'normal' | 'missing' | 'out_of_range' | 'device_offline'

export interface TelemetryRecord {
  id: number
  sensor_channel_id: number
  metric_code: string
  value: number
  raw_value?: number
  quality_flag: TelemetryQualityFlag
  collected_at: string
  ingested_at: string
  batch_id?: number
  created_at: string
}

export interface IngestTelemetryRequest {
  items: {
    sensor_channel_id: number
    metric_code: string
    value: number
    raw_value?: number
    quality_flag?: string
    collected_at: string
    batch_id?: number
  }[]
}

export interface TelemetryQueryParams {
  sensor_channel_id?: number | string
  metric_code?: string
  start_time?: string
  end_time?: string
  batch_id?: number
  quality_flag?: string
  page?: number
  page_size?: number
}

export interface TelemetryListResponse extends PaginatedResponse<TelemetryRecord> {}

export interface TelemetryLatestItem {
  sensor_channel_id: number
  metric_code: string
  value: number
  quality_flag: TelemetryQualityFlag
  collected_at: string
}

export interface TelemetryLatestBatchResponse {
  items: TelemetryLatestItem[]
}

export interface ChannelSnapshot {
  channel_id: number
  device_name: string
  device_code: string
  channel_code: string
  metric_code: string
  unit: string
  latest_value: number | null
  quality_flag: TelemetryQualityFlag
  collected_at: string
  status: 'ONLINE' | 'OFFLINE' | 'FAULT'
}

export interface TelemetrySSEEvent {
  schema_version: number
  sensor_channel_id: number
  metric_code: string
  value: number
  quality_flag?: TelemetryQualityFlag
  collected_at: string
  device_code: string
}
