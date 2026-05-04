import { ref, type Ref } from 'vue'

export interface AlertEvent {
  id: number
  type: string
  level: string
  metric_id: number | null
  device_id: number
  value: number | null
  message: string
  status: string
  triggered_at: string
  resolved_at: string | null
}

export interface UseAlertSSEOptions {
  deviceId?: number
  level?: string
}

export interface UseAlertSSEReturn {
  connected: Ref<boolean>
  lastAlert: Ref<AlertEvent | null>
  alertCount: Ref<number>
  connect: () => void
  disconnect: () => void
}

export function useAlertSSE(options?: UseAlertSSEOptions): UseAlertSSEReturn {
  const connected = ref(false)
  const lastAlert = ref<AlertEvent | null>(null)
  const alertCount = ref(0)

  let eventSource: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      reconnectDelay = Math.min(reconnectDelay * 2, 30000)
      connect()
    }, reconnectDelay)
  }

  function connect() {
    disconnect()

    const token = localStorage.getItem('hydroponic_token')
    if (!token) return

    const params = new URLSearchParams()
    params.set('token', token)
    if (options?.deviceId) params.set('device_id', String(options.deviceId))
    if (options?.level) params.set('level', options.level)

    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
    const url = `${baseURL}/alerts/subscribe?${params.toString()}`

    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      reconnectDelay = 1000
    }

    eventSource.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data)
        if (event.type === 'new_alert' && event.data) {
          const alert = event.data as AlertEvent
          lastAlert.value = alert
          alertCount.value++

          // Browser notification for CRITICAL alerts
          if (alert.level === 'CRITICAL' && Notification.permission === 'granted') {
            new Notification('告警通知', {
              body: alert.message,
              icon: '/favicon.ico',
              tag: String(alert.id)
            })
          }
        }
      } catch {
        // Ignore malformed JSON
      }
    }

    eventSource.onerror = () => {
      connected.value = false
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
  }

  return { connected, lastAlert, alertCount, connect, disconnect }
}
