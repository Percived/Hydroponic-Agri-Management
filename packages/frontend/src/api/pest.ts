import { get, post, put, del } from './request'
import type {
  PestDiseaseObservation,
  PestObservationListResponse,
  CreatePestObservationRequest,
  TreatmentRecord,
  TreatmentRecordListResponse,
  CreateTreatmentRecordRequest
} from '@/types'

// ===== Pest Observations =====

export const getPestObservations = (params?: Record<string, unknown>) =>
  get<PestObservationListResponse>('/pest-observations', params)

export const getPestObservation = (id: number) =>
  get<PestDiseaseObservation>(`/pest-observations/${id}`)

export const createPestObservation = (data: CreatePestObservationRequest) =>
  post<{ id: number }>('/pest-observations', data)

export const updatePestObservation = (id: number, data: Partial<CreatePestObservationRequest>) =>
  put<PestDiseaseObservation>(`/pest-observations/${id}`, data)

export const deletePestObservation = (id: number) =>
  del<void>(`/pest-observations/${id}`)

// 获取观测关联的治疗记录
export const getObservationTreatments = (observationId: number) =>
  get<{ items: TreatmentRecord[] }>(`/pest-observations/${observationId}/treatments`)

// ===== Treatment Records =====

export const getTreatmentRecords = (params?: Record<string, unknown>) =>
  get<TreatmentRecordListResponse>('/treatment-records', params)

export const getTreatmentRecord = (id: number) =>
  get<TreatmentRecord>(`/treatment-records/${id}`)

export const createTreatmentRecord = (data: CreateTreatmentRecordRequest) =>
  post<{ id: number }>('/treatment-records', data)

export const updateTreatmentRecord = (id: number, data: Partial<CreateTreatmentRecordRequest>) =>
  put<TreatmentRecord>(`/treatment-records/${id}`, data)

export const deleteTreatmentRecord = (id: number) =>
  del<void>(`/treatment-records/${id}`)
