export interface DashboardStats {
  active_batches_count: number
  unresolved_alerts: number
  devices_online: number
  devices_offline: number
  energy_kwh_today: number
  water_l_today: number
}

export interface GreenhouseMetrics {
  temperature: number
  humidity: number
  ec: number
  ph: number
  do: number
  co2: number
  lux: number
}

export interface DashboardGreenhouse {
  id: string
  name: string
  health_score: string
  last_collected_at: string | null
  metrics: GreenhouseMetrics
  active_strategies: string[]
}

export interface DashboardTrends {
  timestamps: string[]
  ec_avg: number[]
  ph_avg: number[]
}

export interface DashboardActiveBatch {
  batch_id: string
  crop_name: string
  stage: string
  day: number
  greenhouse_id: string
}

export interface DashboardRecentAlert {
  alert_id: string
  severity: string
  message: string
  timestamp: string
  greenhouse_name: string
}

export interface RecentCommand {
  id: number
  command_type: string
  device_name: string
  status: string
  created_at: string
}

export interface DashboardData {
  stats: DashboardStats
  greenhouses: DashboardGreenhouse[]
  trends: DashboardTrends
  active_batches: DashboardActiveBatch[]
  recent_alerts: DashboardRecentAlert[]
  recent_commands: RecentCommand[]
}
