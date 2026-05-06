import type { PaginatedResponse } from './api'

// ── Control Commands ──
export interface ControlCommand {
  id: number
  actuator_channel_id: number
  command_type: string
  payload: Record<string, unknown>
  status: 'PENDING' | 'QUEUED' | 'SENT' | 'ACKED' | 'TIMEOUT' | 'FAILED'
  sent_at?: string
  acked_at?: string
  request_id?: string
  created_by: number
  created_at: string
  receipts?: ControlCommandReceipt[]
}

export interface ControlCommandReceipt {
  id: number
  command_id: number
  receipt_seq: number
  receipt_status: string
  ack_code?: string
  ack_message?: string
  ack_payload?: Record<string, unknown>
  ack_at?: string
  created_at: string
}

export interface CreateCommandRequest {
  actuator_channel_id: number
  command_type: string
  payload: Record<string, unknown>
}

export interface CommandListResponse extends PaginatedResponse<ControlCommand> {}
