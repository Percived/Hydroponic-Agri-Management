import { get } from './request'
import type { MetricDefinition, MetricListResponse } from '@/types'

// 获取指标定义列表
export const getMetrics = (params?: Record<string, unknown>) =>
  get<MetricListResponse>('/metrics', params)

// 获取指标定义详情
export const getMetric = (id: number) =>
  get<MetricDefinition>(`/metrics/${id}`)
