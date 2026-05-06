// 设备在线状态（资产拓扑）
export enum AssetOnlineStatus {
  ONLINE = 'ONLINE',
  OFFLINE = 'OFFLINE',
  ABNORMAL = 'ABNORMAL'
}

// 通道类型
export enum AssetChannelType {
  SENSOR = 'SENSOR',
  ACTUATOR = 'ACTUATOR'
}

// 命令状态机
export enum AssetCommandStatus {
  QUEUED = 'queued',
  SENT = 'sent',
  ACKED = 'acked',
  FAILED = 'failed',
  TIMEOUT = 'timeout',
  CANCELLED = 'cancelled'
}

// 告警等级
export enum AssetAlertLevel {
  CRITICAL = 'CRITICAL',
  WARN = 'WARN',
  INFO = 'INFO'
}

// 告警状态
export enum AssetAlertStatus {
  OPEN = 'OPEN',
  ACKED = 'ACKED',
  CLOSED = 'CLOSED'
}
