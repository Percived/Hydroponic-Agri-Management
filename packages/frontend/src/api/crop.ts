import { get, post, put, del } from './request'
import type {
  CropVariety,
  CropVarietyListResponse,
  GrowthStage,
  GrowthStageListResponse,
  CropBatch,
  CropBatchListResponse,
  CreateCropBatchRequest,
  BatchStagePlan,
  BatchStagePlanListResponse,
  CreateBatchStagePlanRequest,
  HarvestRecordListResponse,
  CreateHarvestRecordRequest,
  BatchDevice,
  BindDeviceRequest,
  PlantingRecord,
  CreatePlantingRecordRequest,
  StageProgress,
  BatchDashboard
} from '@/types'

// ===== Crop Varieties =====

export const getCropVarieties = (params?: Record<string, unknown>) =>
  get<CropVarietyListResponse>('/crop-varieties', params)

export const getCropVariety = (id: number) =>
  get<CropVariety>(`/crop-varieties/${id}`)

export const createCropVariety = (data: { code: string; name: string; description?: string; default_cycle_days?: number }) =>
  post<{ id: number }>('/crop-varieties', data)

export const updateCropVariety = (id: number, data: { name?: string; description?: string; default_cycle_days?: number }) =>
  put<CropVariety>(`/crop-varieties/${id}`, data)

export const deleteCropVariety = (id: number) =>
  del<void>(`/crop-varieties/${id}`)

// ===== Growth Stages =====

export const getGrowthStages = (params?: Record<string, unknown>) =>
  get<GrowthStageListResponse>('/growth-stages', params)

export const getGrowthStage = (id: number) =>
  get<GrowthStage>(`/growth-stages/${id}`)

export const createGrowthStage = (data: { code: string; name: string; sort_order?: number; default_duration_days?: number }) =>
  post<{ id: number }>('/growth-stages', data)

export const updateGrowthStage = (id: number, data: { name?: string; sort_order?: number; default_duration_days?: number }) =>
  put<GrowthStage>(`/growth-stages/${id}`, data)

export const deleteGrowthStage = (id: number) =>
  del<void>(`/growth-stages/${id}`)

// ===== Crop Batches =====

export const getBatches = (params?: Record<string, unknown>) =>
  get<CropBatchListResponse>('/batches', params)

export const getBatch = (id: number) =>
  get<CropBatch>(`/batches/${id}`)

export const createBatch = (data: CreateCropBatchRequest) =>
  post<{ id: number }>('/batches', data)

export const updateBatch = (id: number, data: Partial<CreateCropBatchRequest>) =>
  put<CropBatch>(`/batches/${id}`, data)

export const deleteBatch = (id: number) =>
  del<void>(`/batches/${id}`)

// 批次状态过渡 (RUNNING / HARVESTING / COMPLETED / ABORTED)
export const transitionBatch = (id: number, data: { status: string; note?: string }) =>
  post<CropBatch>(`/batches/${id}/transition`, data)

// ===== Batch Stage Plans =====

export const getBatchStagePlans = (params?: Record<string, unknown>) =>
  get<BatchStagePlanListResponse>('/batch-stage-plans', params)

export const getBatchStagePlan = (id: number) =>
  get<BatchStagePlan>(`/batch-stage-plans/${id}`)

export const createBatchStagePlan = (data: CreateBatchStagePlanRequest) =>
  post<{ id: number }>('/batch-stage-plans', data)

export const updateBatchStagePlan = (id: number, data: Partial<CreateBatchStagePlanRequest>) =>
  put<BatchStagePlan>(`/batch-stage-plans/${id}`, data)

export const deleteBatchStagePlan = (id: number) =>
  del<void>(`/batch-stage-plans/${id}`)

// ===== Harvest Records =====

export const getHarvests = (params?: Record<string, unknown>) =>
  get<HarvestRecordListResponse>('/harvests', params)

export const createHarvest = (data: CreateHarvestRecordRequest) =>
  post<{ id: number }>('/harvests', data)

export const getHarvestSummary = (batchId: number) =>
  get<{
    batch_id: number
    total_weight_kg: number
    grades: { grade: string; weight_kg: number; count: number }[]
  }>(`/harvests/summary/${batchId}`)

// ===== Batch Devices =====

export const bindDevice = (batchId: number, data: BindDeviceRequest) =>
  post<{ id: number }>(`/batches/${batchId}/devices`, data)

export const unbindDevice = (batchId: number, deviceId: number, deviceType: string) =>
  del<void>(`/batches/${batchId}/devices/${deviceId}?device_type=${deviceType}`)

export const getBatchDevices = (batchId: number, deviceType?: string) =>
  get<{ items: BatchDevice[] }>(`/batches/${batchId}/devices${deviceType ? `?device_type=${deviceType}` : ''}`)

// ===== Planting Records =====

export const createPlantingRecord = (data: CreatePlantingRecordRequest) =>
  post<{ id: number }>('/planting-records', data)

export const getPlantingRecord = (batchId: number) =>
  get<PlantingRecord>(`/planting-records/${batchId}`)

export const updatePlantingRecord = (batchId: number, data: Partial<CreatePlantingRecordRequest>) =>
  put<void>(`/planting-records/${batchId}`, data)

// ===== Stage Progress =====

export const getStageProgress = (batchId: number) =>
  get<StageProgress>(`/batches/${batchId}/stage-progress`)

// ===== Batch Dashboard =====

export const getBatchDashboard = (batchId: number) =>
  get<BatchDashboard>(`/batches/${batchId}/dashboard`)
