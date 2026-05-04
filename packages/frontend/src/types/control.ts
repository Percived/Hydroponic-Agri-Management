// 控制命令类型
export enum CommandType {
  SWITCH = 'SWITCH',
  SET_VALUE = 'SET_VALUE',
  CALIBRATE = 'CALIBRATE'
}

// 控制命令状态
export enum CommandStatus {
  PENDING = 'PENDING',
  SENT = 'SENT',
  EXECUTED = 'EXECUTED',
  FAILED = 'FAILED'
}

// 控制命令
export interface ControlCommand {
  id: number
  device_id: number
  device_name: string
  command_type: CommandType
  payload: Record<string, unknown>
  status: CommandStatus
  error_message: string | null
  created_at: string
  executed_at: string | null
}

// 创建控制命令参数
export interface CreateCommandParams {
  device_id: number
  command_type: CommandType
  payload: Record<string, unknown>
}

// 控制命令查询参数
export interface CommandQueryParams {
  page?: number
  page_size?: number
  device_id?: number
  status?: CommandStatus
}

// 控制规则
export interface ControlRule {
  id: number
  name: string
  metric_code: string
  operator: string
  threshold: number
  target_device_id: number
  target_device_name?: string
  command_type: CommandType
  command_payload: Record<string, unknown>
  enabled: boolean
  created_at: string
  updated_at: string
}

// 创建/更新控制规则参数
export interface ControlRuleFormData {
  name: string
  metric_code: string
  operator: string
  threshold: number
  target_device_id: number
  command_type: CommandType
  command_payload: Record<string, unknown>
  enabled: boolean
}

// 控制规则查询参数
export interface RuleQueryParams {
  page?: number
  page_size?: number
  metric_code?: string
  enabled?: boolean
}
