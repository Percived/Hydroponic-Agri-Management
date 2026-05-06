import type { PaginatedResponse } from './api'

export interface ControlPolicy {
  id: number
  policy_code: string
  name: string
  policy_type: 'THRESHOLD' | 'SCHEDULE' | 'DURATION'
  greenhouse_id: number
  growing_zone_id?: number
  priority: number
  retry_limit: number
  timeout_sec: number
  enabled: number
  version: string
  effective_from?: string
  effective_to?: string
  created_by?: number
  published_by?: number
  published_at?: string
  created_at: string
  updated_at: string
  conditions?: PolicyCondition[]
  targets?: PolicyTarget[]
}

export interface PolicyCondition {
  id: number
  policy_id: number
  metric_code: string
  operator: string
  threshold_value: number
  hysteresis?: number
  window_sec?: number
  required_duration_sec?: number
  aggregation?: string
  enabled: number
  created_at: string
  updated_at: string
}

export interface PolicyTarget {
  id: number
  policy_id: number
  actuator_channel_id: number
  command_type: string
  command_payload: Record<string, unknown>
  execution_order: number
  enabled: number
  created_at: string
  updated_at: string
}

export interface PolicyExecution {
  id: number
  policy_id: number
  trigger_source: 'TELEMETRY' | 'SCHEDULE' | 'MANUAL'
  trigger_metric_code?: string
  trigger_value?: number
  decision: 'EXECUTED' | 'SKIPPED' | 'FAILED' | 'CONFLICT'
  decision_reason?: string
  command_id?: number
  batch_id?: number
  executed_at?: string
  created_at: string
}

// Request types
export interface CreatePolicyRequest {
  policy_code: string
  name: string
  policy_type: string
  greenhouse_id: number
  growing_zone_id?: number
  priority?: number
  retry_limit?: number
  timeout_sec?: number
  conditions?: CreateConditionRequest[]
  targets?: CreateTargetRequest[]
}

export interface CreateConditionRequest {
  metric_code: string
  operator: string
  threshold_value: number
  hysteresis?: number
  window_sec?: number
  required_duration_sec?: number
  aggregation?: string
}

export interface CreateTargetRequest {
  actuator_channel_id: number
  command_type: string
  command_payload: Record<string, unknown>
  execution_order?: number
}

// List responses
export interface ControlPolicyListResponse extends PaginatedResponse<ControlPolicy> {}
export interface PolicyExecutionListResponse extends PaginatedResponse<PolicyExecution> {}
