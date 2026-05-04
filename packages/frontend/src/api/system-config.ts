import { get, put } from './request'

export interface SystemConfigItem {
  id: number
  config_key: string
  config_value: string
  description: string
  updated_at: string
}

export interface SystemConfigListResponse {
  items: SystemConfigItem[]
}

export interface UpdateConfigRequest {
  config_key: string
  config_value: string
  description?: string
}

export function getSystemConfigs(): Promise<SystemConfigListResponse> {
  return get<SystemConfigListResponse>('/telemetry/system-configs')
}

export function updateSystemConfig(data: UpdateConfigRequest): Promise<{ id: number }> {
  return put<{ id: number }>('/telemetry/system-configs', data)
}
