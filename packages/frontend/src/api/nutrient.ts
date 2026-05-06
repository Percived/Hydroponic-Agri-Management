import { get, post, put, del } from './request'
import type {
  NutrientTank,
  NutrientTankListResponse,
  CreateNutrientTankRequest,
  SolutionChangeListResponse,
  CreateSolutionChangeRequest,
  IonTestRecord,
  IonTestListResponse,
  CreateIonTestRequest,
  NutrientConcentrateInventory,
  ConcentrateInventoryListResponse,
  CreateConcentrateInventoryRequest,
  ConcentrateUsageLogListResponse
} from '@/types'

// ===== Nutrient Tanks =====

export const getNutrientTanks = (params?: Record<string, unknown>) =>
  get<NutrientTankListResponse>('/nutrient-tanks', params)

export const getNutrientTank = (id: number) =>
  get<NutrientTank>(`/nutrient-tanks/${id}`)

export const createNutrientTank = (data: CreateNutrientTankRequest) =>
  post<{ id: number }>('/nutrient-tanks', data)

export const updateNutrientTank = (id: number, data: Partial<CreateNutrientTankRequest>) =>
  put<NutrientTank>(`/nutrient-tanks/${id}`, data)

export const deleteNutrientTank = (id: number) =>
  del<void>(`/nutrient-tanks/${id}`)

// ===== Solution Change Events =====

export const getSolutionChanges = (params?: Record<string, unknown>) =>
  get<SolutionChangeListResponse>('/solution-changes', params)

export const createSolutionChange = (data: CreateSolutionChangeRequest) =>
  post<{ id: number }>('/solution-changes', data)

// ===== Ion Test Records =====

export const getIonTests = (params?: Record<string, unknown>) =>
  get<IonTestListResponse>('/ion-tests', params)

export const getIonTest = (id: number) =>
  get<IonTestRecord>(`/ion-tests/${id}`)

export const createIonTest = (data: CreateIonTestRequest) =>
  post<{ id: number }>('/ion-tests', data)

export const updateIonTest = (id: number, data: Partial<CreateIonTestRequest>) =>
  put<IonTestRecord>(`/ion-tests/${id}`, data)

export const deleteIonTest = (id: number) =>
  del<void>(`/ion-tests/${id}`)

// ===== Concentrate Inventory =====

export const getConcentrateInventory = (params?: Record<string, unknown>) =>
  get<ConcentrateInventoryListResponse>('/concentrate-inventory', params)

export const getConcentrateItem = (id: number) =>
  get<NutrientConcentrateInventory>(`/concentrate-inventory/${id}`)

export const createConcentrateItem = (data: CreateConcentrateInventoryRequest) =>
  post<{ id: number }>('/concentrate-inventory', data)

export const updateConcentrateItem = (id: number, data: Partial<CreateConcentrateInventoryRequest>) =>
  put<NutrientConcentrateInventory>(`/concentrate-inventory/${id}`, data)

export const deleteConcentrateItem = (id: number) =>
  del<void>(`/concentrate-inventory/${id}`)

// ===== Concentrate Usage Logs =====

export const getConcentrateUsageLogs = (params?: Record<string, unknown>) =>
  get<ConcentrateUsageLogListResponse>('/concentrate-usage-logs', params)

export const createConcentrateUsageLog = (data: {
  inventory_id: number
  solution_change_id?: number
  tank_id?: number
  volume_used_ml: number
  used_at: string
}) =>
  post<{ id: number }>('/concentrate-usage-logs', data)
