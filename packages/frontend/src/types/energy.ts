import type { PaginatedResponse } from './api'

export interface EnergyConsumptionRecord {
  id: number
  greenhouse_id: number
  record_type: 'ELECTRICITY' | 'WATER' | 'CO2_GAS'
  consumption_value: number
  unit: string
  record_period_start: string
  record_period_end: string
  meter_reading_start?: number
  meter_reading_end?: number
  batch_id?: number
  recorded_by?: number
  created_at: string
}

export interface CreateEnergyRecordRequest {
  greenhouse_id: number
  record_type: string
  consumption_value: number
  unit: string
  record_period_start: string
  record_period_end: string
  meter_reading_start?: number
  meter_reading_end?: number
  batch_id?: number
}

export interface EnergySummary {
  record_type: string
  total_consumption: number
  unit: string
}

export interface EnergyRecordListResponse extends PaginatedResponse<EnergyConsumptionRecord> {}
