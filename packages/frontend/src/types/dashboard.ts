// 温室概览
export interface GreenhouseSummary {
  greenhouse_id: number
  name: string
  sensor_count: number
  actuator_count: number
  zone_count: number
  avg_temp: number | null
  avg_humidity: number | null
}

// 设备类型分布
export interface DeviceTypeDistItem {
  type: string
  count: number
}

// 最近控制命令
export interface RecentCommand {
  id: number
  command_type: string
  device_name: string
  status: string
  created_at: string
}

// 仪表盘概览数据
export interface DashboardOverview {
  sensors_online: number
  sensors_offline: number
  sensors_total: number
  actuators_online: number
  actuators_offline: number
  actuators_total: number
  devices_online: number
  devices_offline: number
  devices_total: number
  alerts_open: number
  alerts_critical: number
  alerts_today: number
  device_type_distribution: DeviceTypeDistItem[]
  greenhouse_summary: GreenhouseSummary[]
  recent_commands: RecentCommand[]
}

// 仪表盘完整数据（同 DashboardOverview）
export type DashboardData = DashboardOverview
