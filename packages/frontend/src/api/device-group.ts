import { get, post, put, del } from './request'
import { DeviceGroup } from '@/types'

// 获取分组列表
export function getDeviceGroups(greenhouseId?: number): Promise<{ items: DeviceGroup[] }> {
  return get<{ items: DeviceGroup[] }>('/device-groups', greenhouseId ? { greenhouse_id: greenhouseId } : undefined)
}

// 获取分组详情
export function getDeviceGroup(id: number): Promise<DeviceGroup> {
  return get<DeviceGroup>(`/device-groups/${id}`)
}

// 新增分组
export function createDeviceGroup(data: { greenhouse_id: number; name: string; description?: string }): Promise<{ id: number }> {
  return post<{ id: number }>('/device-groups', data)
}

// 更新分组
export function updateDeviceGroup(id: number, data: { name?: string; description?: string }): Promise<void> {
  return put<void>(`/device-groups/${id}`, data)
}

// 删除分组
export function deleteDeviceGroup(id: number): Promise<void> {
  return del<void>(`/device-groups/${id}`)
}

// 添加设备到分组
export function addDeviceToGroup(groupId: number, deviceId: number): Promise<void> {
  return post<void>(`/device-groups/${groupId}/devices/${deviceId}`)
}

// 从分组移除设备
export function removeDeviceFromGroup(groupId: number, deviceId: number): Promise<void> {
  return del<void>(`/device-groups/${groupId}/devices/${deviceId}`)
}
