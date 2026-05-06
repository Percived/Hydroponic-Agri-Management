import { get, post } from './request'
import type {
  TelemetryRecord,
  TelemetryListResponse,
  IngestTelemetryRequest,
  TelemetryQueryParams
} from '@/types'

// 批量摄入遥测数据
export const ingestTelemetry = (data: IngestTelemetryRequest) =>
  post<{ accepted: number }>('/telemetry/ingest', data)

// 查询遥测数据
export const queryTelemetry = (params: TelemetryQueryParams) =>
  get<TelemetryListResponse>('/telemetry/query', params as Record<string, unknown>)

// 查询通道最新遥测值
export const getChannelLatest = (channelId: number) =>
  get<TelemetryRecord>(`/telemetry/channels/${channelId}/latest`)

// 查询通道遥测历史
export const getChannelHistory = (channelId: number, params?: TelemetryQueryParams) =>
  get<TelemetryListResponse>(`/telemetry/channels/${channelId}/history`, params as Record<string, unknown>)
