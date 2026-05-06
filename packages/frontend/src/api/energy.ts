import { get, post, put, del } from './request'
import type {
  EnergyConsumptionRecord,
  EnergyRecordListResponse,
  CreateEnergyRecordRequest,
  EnergySummary
} from '@/types'

// ===== Energy Records =====

export const getEnergyRecords = (params?: Record<string, unknown>) =>
  get<EnergyRecordListResponse>('/energy-records', params)

export const getEnergyRecord = (id: number) =>
  get<EnergyConsumptionRecord>(`/energy-records/${id}`)

export const createEnergyRecord = (data: CreateEnergyRecordRequest) =>
  post<{ id: number }>('/energy-records', data)

export const updateEnergyRecord = (id: number, data: Partial<CreateEnergyRecordRequest>) =>
  put<EnergyConsumptionRecord>(`/energy-records/${id}`, data)

export const deleteEnergyRecord = (id: number) =>
  del<void>(`/energy-records/${id}`)

// 按温室查询能耗记录
export const getEnergyRecordsByGreenhouse = (greenhouseId: number, params?: Record<string, unknown>) =>
  get<EnergyRecordListResponse>(`/energy-records/greenhouse/${greenhouseId}`, params)

// 按批次查询能耗记录
export const getEnergyRecordsByBatch = (batchId: number, params?: Record<string, unknown>) =>
  get<EnergyRecordListResponse>(`/energy-records/batch/${batchId}`, params)

// 能耗汇总
export const getEnergySummary = (params?: Record<string, unknown>) =>
  get<{ items: EnergySummary[] }>('/energy-records/summary', params)
