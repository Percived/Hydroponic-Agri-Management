import { get } from './request'
import type { AuditLog, AuditLogQueryParams, PaginatedResponse } from '@/types'

// 获取审计日志列表
export const getAuditLogs = (params?: AuditLogQueryParams) =>
  get<PaginatedResponse<AuditLog>>('/audit-logs', params)
