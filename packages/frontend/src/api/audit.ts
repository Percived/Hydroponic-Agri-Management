import { get } from './request'
import type { AuditLog, AuditLogQueryParams, PaginatedData } from '@/types'

// 获取审计日志列表
export const getAuditLogs = (params?: AuditLogQueryParams) =>
  get<PaginatedData<AuditLog>>('/audit-logs', params)
