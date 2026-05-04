import { get, post, put, del } from './request'
import type {
  ControlCommand,
  CreateCommandParams,
  CommandQueryParams,
  ControlRule,
  ControlRuleFormData,
  RuleQueryParams,
  PaginatedData
} from '@/types'
import { CommandType } from '@/types'

// ===== 控制命令 =====

// 获取命令列表
export const getCommands = (params?: CommandQueryParams) =>
  get<PaginatedData<ControlCommand>>('/controls/commands', params)

// 获取命令详情
export const getCommandDetail = (id: number) =>
  get<ControlCommand>(`/controls/commands/${id}`)

// 下发命令
export const createCommand = (data: CreateCommandParams) =>
  post<ControlCommand>('/controls/commands', data)

// ===== 控制规则 =====

// 获取规则列表
export const getRules = async (params?: RuleQueryParams): Promise<PaginatedData<ControlRule>> => {
  const result = await get<PaginatedData<Record<string, unknown>>>('/controls/rules', params)
  return {
    ...result,
    items: (result.items as Array<Record<string, unknown>>).map((it) => {
      const action = (it.action ?? it.Action) as Record<string, unknown> | undefined
      return {
        id: it.id as number,
        name: it.name as string,
        metric_code: (it.metric_code ?? it.metric) as string,
        operator: it.operator as string,
        threshold: it.threshold as number,
        action: action,
        target_device_id: it.target_device_id as number,
        target_device_name: (it.target_device_name ?? it.TargetDeviceName ?? '') as string,
        command_type: ((action?.command_type as string) ?? (it.command_type as string) ?? 'SWITCH') as CommandType,
        command_payload: (action?.payload as Record<string, unknown>) ?? (it.command_payload as Record<string, unknown>) ?? {},
        enabled: (it.enabled as boolean) ?? true,
        created_at: it.created_at as string,
        updated_at: (it.updated_at ?? it.UpdatedAt ?? it.created_at) as string
      } satisfies ControlRule
    })
  }
}

// 获取规则详情
export const getRuleDetail = (id: number) =>
  get<ControlRule>(`/controls/rules/${id}`)

function toRuleRequest(data: ControlRuleFormData): Record<string, unknown> {
  return {
    name: data.name,
    metric_code: data.metric_code,
    operator: data.operator,
    threshold: data.threshold,
    target_device_id: data.target_device_id,
    enabled: data.enabled,
    action: {
      command_type: data.command_type,
      payload: data.command_payload
    }
  }
}

// 创建规则
export const createRule = (data: ControlRuleFormData) =>
  post<ControlRule>('/controls/rules', toRuleRequest(data))

// 更新规则
export const updateRule = (id: number, data: ControlRuleFormData) =>
  put<ControlRule>(`/controls/rules/${id}`, {
    name: data.name,
    operator: data.operator,
    threshold: data.threshold,
    target_device_id: data.target_device_id,
    enabled: data.enabled,
    action: {
      command_type: data.command_type,
      payload: data.command_payload
    }
  })

// 删除规则
export const deleteRule = (id: number) =>
  del<void>(`/controls/rules/${id}`)

// 启用/禁用规则
export const toggleRule = (id: number, enabled: boolean) =>
  put<ControlRule>(`/controls/rules/${id}`, { enabled })
