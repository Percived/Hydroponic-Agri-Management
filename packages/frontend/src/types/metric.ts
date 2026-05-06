import type { PaginatedResponse } from './api'

export interface MetricDefinition {
  id: number
  code: string
  name: string
  unit: string
  precision_digits: number
  normal_range_min?: number
  normal_range_max?: number
  is_core: number
  status: string
  created_at: string
  updated_at: string
}

export interface MetricListResponse extends PaginatedResponse<MetricDefinition> {}
