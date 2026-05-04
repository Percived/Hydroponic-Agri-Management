// 仪表盘概览数据（后端返回扁平结构）
export interface DashboardOverview {
  devices_online: number
  devices_offline: number
  alerts_open: number
  today_data_points?: number
}

// 告警摘要
export interface AlertSummary {
  id: number
  type: string
  level: string
  message: string
  device_name: string
  created_at: string
}

// 设备类型分布
export interface DeviceTypeDistribution {
  type: string
  count: number
}

// 设备分组分布
export interface DeviceGroupDistribution {
  group_name: string
  count: number
}

// 仪表盘完整数据（后端当前只返回概览数据）
export type DashboardData = DashboardOverview
