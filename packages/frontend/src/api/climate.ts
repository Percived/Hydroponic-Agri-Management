import { get, post, put, del } from './request'
import type {
  ClimateProfile,
  ClimateProfileListResponse,
  CreateClimateProfileRequest,
  ClimateStage,
  CreateClimateStageRequest,
  ClimateStageAction,
  CreateClimateStageActionRequest,
  ClimateExecutionLog,
  ClimateExecutionLogListResponse
} from '@/types'

// ===== Climate Profiles =====

export const getClimateProfiles = (params?: Record<string, unknown>) =>
  get<ClimateProfileListResponse>('/climate-profiles', params)

export const getClimateProfile = (id: number) =>
  get<ClimateProfile>(`/climate-profiles/${id}`)

export const createClimateProfile = (data: CreateClimateProfileRequest) =>
  post<{ id: number }>('/climate-profiles', data)

export const updateClimateProfile = (id: number, data: Partial<CreateClimateProfileRequest>) =>
  put<ClimateProfile>(`/climate-profiles/${id}`, data)

export const deleteClimateProfile = (id: number) =>
  del<void>(`/climate-profiles/${id}`)

// ===== Climate Profile Stages =====

export const getClimateProfileStages = (profileId: number) =>
  get<{ items: ClimateStage[] }>(`/climate-profiles/${profileId}/stages`)

export const createClimateProfileStage = (profileId: number, data: CreateClimateStageRequest) =>
  post<{ id: number }>(`/climate-profiles/${profileId}/stages`, data)

export const updateClimateProfileStage = (profileId: number, stageId: number, data: Partial<CreateClimateStageRequest>) =>
  put<ClimateStage>(`/climate-profiles/${profileId}/stages/${stageId}`, data)

export const deleteClimateProfileStage = (profileId: number, stageId: number) =>
  del<void>(`/climate-profiles/${profileId}/stages/${stageId}`)

// ===== Climate Stage Actions =====

export const getClimateStageActions = (profileId: number, stageId: number) =>
  get<{ items: ClimateStageAction[] }>(`/climate-profiles/${profileId}/stages/${stageId}/actions`)

export const createClimateStageAction = (profileId: number, stageId: number, data: CreateClimateStageActionRequest) =>
  post<{ id: number }>(`/climate-profiles/${profileId}/stages/${stageId}/actions`, data)

export const updateClimateStageAction = (profileId: number, stageId: number, actionId: number, data: Partial<CreateClimateStageActionRequest>) =>
  put<ClimateStageAction>(`/climate-profiles/${profileId}/stages/${stageId}/actions/${actionId}`, data)

export const deleteClimateStageAction = (profileId: number, stageId: number, actionId: number) =>
  del<void>(`/climate-profiles/${profileId}/stages/${stageId}/actions/${actionId}`)

// ===== Climate Executions =====

export const executeClimateProfile = (profileId: number) =>
  post<ClimateExecutionLog>(`/climate-profiles/${profileId}/execute`)

export const getClimateExecutionLogs = (params?: Record<string, unknown>) =>
  get<ClimateExecutionLogListResponse>('/climate-execution-logs', params)
