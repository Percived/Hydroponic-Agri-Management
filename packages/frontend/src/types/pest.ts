import type { PaginatedResponse } from './api'

export interface PestDiseaseObservation {
  id: number
  greenhouse_id: number
  growing_zone_id?: number
  batch_id?: number
  observed_at: string
  pest_or_disease: string
  severity: 'LIGHT' | 'MODERATE' | 'SEVERE'
  affected_area_pct?: number
  affected_plant_count?: number
  symptoms?: string
  photo_urls?: string[]
  observed_by?: number
  created_at: string
}

export interface TreatmentRecord {
  id: number
  observation_id?: number
  greenhouse_id: number
  growing_zone_id?: number
  batch_id?: number
  treatment_type: 'CHEMICAL' | 'BIOLOGICAL' | 'PHYSICAL'
  product_name: string
  active_ingredient?: string
  dosage: string
  application_method: 'SPRAY' | 'DRENCH' | 'FOG' | 'RELEASE'
  safety_interval_days?: number
  reentry_interval_hours?: number
  treated_at: string
  treated_by?: number
  note?: string
  created_at: string
}

// Request types
export interface CreatePestObservationRequest {
  greenhouse_id: number
  growing_zone_id?: number
  batch_id?: number
  observed_at: string
  pest_or_disease: string
  severity: string
  affected_area_pct?: number
  affected_plant_count?: number
  symptoms?: string
  photo_urls?: string[]
}

export interface CreateTreatmentRecordRequest {
  observation_id?: number
  greenhouse_id: number
  growing_zone_id?: number
  batch_id?: number
  treatment_type: string
  product_name: string
  active_ingredient?: string
  dosage: string
  application_method: string
  safety_interval_days?: number
  reentry_interval_hours?: number
  treated_at: string
  note?: string
}

export interface PestObservationListResponse extends PaginatedResponse<PestDiseaseObservation> {}
export interface TreatmentRecordListResponse extends PaginatedResponse<TreatmentRecord> {}
