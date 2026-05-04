// 设备类型
export enum DeviceType {
  SENSOR = 'SENSOR',
  ACTUATOR = 'ACTUATOR'
}

// 设备分类
export enum DeviceCategory {
  TEMP = 'TEMP',
  HUMIDITY = 'HUMIDITY',
  PH = 'PH',
  EC = 'EC',
  CO2 = 'CO2',
  LIGHT = 'LIGHT',
  FAN = 'FAN',
  PUMP = 'PUMP',
  VALVE = 'VALVE'
}

// 设备状态
export enum DeviceStatus {
  ENABLED = 'ENABLED',
  DISABLED = 'DISABLED'
}

// 通信协议
export enum DeviceProtocol {
  MQTT = 'MQTT',
  HTTP = 'HTTP'
}

// 设备
export interface Device {
  id: number
  device_code: string
  name: string
  type: DeviceType
  category: string
  greenhouse_id: number | null
  group_id: number | null
  status: DeviceStatus
  protocol: DeviceProtocol
  last_seen_at: string | null
  created_at: string
  updated_at: string
  sampling_interval_sec?: number
}

// 设备健康状态
export interface DeviceHealth {
  device_id: number
  online: boolean
  last_seen_at: string | null
}

// 设备分组
export interface DeviceGroup {
  id: number
  greenhouse_id: number
  name: string
  description: string | null
  created_at: string
  updated_at: string
  device_count?: number
}

// 设备查询参数
export interface DeviceQueryParams {
  page?: number
  page_size?: number
  type?: DeviceType
  category?: string
  group_id?: number
  greenhouse_id?: number
  status?: DeviceStatus
  keyword?: string
}

// 新增/编辑设备参数
export interface DeviceFormData {
  device_code?: string
  name: string
  type: DeviceType
  category: string
  protocol: DeviceProtocol
  greenhouse_id?: number | null
  group_id?: number | null
  sampling_interval_sec?: number
}