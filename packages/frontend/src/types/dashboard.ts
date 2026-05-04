// 设备类型分布
export interface DeviceTypeDistribution {
  type: string
  count: number
}

// 温室概览
export interface GreenhouseSummary {
  greenhouse_id: number
  name: string
  device_count: number
  avg_temp: number | null
  avg_humidity: number | null
}

// 最近控制命令
export interface RecentCommand {
  id: number
  command_type: string
  device_name: string
  status: string
  created_at: string
}

// 仪表盘概览数据（后端返回扁平结构）
export interface DashboardOverview {
  devices_online: number
  devices_offline: number
  devices_total: number
  alerts_open: number
  alerts_critical: number
  alerts_today: number
  device_type_distribution: DeviceTypeDistribution[]
  greenhouse_summary: GreenhouseSummary[]
  recent_commands: RecentCommand[]
}

// 仪表盘完整数据（同 DashboardOverview）
export type DashboardData = DashboardOverview
