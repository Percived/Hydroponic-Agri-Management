import type { PaginatedResponse } from './api'

export interface Greenhouse {
  id: number
  code: string
  name: string
  location?: string
  area_sqm?: number
  description?: string
  status: 'ENABLED' | 'DISABLED'
  created_at: string
  updated_at: string
  zone_count?: number
}

export interface GrowingZone {
  id: number
  greenhouse_id: number
  code: string
  name: string
  system_type: 'DWC' | 'NFT' | 'EBB_FLOW' | 'DRIP'
  tank_volume_liter?: number
  planting_density_per_sqm?: number
  status: 'ENABLED' | 'DISABLED'
  created_at: string
  updated_at: string
}

export interface CreateGreenhouseRequest {
  code: string
  name: string
  location?: string
  area_sqm?: number
  description?: string
}

export interface UpdateGreenhouseRequest {
  name?: string
  location?: string
  area_sqm?: number
  description?: string
  status?: string
}

export interface CreateGrowingZoneRequest {
  greenhouse_id: number
  code: string
  name: string
  system_type?: string
  tank_volume_liter?: number
  planting_density_per_sqm?: number
}

export interface UpdateGrowingZoneRequest {
  name?: string
  system_type?: string
  tank_volume_liter?: number
  planting_density_per_sqm?: number
  status?: string
}

export interface GreenhouseListResponse extends PaginatedResponse<Greenhouse> {}
export interface GrowingZoneListResponse extends PaginatedResponse<GrowingZone> {}
