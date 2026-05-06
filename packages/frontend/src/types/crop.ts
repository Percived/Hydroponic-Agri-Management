import type { PaginatedResponse } from './api'

export interface CropVariety {
  id: number
  code: string
  name: string
  description?: string
  default_cycle_days?: number
  created_at: string
  updated_at: string
}

export interface GrowthStage {
  id: number
  code: string
  name: string
  sort_order: number
  default_duration_days?: number
  created_at: string
  updated_at: string
}

export interface CropBatch {
  id: number
  batch_no: string
  greenhouse_id: number
  growing_zone_id?: number
  crop_variety_id: number
  variety_code?: string
  variety_name?: string
  status: 'PLANNED' | 'RUNNING' | 'HARVESTING' | 'COMPLETED' | 'ABORTED'
  planting_density?: number
  total_plants?: number
  started_at?: string
  ended_at?: string
  expected_harvest_at?: string
  recipe_version?: string
  policy_version?: string
  note?: string
  created_by?: number
  created_at: string
  updated_at: string
}

export interface BatchStagePlan {
  id: number
  batch_id: number
  growth_stage_id: number
  stage_start_at: string
  stage_end_at: string
  target_ec_min?: number
  target_ec_max?: number
  target_ph_min?: number
  target_ph_max?: number
  created_at: string
  updated_at: string
}

export interface HarvestRecord {
  id: number
  batch_id: number
  harvested_at: string
  harvest_weight_kg: number
  grade: 'A' | 'B' | 'C' | 'Waste'
  grade_weight_kg: number
  note?: string
  harvested_by?: number
  created_at: string
}

// Request types
export interface CreateCropBatchRequest {
  batch_no: string
  greenhouse_id: number
  growing_zone_id?: number
  crop_variety_id: number
  planting_density?: number
  total_plants?: number
  started_at?: string
  expected_harvest_at?: string
  recipe_version?: string
  policy_version?: string
  note?: string
}

export interface CreateBatchStagePlanRequest {
  batch_id: number
  growth_stage_id: number
  stage_start_at: string
  stage_end_at: string
  target_ec_min?: number
  target_ec_max?: number
  target_ph_min?: number
  target_ph_max?: number
}

export interface CreateHarvestRecordRequest {
  batch_id: number
  harvested_at: string
  harvest_weight_kg: number
  grade: string
  grade_weight_kg: number
  note?: string
}

// List responses
export interface CropVarietyListResponse extends PaginatedResponse<CropVariety> {}
export interface GrowthStageListResponse extends PaginatedResponse<GrowthStage> {}
export interface CropBatchListResponse extends PaginatedResponse<CropBatch> {}
export interface BatchStagePlanListResponse extends PaginatedResponse<BatchStagePlan> {}
export interface HarvestRecordListResponse extends PaginatedResponse<HarvestRecord> {}
