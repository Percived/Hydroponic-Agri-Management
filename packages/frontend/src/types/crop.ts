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
  active_recipe_id?: number
  active_policy_id?: number
  active_climate_profile_id?: number
  note?: string
  created_by?: number
  created_at: string
  updated_at: string
}

export interface BatchStagePlan {
  id: number
  batch_id: number
  growth_stage_id: number
  recipe_id?: number
  policy_id?: number
  climate_profile_id?: number
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
  recipe_id?: number
  policy_id?: number
  climate_profile_id?: number
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

// Batch device binding
export interface BatchDevice {
  id: number
  batch_id: number
  device_type: 'sensor' | 'actuator'
  device_id: number
  device_name?: string
  device_code?: string
  is_active: boolean
  bound_at: string
  unbound_at?: string
}

export interface BindDeviceRequest {
  device_type: 'sensor' | 'actuator'
  device_id: number
}

// Planting record
export interface PlantingRecord {
  id: number
  batch_id: number
  seed_source?: string
  seed_batch_no?: string
  seedling_age_days?: number
  seeded_at?: string
  planted_at?: string
  actual_plant_count?: number
  initial_ec?: number
  initial_ph?: number
  initial_water_temp?: number
  initial_nutrient_recipe_id?: number
  planted_by?: number
  note?: string
  created_at: string
  updated_at: string
}

export interface CreatePlantingRecordRequest {
  batch_id: number
  seed_source?: string
  seed_batch_no?: string
  seedling_age_days?: number
  seeded_at?: string
  planted_at?: string
  actual_plant_count?: number
  initial_ec?: number
  initial_ph?: number
  initial_water_temp?: number
  initial_nutrient_recipe_id?: number
  note?: string
}

// Stage progress
export interface StageProgress {
  batch_id: number
  current_stage_id?: number
  current_stage_name: string
  current_stage_code: string
  progress_percent: number
  days_elapsed: number
  days_remaining: number
  target_ec_min?: number
  target_ec_max?: number
  target_ph_min?: number
  target_ph_max?: number
}

// Batch dashboard (Phase 3)
export interface BatchDashboard {
  batch: CropBatch
  variety?: CropVariety
  greenhouse_name?: string
  zone_name?: string
  planting_record?: PlantingRecord
  stage_progress?: StageProgress
  devices: BatchDevice[]
  latest_telemetry: TelemetrySnapshot[]
  recent_alerts: AlertSummary[]
  recent_commands: CommandSummary[]
  harvest_summary?: HarvestSummary
  stage_plans: BatchStagePlan[]
}

export interface TelemetrySnapshot {
  metric_code: string
  metric_name: string
  value: number
  unit: string
  collected_at: string
}

export interface AlertSummary {
  id: number
  type: string
  level: string
  message: string
  status: string
  triggered_at: string
}

export interface CommandSummary {
  id: number
  command_type: string
  status: string
  created_at: string
}

export interface HarvestSummary {
  total_weight_kg: number
  grades: { grade: string; weight_kg: number; count: number }[]
}

// List responses
export interface CropVarietyListResponse extends PaginatedResponse<CropVariety> {}
export interface GrowthStageListResponse extends PaginatedResponse<GrowthStage> {}
export interface CropBatchListResponse extends PaginatedResponse<CropBatch> {}
export interface BatchStagePlanListResponse extends PaginatedResponse<BatchStagePlan> {}
export interface HarvestRecordListResponse extends PaginatedResponse<HarvestRecord> {}
