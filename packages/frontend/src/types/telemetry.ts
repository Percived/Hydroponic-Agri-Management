import type { PaginatedResponse } from './api'

export interface TelemetryRecord {
  id: number
  sensor_channel_id: number
  metric_code: string
  value: number
  raw_value?: number
  quality_flag: 'normal' | 'outlier' | 'missing' | 'interpolated'
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
  sensor_channel_id?: number
  metric_code?: string
  start_time?: string
  end_time?: string
  batch_id?: number
  quality_flag?: string
  page?: number
  page_size?: number
}

export interface TelemetryListResponse extends PaginatedResponse<TelemetryRecord> {}
