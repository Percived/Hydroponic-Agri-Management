import { get, post } from './request'
import { TelemetryPoint, TelemetryStats, TelemetryLatestParams, TelemetryHistoryParams, TelemetryStatsParams, PaginatedData } from '@/types'

// 获取最新遥测数据
export function getLatestTelemetry(params: TelemetryLatestParams): Promise<{ items: TelemetryPoint[] }> {
  return get<{ items: TelemetryPoint[] }>('/telemetry/latest', params as unknown as Record<string, string>)
}

// 获取历史遥测数据
export function getHistoryTelemetry(params: TelemetryHistoryParams): Promise<PaginatedData<TelemetryPoint>> {
  return get<PaginatedData<TelemetryPoint>>('/telemetry/history', params as unknown as Record<string, string>)
}

// 获取遥测统计
export function getTelemetryStats(params: TelemetryStatsParams): Promise<TelemetryStats> {
  return get<TelemetryStats>('/telemetry/stats', params as unknown as Record<string, string>)
}

// 上报遥测数据
export function postTelemetry(data: {
  device_code: string
  collected_at?: string
  metrics: Array<{ code: string; value: number; unit?: string }>
}): Promise<{ accepted: number }> {
  return post<{ accepted: number }>('/telemetry', data)
}