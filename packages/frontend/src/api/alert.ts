import { get, patch } from './request'
import type { Alert, AlertStats, AlertQueryParams, UpdateAlertStatusParams, PaginatedData } from '@/types'

// 获取告警列表
export const getAlerts = (params?: AlertQueryParams) =>
  get<PaginatedData<Alert>>('/alerts', params)

// 获取告警详情
export const getAlertDetail = (id: number) =>
  get<Alert>(`/alerts/${id}`)

// 获取告警统计
export const getAlertStats = () =>
  get<AlertStats>('/alerts/stats')

// 更新告警状态
export const updateAlertStatus = (id: number, data: UpdateAlertStatusParams) =>
  patch<Alert>(`/alerts/${id}/status`, data)
