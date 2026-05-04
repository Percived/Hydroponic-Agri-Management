import { get, post, put } from './request'
import type { Greenhouse, GreenhouseQueryParams, GreenhouseFormData, PaginatedData } from '@/types'

// 获取温室列表
export const getGreenhouses = (params?: GreenhouseQueryParams) =>
  get<PaginatedData<Greenhouse>>('/devices/greenhouses', params)

// 获取温室详情
export const getGreenhouseDetail = (id: number) =>
  get<Greenhouse>(`/devices/greenhouses/${id}`)

// 创建温室
export const createGreenhouse = (data: GreenhouseFormData) =>
  post<{ id: number }>('/devices/greenhouses', data)

// 更新温室
export const updateGreenhouse = (id: number, data: Partial<GreenhouseFormData>) =>
  put<void>(`/devices/greenhouses/${id}`, data)
