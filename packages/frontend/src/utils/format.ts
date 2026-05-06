// 日期格式化
export function formatDate(date: string | Date | null | undefined, format: string = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!date) return '-'
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return '-'

  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

// 数值格式化
export function formatNumber(value: number | null | undefined, decimals: number = 2): string {
  if (value === null || value === undefined) return '-'
  return value.toFixed(decimals)
}

// 相对时间
export function formatRelativeTime(date: string | Date | null | undefined): string {
  if (!date) return '-'
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return '-'

  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}天前`
  if (hours > 0) return `${hours}小时前`
  if (minutes > 0) return `${minutes}分钟前`
  if (seconds > 0) return `${seconds}秒前`
  return '刚刚'
}

// 设备类型名称
export function getDeviceTypeName(type: string): string {
  const typeMap: Record<string, string> = {
    SENSOR: '传感器',
    ACTUATOR: '执行器'
  }
  return typeMap[type] || type
}

// 设备分类名称
export function getCategoryName(category: string): string {
  const categoryMap: Record<string, string> = {
    TEMP: '温度',
    HUMIDITY: '湿度',
    PH: 'pH值',
    EC: '电导率',
    CO2: 'CO2',
    LIGHT: '光照',
    FAN: '风机',
    PUMP: '水泵',
    VALVE: '阀门'
  }
  return categoryMap[category] || category
}

// 日期时间格式化（中文本地化）
export function formatDateTime(dateStr: string | undefined | null): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 告警类型名称
export function getAlertTypeName(type: string): string {
  const map: Record<string, string> = {
    THRESHOLD: '阈值',
    DEVICE_OFFLINE: '设备离线',
    SYSTEM: '系统'
  }
  return map[type] || type
}

// 告警级别样式类型
export function getAlertLevelType(level: string): string {
  const map: Record<string, string> = {
    CRITICAL: 'danger',
    WARN: 'warning',
    INFO: 'info'
  }
  return map[level] || 'info'
}

// 告警级别名称
export function getAlertLevelName(level: string): string {
  const map: Record<string, string> = {
    CRITICAL: '严重',
    WARN: '警告',
    INFO: '信息'
  }
  return map[level] || level
}

// 告警状态样式类型
export function getAlertStatusType(status: string): string {
  const map: Record<string, string> = {
    OPEN: 'danger',
    ACKNOWLEDGED: 'warning',
    RESOLVED: 'success',
    IGNORED: 'info'
  }
  return map[status] || 'info'
}

// 告警状态名称
export function getAlertStatusName(status: string): string {
  const map: Record<string, string> = {
    OPEN: '开放',
    ACKNOWLEDGED: '已确认',
    RESOLVED: '已解决',
    IGNORED: '已忽略'
  }
  return map[status] || status
}

// 命令类型名称
export function getCommandTypeName(type: string): string {
  const map: Record<string, string> = {
    SWITCH: '开关',
    SET_VALUE: '设置值',
    CALIBRATE: '校准'
  }
  return map[type] || type
}

// 命令状态样式类型
export function getCommandStatusType(status: string): string {
  const normalized = status.toUpperCase()
  const map: Record<string, string> = {
    PENDING: 'info',
    SENT: 'warning',
    EXECUTED: 'success',
    FAILED: 'danger',
    QUEUED: 'info',
    ACKED: 'success',
    TIMEOUT: 'danger',
    CANCELLED: 'info'
  }
  return map[normalized] || 'info'
}

// 命令状态名称
export function getCommandStatusName(status: string): string {
  const normalized = status.toUpperCase()
  const map: Record<string, string> = {
    PENDING: '待发送',
    SENT: '已发送',
    EXECUTED: '已执行',
    FAILED: '失败',
    QUEUED: '排队中',
    ACKED: '已确认',
    TIMEOUT: '超时',
    CANCELLED: '已取消'
  }
  return map[normalized] || status
}

// 角色名称
export function getRoleName(role: string): string {
  const map: Record<string, string> = {
    ADMIN: '管理员',
    OPERATOR: '操作员',
    VIEWER: '只读'
  }
  return map[role] || role
}

// 审计操作类型名称
export function getAuditActionName(action: string): string {
  const map: Record<string, string> = {
    LOGIN: '登录',
    LOGOUT: '登出',
    CREATE_DEVICE: '创建设备',
    UPDATE_DEVICE: '更新设备',
    DELETE_DEVICE: '删除设备',
    CONTROL_CMD: '控制命令',
    CREATE_RULE: '创建规则',
    UPDATE_RULE: '更新规则',
    DELETE_RULE: '删除规则',
    CREATE_USER: '创建用户',
    UPDATE_USER: '更新用户',
    UPDATE_ALERT: '处理告警'
  }
  return map[action] || action
}

// 审计目标类型名称
export function getTargetTypeName(type: string): string {
  const map: Record<string, string> = {
    USER: '用户',
    DEVICE: '设备',
    RULE: '规则',
    ALERT: '告警',
    COMMAND: '命令'
  }
  return map[type] || type
}

// 指标名称
export function getMetricName(metric: string): string {
  const map: Record<string, string> = {
    TEMP: '温度',
    HUMIDITY: '湿度',
    PH: 'pH值',
    EC: '电导率',
    CO2: 'CO2',
    LIGHT: '光照'
  }
  return map[metric] || metric
}
