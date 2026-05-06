import type { PaginatedResponse } from './api'

export interface NutrientTank {
  id: number
  growing_zone_id: number
  code: string
  total_volume_liter: number
  current_volume_liter?: number
  status: string
  created_at: string
  updated_at: string
}

export interface SolutionChangeEvent {
  id: number
  tank_id: number
  change_type: 'FULL_REPLACE' | 'PARTIAL_REFRESH' | 'TOP_UP'
  volume_replaced_liter: number
  source_water_ec?: number
  source_water_ph?: number
  before_ec?: number
  before_ph?: number
  after_ec?: number
  after_ph?: number
  nutrient_a_added_ml?: number
  nutrient_b_added_ml?: number
  acid_added_ml?: number
  alkali_added_ml?: number
  note?: string
  operated_by?: number
  operated_at: string
  created_at: string
}

export interface IonTestRecord {
  id: number
  tank_id: number
  batch_id?: number
  sample_code: string
  sampled_at: string
  tested_at?: string
  test_method: 'LAB' | 'STRIP' | 'METER'
  no3_n?: number
  nh4_n?: number
  p?: number
  k?: number
  ca?: number
  mg?: number
  s?: number
  fe?: number
  mn?: number
  zn?: number
  b?: number
  cu?: number
  mo?: number
  ec_at_sample?: number
  ph_at_sample?: number
  lab_name?: string
  report_url?: string
  note?: string
  created_by?: number
  created_at: string
}

export interface NutrientConcentrateInventory {
  id: number
  greenhouse_id: number
  concentrate_type: 'A' | 'B' | 'ACID' | 'ALKALI'
  brand?: string
  product_name?: string
  total_volume_ml: number
  remaining_volume_ml: number
  unit_price?: number
  batch_no?: string
  expired_at?: string
  status: 'IN_USE' | 'EMPTY' | 'EXPIRED'
  created_at: string
  updated_at: string
}

export interface ConcentrateUsageLog {
  id: number
  inventory_id: number
  solution_change_id?: number
  tank_id?: number
  volume_used_ml: number
  used_by?: number
  used_at: string
  created_at: string
}

// Request types
export interface CreateNutrientTankRequest {
  growing_zone_id: number
  code: string
  total_volume_liter: number
}

export interface CreateSolutionChangeRequest {
  tank_id: number
  change_type: string
  volume_replaced_liter: number
  source_water_ec?: number
  source_water_ph?: number
  before_ec?: number
  before_ph?: number
  after_ec?: number
  after_ph?: number
  nutrient_a_added_ml?: number
  nutrient_b_added_ml?: number
  acid_added_ml?: number
  alkali_added_ml?: number
  note?: string
  operated_at: string
}

export interface CreateIonTestRequest {
  tank_id: number
  batch_id?: number
  sample_code: string
  sampled_at: string
  test_method?: string
  no3_n?: number
  nh4_n?: number
  p?: number
  k?: number
  ca?: number
  mg?: number
  s?: number
  fe?: number
  mn?: number
  zn?: number
  b?: number
  cu?: number
  mo?: number
  ec_at_sample?: number
  ph_at_sample?: number
  lab_name?: string
  report_url?: string
  note?: string
}

export interface CreateConcentrateInventoryRequest {
  greenhouse_id: number
  concentrate_type: string
  brand?: string
  product_name?: string
  total_volume_ml: number
  unit_price?: number
  batch_no?: string
  expired_at?: string
}

// List responses
export interface NutrientTankListResponse extends PaginatedResponse<NutrientTank> {}
export interface SolutionChangeListResponse extends PaginatedResponse<SolutionChangeEvent> {}
export interface IonTestListResponse extends PaginatedResponse<IonTestRecord> {}
export interface ConcentrateInventoryListResponse extends PaginatedResponse<NutrientConcentrateInventory> {}
export interface ConcentrateUsageLogListResponse extends PaginatedResponse<ConcentrateUsageLog> {}
