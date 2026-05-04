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