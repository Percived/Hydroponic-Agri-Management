import { get } from './request'
import type { DashboardData } from '@/types'

// 获取仪表盘数据
export const getDashboardData = () => get<DashboardData>('/overview/dashboard')
