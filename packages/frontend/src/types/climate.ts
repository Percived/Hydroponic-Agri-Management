import type { PaginatedResponse } from './api'

export interface ClimateProfile {
  id: number
  greenhouse_id: number
  code: string
  name: string
  description?: string
  trigger_metric_code: string
  enabled: boolean
  created_at: string
  updated_at: string
  stages?: ClimateStage[]
  stages_count?: number
}

export interface ClimateStage {
  id: number
  profile_id: number
  stage_level: number
  name: string
  trigger_operator: string
  trigger_threshold: number
  hysteresis: number
  created_at: string
  updated_at: string
  actions?: ClimateStageAction[]
  action_count?: number
}

export interface ClimateStageAction {
  id: number
  stage_id: number
  actuator_channel_id: number
  command_type: string
  command_payload: string
  execution_order: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface ClimateExecutionLog {
  id: number
  profile_id: number
  from_stage_level?: number
  to_stage_level: number
  trigger_value: number
  executed_actions_count: number
  executed_at: string
  created_at: string
}

// Request types
export interface CreateClimateProfileRequest {
  greenhouse_id: number
  code: string
  name: string
  description?: string
  trigger_metric_code: string
  enabled?: boolean
}

export interface CreateClimateStageRequest {
  stage_level: number
  name: string
  trigger_operator: string
  trigger_threshold: number
  hysteresis?: number
}

export interface StageWithActions {
  stage_level: number
  name: string
  trigger_operator: string
  trigger_threshold: number
  hysteresis?: number
  actions: CreateClimateStageActionRequest[]
}

export interface CreateClimateProfileWithStagesRequest {
  greenhouse_id: number
  code: string
  name: string
  description?: string
  trigger_metric_code: string
  enabled?: boolean
  stages: StageWithActions[]
}

export interface ExecuteClimateProfileRequest {
  trigger_value: number
  from_stage_level?: number
  to_stage_level: number
}

export interface CreateClimateStageActionRequest {
  actuator_channel_id: number
  command_type: string
  command_payload: Record<string, unknown>
  execution_order?: number
  enabled?: boolean
}

export interface ClimateProfileListResponse extends PaginatedResponse<ClimateProfile> {}
export interface ClimateExecutionLogListResponse extends PaginatedResponse<ClimateExecutionLog> {}
