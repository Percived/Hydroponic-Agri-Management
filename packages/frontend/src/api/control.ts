import { get, post, put } from './request'
import type {
  ControlCommand,
  CommandListResponse,
  CreateCommandRequest,
  ControlCommandReceipt
} from '@/types'

// 获取命令列表
export const getCommands = (params?: Record<string, unknown>) =>
  get<CommandListResponse>('/commands', params)

// 获取命令详情
export const getCommand = (id: number) =>
  get<ControlCommand>(`/commands/${id}`)

// 创建命令
export const createCommand = (data: CreateCommandRequest) =>
  post<ControlCommand>('/commands', data)

// 发送命令
export const sendCommand = (id: number) =>
  put<ControlCommand>(`/commands/${id}/send`)

// 确认命令
export const ackCommand = (id: number, data?: Record<string, unknown>) =>
  post<ControlCommand>(`/commands/${id}/ack`, data)

// 获取命令回执列表
export const getCommandReceipts = (id: number) =>
  get<{ items: ControlCommandReceipt[] }>(`/commands/${id}/receipts`)

// 创建命令回执
export const createCommandReceipt = (id: number, data: {
  receipt_seq: number
  receipt_status: string
  ack_code?: string
  ack_message?: string
  ack_payload?: Record<string, unknown>
  ack_at?: string
}) =>
  post<{ id: number }>(`/commands/${id}/receipts`, data)
