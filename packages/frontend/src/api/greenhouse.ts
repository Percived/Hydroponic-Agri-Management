import { get, post, put, del } from './request'
import type {
  Greenhouse,
  GreenhouseListResponse,
  CreateGreenhouseRequest,
  UpdateGreenhouseRequest,
  GrowingZone,
  GrowingZoneListResponse,
  CreateGrowingZoneRequest,
  UpdateGrowingZoneRequest
} from '@/types'

// ===== Greenhouses =====

export const getGreenhouses = (params?: Record<string, unknown>) =>
  get<GreenhouseListResponse>('/greenhouses', params)

export const getGreenhouse = (id: number) =>
  get<Greenhouse>(`/greenhouses/${id}`)

export const createGreenhouse = (data: CreateGreenhouseRequest) =>
  post<{ id: number }>('/greenhouses', data)

export const updateGreenhouse = (id: number, data: UpdateGreenhouseRequest) =>
  put<Greenhouse>(`/greenhouses/${id}`, data)

export const deleteGreenhouse = (id: number) =>
  del<void>(`/greenhouses/${id}`)

// 获取温室下的种植区
export const getGreenhouseZones = (id: number) =>
  get<{ items: GrowingZone[] }>(`/greenhouses/${id}/zones`)

// ===== Growing Zones =====

export const getGrowingZones = (params?: Record<string, unknown>) =>
  get<GrowingZoneListResponse>('/growing-zones', params)

export const getGrowingZone = (id: number) =>
  get<GrowingZone>(`/growing-zones/${id}`)

export const createGrowingZone = (data: CreateGrowingZoneRequest) =>
  post<{ id: number }>('/growing-zones', data)

export const updateGrowingZone = (id: number, data: UpdateGrowingZoneRequest) =>
  put<GrowingZone>(`/growing-zones/${id}`, data)

export const deleteGrowingZone = (id: number) =>
  del<void>(`/growing-zones/${id}`)
