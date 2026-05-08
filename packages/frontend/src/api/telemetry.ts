import { get, post } from './request'
import type {
  TelemetryRecord,
  TelemetryListResponse,
  IngestTelemetryRequest,
  TelemetryQueryParams,
  TelemetryLatestBatchResponse
} from '@/types'

// 批量摄入遥测数据
export const ingestTelemetry = (data: IngestTelemetryRequest) =>
  post<{ accepted: number }>('/telemetry/ingest', data)

// 查询遥测数据
export const queryTelemetry = (params: TelemetryQueryParams) =>
  get<TelemetryListResponse>('/telemetry/query', params as Record<string, unknown>)

// 查询通道最新遥测值（无数据时静默处理，不弹错误提示）
export const getChannelLatest = (channelId: number) =>
  get<TelemetryRecord>(`/telemetry/channels/${channelId}/latest`, undefined, { silent: true } as any)

// 批量查询通道最新遥测值
export const getChannelsLatest = (channelIds: number[]) =>
  get<TelemetryLatestBatchResponse>('/telemetry/channels/latest', { ids: channelIds.join(',') }, { silent: true } as any)

// 查询通道遥测历史
export const getChannelHistory = (channelId: number, params?: TelemetryQueryParams) =>
  get<TelemetryListResponse>(`/telemetry/channels/${channelId}/history`, params as Record<string, unknown>)
