import { get, post, put, del } from './request'
import type { NotificationChannel, CreateChannelRequest, UpdateChannelRequest } from '@/types'

export function getChannels(): Promise<{ items: NotificationChannel[] }> {
  return get<{ items: NotificationChannel[] }>('/notification-channels')
}

export function createChannel(data: CreateChannelRequest): Promise<{ id: number }> {
  return post<{ id: number }>('/notification-channels', data)
}

export function updateChannel(id: number, data: UpdateChannelRequest): Promise<void> {
  return put<void>(`/notification-channels/${id}`, data)
}

export function deleteChannel(id: number): Promise<void> {
  return del<void>(`/notification-channels/${id}`)
}

export function testChannel(id: number): Promise<{ sent: boolean }> {
  return post<{ sent: boolean }>(`/notification-channels/${id}/test`)
}
