<template>
  <div class="realtime-page">
    <div class="page-header">
      <h1 class="page-title">实时曲线</h1>
      <span v-if="sseConnected" class="sse-status connected">&#x25CF; 实时连接中</span>
      <span v-else class="sse-status disconnected">&#x25CB; 连接断开</span>
    </div>

    <div class="filter-section">
      <el-select v-model="selectedDeviceId" placeholder="选择传感器设备" filterable style="width: 280px" @change="onDeviceChange">
        <el-option v-for="device in sensorDevices" :key="device.id" :label="`${device.name} (${device.device_code})`" :value="device.id" />
      </el-select>
      <el-select v-model="selectedChannelId" placeholder="选择通道" filterable style="width: 220px" @change="onChannelChange">
        <el-option v-for="ch in deviceChannels" :key="ch.id" :label="`${ch.channel_code} - ${getMetricName(ch.metric_code)}`" :value="ch.id" />
      </el-select>
      <el-select v-model="selectedMetricForChart" placeholder="选择指标" style="width: 180px">
        <el-option v-for="item in metricOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-button type="primary" :loading="loading" @click="fetchInitialData">刷新数据</el-button>
    </div>

    <div v-if="selectedChannelId" class="data-section">
      <div v-if="loading && filteredTelemetry.length === 0" class="loading-placeholder">
        <el-skeleton :rows="5" animated />
      </div>
      <template v-else>
        <div v-if="filteredTelemetry.length === 0" class="empty-placeholder">
          <el-empty description="当前筛选条件下暂无数据" />
        </div>
        <template v-else>
          <div class="metrics-grid">
            <el-card v-for="item in filteredTelemetry" :key="item.metric_code" class="metric-card" :class="{ updated: updatedMetrics.has(item.metric_code) }">
              <div class="metric-header">
                <span class="metric-name">{{ getMetricName(item.metric_code) }}</span>
                <el-tag :type="item.quality_flag === 'normal' ? 'success' : 'danger'" size="small">
                  {{ item.quality_flag }}
                </el-tag>
              </div>
              <div class="metric-value">
                {{ formatNumber(item.value) }}
              </div>
              <div class="metric-time">{{ formatDate(item.collected_at) }}</div>
            </el-card>
          </div>

          <el-card class="chart-card" v-if="selectedMetricForChart">
            <template #header>
              <div class="chart-header">
                <span>趋势图（最近1小时）</span>
              </div>
            </template>
            <metric-trend-chart :series="chartSeries" :y-axis-name="''" />
          </el-card>
        </template>
      </template>
    </div>

    <el-empty v-else description="请选择设备和通道查看实时数据" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { deviceApi, telemetryApi } from '@/api'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { formatDate, formatNumber, getMetricName } from '@/utils/format'
import type { SensorDevice, SensorChannel, TelemetryRecord } from '@/types'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'

const sensorDevices = ref<SensorDevice[]>([])
const deviceChannels = ref<SensorChannel[]>([])
const selectedDeviceId = ref<number | null>(null)
const selectedChannelId = ref<number | null>(null)
const selectedMetricForChart = ref('')
const telemetryData = ref<TelemetryRecord[]>([])
const chartHistory = ref<TelemetryRecord[]>([])
const loading = ref(false)
const updatedMetrics = ref(new Set<string>())
const sseConnected = ref(false)

const metricOptions = computed(() => {
  const set = new Set(telemetryData.value.map((i) => i.metric_code))
  return [...set].map((code) => ({ value: code, label: getMetricName(code) }))
})

const filteredTelemetry = computed(() => {
  if (!selectedMetricForChart.value) return telemetryData.value
  return telemetryData.value.filter((item) => item.metric_code === selectedMetricForChart.value)
})

const chartSeries = computed(() => [
  {
    name: getMetricName(selectedMetricForChart.value),
    data: chartHistory.value.map((i) => ({ time: i.collected_at, value: i.value }))
  }
])

async function loadSensorDevices() {
  try {
    const result = await deviceApi.getSensorDevices({ page_size: LARGE_PAGE_SIZE })
    sensorDevices.value = result.items
  } catch { /* ignore */ }
}

async function onDeviceChange() {
  selectedChannelId.value = null
  telemetryData.value = []
  chartHistory.value = []

  if (!selectedDeviceId.value) {
    deviceChannels.value = []
    return
  }

  try {
    const result = await deviceApi.getSensorChannels({
      sensor_device_id: selectedDeviceId.value,
      page_size: LARGE_PAGE_SIZE
    })
    deviceChannels.value = result.items
  } catch { /* ignore */ }
}

async function onChannelChange() {
  if (selectedChannelId.value) {
    fetchInitialData()
  }
}

async function fetchInitialData() {
  if (!selectedChannelId.value) return
  loading.value = true
  try {
    const latest = await telemetryApi.getChannelLatest(selectedChannelId.value)
    telemetryData.value = [latest]
    if (latest.metric_code) {
      selectedMetricForChart.value = latest.metric_code
    }

    // Fetch history for chart
    const end = new Date()
    const start = new Date(end.getTime() - 60 * 60 * 1000)
    const history = await telemetryApi.getChannelHistory(selectedChannelId.value, {
      start_time: start.toISOString(),
      end_time: end.toISOString(),
      page: 1,
      page_size: LARGE_PAGE_SIZE
    })
    chartHistory.value = history.items
  } finally {
    loading.value = false
  }
}

// Set up polling for real-time updates (no SSE, use polling as fallback)
let pollingTimer: ReturnType<typeof setInterval> | null = null

function startPolling() {
  stopPolling()
  pollingTimer = setInterval(() => {
    if (selectedChannelId.value) {
      telemetryApi.getChannelLatest(selectedChannelId.value).then((record) => {
        const idx = telemetryData.value.findIndex((it) => it.metric_code === record.metric_code)
        if (idx >= 0) {
          telemetryData.value[idx] = record
        } else {
          telemetryData.value.push(record)
        }
        updatedMetrics.value = new Set(updatedMetrics.value).add(record.metric_code)
        sseConnected.value = true
        setTimeout(() => {
          const next = new Set(updatedMetrics.value)
          next.delete(record.metric_code)
          updatedMetrics.value = next
        }, 1000)
      }).catch(() => {
        sseConnected.value = false
      })
    }
  }, 5000)
  sseConnected.value = true
}

function stopPolling() {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
}

onMounted(() => {
  loadSensorDevices()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped lang="scss">
.realtime-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .page-title {
    font-size: 22px;
    font-weight: 700;
    margin: 0;
  }

  .sse-status {
    font-size: 13px;
    padding: 4px 10px;
    border-radius: 12px;
    &.connected {
      color: var(--color-primary);
      background: var(--color-primary-bg);
    }
    &.disconnected {
      color: var(--color-text-secondary);
      background: var(--bg-page);
    }
  }

  .filter-section {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
    margin: 12px 0;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 12px;
    margin-top: 12px;
    margin-bottom: 16px;
  }

  .metric-card {
    transition: box-shadow 0.2s ease;
    &.updated {
      box-shadow: 0 0 0 2px var(--color-primary);
    }
  }

  .metric-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
  }
  .metric-name {
    font-size: 13px;
    color: var(--color-text-regular);
  }
  .metric-value {
    font-size: 28px;
    font-weight: 700;
  }
  .metric-time {
    margin-top: 6px;
    font-size: 12px;
    color: var(--color-text-secondary);
  }
}
</style>
