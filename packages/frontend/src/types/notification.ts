export enum ChannelType {
  EMAIL = 'EMAIL',
  SMS = 'SMS',
  WEBHOOK = 'WEBHOOK'
}

export const ChannelTypeNames: Record<string, string> = {
  EMAIL: '邮件',
  SMS: '短信',
  WEBHOOK: 'Webhook'
}

export interface NotificationChannel {
  id: number
  user_id: number
  channel_type: string
  name: string
  config: Record<string, any>
  min_alert_level: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateChannelRequest {
  channel_type: string
  name: string
  config: Record<string, any>
  min_alert_level?: string
  enabled: boolean
}

export interface UpdateChannelRequest {
  name?: string
  config?: Record<string, any>
  min_alert_level?: string
  enabled?: boolean
}
