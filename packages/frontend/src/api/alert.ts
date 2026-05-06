import { get, post, patch } from './request'
import type {
  Alert,
  AlertListResponse,
  CreateAlertRequest,
  UpdateAlertStatusRequest,
  AlertStats,
  AlertTimelineEvent,
  CreateTimelineEventRequest
} from '@/types'

// 获取告警列表
export const getAlerts = (params?: Record<string, unknown>) =>
  get<AlertListResponse>('/alerts', params)

// 创建告警
export const createAlert = (data: CreateAlertRequest) =>
  post<Alert>('/alerts', data)

// 获取告警统计
export const getAlertStats = () =>
  get<AlertStats>('/alerts/stats')

// 获取告警详情
export const getAlert = (id: number) =>
  get<Alert>(`/alerts/${id}`)

// 更新告警状态
export const updateAlertStatus = (id: number, data: UpdateAlertStatusRequest) =>
  patch<Alert>(`/alerts/${id}/status`, data)

// 获取告警时间线
export const getAlertTimeline = (id: number) =>
  get<{ items: AlertTimelineEvent[] }>(`/alerts/${id}/timeline`)

// 添加告警时间线事件
export const createAlertTimelineEvent = (id: number, data: CreateTimelineEventRequest) =>
  post<AlertTimelineEvent>(`/alerts/${id}/timeline`, data)
