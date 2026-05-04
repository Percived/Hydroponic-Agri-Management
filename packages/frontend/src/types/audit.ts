// 审计日志操作类型
export enum AuditAction {
  LOGIN = 'LOGIN',
  LOGOUT = 'LOGOUT',
  CREATE_DEVICE = 'CREATE_DEVICE',
  UPDATE_DEVICE = 'UPDATE_DEVICE',
  DELETE_DEVICE = 'DELETE_DEVICE',
  CONTROL_CMD = 'CONTROL_CMD',
  CREATE_RULE = 'CREATE_RULE',
  UPDATE_RULE = 'UPDATE_RULE',
  DELETE_RULE = 'DELETE_RULE',
  CREATE_USER = 'CREATE_USER',
  UPDATE_USER = 'UPDATE_USER',
  UPDATE_ALERT = 'UPDATE_ALERT'
}

// 审计日志
export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: AuditAction
  target_type: string
  target_id: number | null
  detail: string | null
  ip_address: string | null
  created_at: string
}

// 审计日志查询参数
export interface AuditLogQueryParams {
  page?: number
  page_size?: number
  user_id?: number
  action?: AuditAction
  start_time?: string
  end_time?: string
}
