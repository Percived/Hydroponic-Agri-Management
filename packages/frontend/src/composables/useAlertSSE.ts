import { ref, type Ref } from 'vue'
import type { Alert } from '@/types/alert'

export interface UseAlertSSEOptions {
  deviceCode?: string
  level?: string
}

export interface UseAlertSSEReturn {
  connected: Ref<boolean>
  lastAlert: Ref<Alert | null>
  alertCount: Ref<number>
  connect: () => void
  disconnect: () => void
}

export function useAlertSSE(options?: UseAlertSSEOptions): UseAlertSSEReturn {
  const connected = ref(false)
  const lastAlert = ref<Alert | null>(null)
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
    if (options?.deviceCode) params.set('device_codes', options.deviceCode)
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
          const alert = event.data as Alert
          if (alert.schema_version !== 1) {
            connected.value = false
            eventSource?.close()
            eventSource = null
            return
          }
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
