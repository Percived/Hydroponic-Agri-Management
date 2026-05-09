import { ref, watch, type Ref } from 'vue'
import type { TelemetrySSEEvent } from '@/types'

export interface TelemetryMetric {
  code: string
  value: number
  unit?: string
}

export interface TelemetryUpdate {
  device_code: string
  collected_at: string
  metrics: TelemetryMetric[]
}

export interface UseTelemetrySSEOptions {
  deviceCodes?: Ref<string[]>
  metricCodes?: Ref<string[]>
}

export interface UseTelemetrySSEReturn {
  connected: Ref<boolean>
  status: Ref<'disconnected' | 'connecting' | 'connected' | 'error'>
  lastError: Ref<string>
  latestUpdate: Ref<TelemetryUpdate | null>
  channelValues: Ref<Map<number, TelemetrySSEEvent>>
  connect: () => void
  disconnect: () => void
}

export function useTelemetrySSE(options?: UseTelemetrySSEOptions): UseTelemetrySSEReturn {
  const connected = ref(false)
  const status = ref<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected')
  const lastError = ref('')
  const latestUpdate = ref<TelemetryUpdate | null>(null)
  const channelValues = ref<Map<number, TelemetrySSEEvent>>(new Map())

  let eventSource: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000

  function buildURL(): string | null {
    const token = localStorage.getItem('hydroponic_token')
    if (!token) return null

    const params = new URLSearchParams()
    params.set('token', token)

    const deviceCodes = options?.deviceCodes?.value || []
    const metricCodes = options?.metricCodes?.value || []
    if (deviceCodes.length > 0) params.set('device_codes', deviceCodes.join(','))
    if (metricCodes.length > 0) params.set('metric_codes', metricCodes.join(','))

    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
    return `${baseURL}/telemetry/subscribe?${params.toString()}`
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    status.value = 'connecting'
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      reconnectDelay = Math.min(reconnectDelay * 2, 30000)
      connect()
    }, reconnectDelay)
  }

  function connect() {
    disconnect()

    const url = buildURL()
    if (!url) {
      status.value = 'error'
      lastError.value = '未登录或缺少 token'
      return
    }

    status.value = 'connecting'
    lastError.value = ''
    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      status.value = 'connected'
      reconnectDelay = 1000
    }

    eventSource.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data)
        if (event.type === 'telemetry_update' && event.data) {
          // Backend SSE data format: { sensor_channel_id, metric_code, value, collected_at, device_code }
          const sseEvent = event.data as TelemetrySSEEvent

          // Update channelValues map by sensor_channel_id for O(1) card matching
          if (sseEvent.sensor_channel_id) {
            const next = new Map(channelValues.value)
            next.set(sseEvent.sensor_channel_id, sseEvent)
            channelValues.value = next
          }

          // Keep latestUpdate for backward compatibility
          latestUpdate.value = {
            device_code: sseEvent.device_code || '',
            collected_at: sseEvent.collected_at,
            metrics: [{ code: sseEvent.metric_code, value: sseEvent.value }]
          }
        }
      } catch {
        // Ignore malformed JSON
      }
    }

    eventSource.onerror = () => {
      connected.value = false
      status.value = 'error'
      lastError.value = '连接失败，正在重试'
      eventSource?.close()
      eventSource = null
      scheduleReconnect()
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connected.value = false
    status.value = 'disconnected'
  }

  // If deviceCodes or metricCodes are reactive refs, reconnect when they change
  if (options?.deviceCodes) {
    watch(options.deviceCodes, () => connect(), { deep: true })
  }
  if (options?.metricCodes) {
    watch(options.metricCodes, () => connect(), { deep: true })
  }

  return { connected, status, lastError, latestUpdate, channelValues, connect, disconnect }
}

/** Request browser notification permission. Call on user interaction. */
export function requestNotificationPermission(): void {
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
}
