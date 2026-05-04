// 告警类型
export enum AlertType {
  THRESHOLD = 'THRESHOLD',
  OFFLINE = 'OFFLINE',
  DEVICE_ERROR = 'DEVICE_ERROR'
}

// 告警级别
export enum AlertLevel {
  INFO = 'INFO',
  WARN = 'WARN',
  CRITICAL = 'CRITICAL'
}

// 告警状态
export enum AlertStatus {
  OPEN = 'OPEN',
  ACK = 'ACK',
  CLOSED = 'CLOSED'
}

// 告警
export interface Alert {
  id: number
  type: AlertType
  level: AlertLevel
  message: string
  device_id: number
  device_name: string
  status: AlertStatus
  remark: string | null
  created_at: string
  updated_at: string
}

// 告警统计
export interface AlertStats {
  open_count: number
  ack_count: number
  closed_count: number
}

// 告警查询参数
export interface AlertQueryParams {
  page?: number
  page_size?: number
  type?: AlertType
  level?: AlertLevel
  status?: AlertStatus
}

// 更新告警状态参数
export interface UpdateAlertStatusParams {
  status: AlertStatus
  comment?: string
}
