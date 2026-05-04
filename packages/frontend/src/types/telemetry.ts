// 遥测数据点
export interface TelemetryPoint {
  device_id: number
  metric_code: string
  value: number
  raw_value: number
  quality: number
  collected_at: string
}

// 遥测统计
export interface TelemetryStats {
  avg: number
  max: number
  min: number
}

// 最新遥测查询参数
export interface TelemetryLatestParams {
  device_id: number
  metrics?: string
}

// 历史遥测查询参数
export interface TelemetryHistoryParams {
  device_id: number
  metric_code: string
  start_time: string
  end_time: string
  include_raw?: boolean
  page?: number
  page_size?: number
}

// 统计查询参数
export interface TelemetryStatsParams {
  device_id: number
  metric_code: string
  start_time: string
  end_time: string
}

// 指标枚举
export enum MetricCode {
  TEMP = 'TEMP',
  HUMIDITY = 'HUMIDITY',
  PH = 'PH',
  EC = 'EC',
  CO2 = 'CO2',
  LIGHT = 'LIGHT',
  WATER_LEVEL = 'WATER_LEVEL',
  FLOW_RATE = 'FLOW_RATE'
}

// 指标单位映射
export const MetricUnits: Record<string, string> = {
  TEMP: '°C',
  HUMIDITY: '%',
  PH: '',
  EC: 'mS/cm',
  CO2: 'ppm',
  LIGHT: 'lux',
  WATER_LEVEL: 'cm',
  FLOW_RATE: 'L/min'
}

// 指标名称映射
export const MetricNames: Record<string, string> = {
  TEMP: '温度',
  HUMIDITY: '湿度',
  PH: 'pH值',
  EC: '电导率',
  CO2: 'CO2',
  LIGHT: '光照',
  WATER_LEVEL: '水位',
  FLOW_RATE: '流量'
}