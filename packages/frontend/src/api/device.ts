import { get, post, put, patch, del } from './request'
import { Device, DeviceHealth, DeviceQueryParams, DeviceFormData, PaginatedData } from '@/types'

// 获取设备列表
export function getDevices(params: DeviceQueryParams): Promise<PaginatedData<Device>> {
  return get<PaginatedData<Device>>('/devices', params)
}

// 获取设备详情
export function getDevice(id: number): Promise<Device> {
  return get<Device>(`/devices/${id}`)
}

// 新增设备
export function createDevice(data: DeviceFormData): Promise<{ id: number }> {
  return post<{ id: number }>('/devices', data)
}

// 更新设备
export function updateDevice(id: number, data: Partial<DeviceFormData>): Promise<void> {
  return put<void>(`/devices/${id}`, data)
}

// 更新设备状态
export function updateDeviceStatus(id: number, status: string): Promise<void> {
  return patch<void>(`/devices/${id}/status`, { status })
}

// 获取设备健康状态
export function getDeviceHealth(id: number): Promise<DeviceHealth> {
  return get<DeviceHealth>(`/devices/${id}/health`)
}

// 删除设备
export function deleteDevice(id: number): Promise<void> {
  return del<void>(`/devices/${id}`)
}

// 设备遥测概览
export interface TelemetryMetricSummary {
  code: string
  name: string
  unit: string
  avg: number | null
  max: number | null
  min: number | null
  alerts: number
  hourly: { hour: string; avg: number }[]
}

export interface AlertEvent {
  id: number
  type: string
  level: string
  message: string
  status: string
  triggered_at: string
}

export interface TelemetrySummary {
  metrics: Record<string, TelemetryMetricSummary>
  online_rate: number
  alert_events: AlertEvent[]
}

export function getTelemetrySummary(id: number, from?: string, to?: string): Promise<TelemetrySummary> {
  return get<TelemetrySummary>(`/devices/${id}/telemetry-summary`, { from, to })
}

// 批量更新设备
export function batchUpdateDevices(deviceIds: number[], updates: Record<string, any>): Promise<{ affected: number }> {
  return post<{ affected: number }>('/devices/batch-update', { device_ids: deviceIds, updates })
}

// 批量删除设备
export function batchDeleteDevices(deviceIds: number[], reason?: string): Promise<{ deleted: number }> {
  return del<{ deleted: number }>('/devices/batch', { data: { device_ids: deviceIds, reason } })
}

// 批量下发命令
export interface BatchCommandRequest {
  target_type: 'greenhouse' | 'device_group' | 'devices'
  target_ids: number[]
  command_type: string
  payload: Record<string, any>
  remark?: string
}

export interface BatchCommandResult {
  device_id: number
  command_id?: number
  status: string
  message?: string
}

export function batchCommands(data: BatchCommandRequest): Promise<{ results: BatchCommandResult[] }> {
  return post<{ results: BatchCommandResult[] }>('/controls/batch-commands', data)
}