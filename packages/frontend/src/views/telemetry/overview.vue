<template>
  <div class="overview-page">
    <div class="page-header">
      <h1 class="page-title">实时总览</h1>
      <div class="header-right">
        <span class="sse-status" :class="sseStatusClass">{{ sseStatusText }}</span>
        <el-button v-if="sseStatus !== 'connected'" text type="primary" @click="reconnectSSE">重连</el-button>
      </div>
    </div>
    <div v-if="sseLastError" class="sse-error">{{ sseLastError }}</div>
    <el-alert v-if="pageError" :title="pageError" type="error" show-icon closable style="margin-bottom: 12px" @close="pageError = ''">
      <template #default>
        <el-button text type="primary" @click="retryLoad">重试</el-button>
      </template>
    </el-alert>

    <div class="filter-section">
      <el-select v-model="selectedGreenhouseId" placeholder="选择温室" filterable style="width: 200px" @change="onGreenhouseChange">
        <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
      </el-select>
      <el-select v-model="selectedZoneId" placeholder="选择种植区" filterable style="width: 180px" :disabled="!selectedGreenhouseId" @change="onZoneChange">
        <el-option v-for="zone in zones" :key="zone.id" :label="zone.name" :value="zone.id" />
        <el-option v-if="zones.length === 0 && selectedGreenhouseId" :value="0" label="(全部种植区)" />
      </el-select>
      <el-select v-model="selectedDeviceIds" placeholder="选择设备（可多选）" filterable multiple collapse-tags collapse-tags-tooltip style="width: 280px" :disabled="!selectedGreenhouseId" @change="onDeviceChange">
        <el-option v-for="dev in devices" :key="dev.id" :label="`${dev.name} (${dev.device_code})`" :value="dev.id" />
      </el-select>
      <el-select v-model="selectedMetricCodes" placeholder="图表指标（可多选）" filterable multiple collapse-tags collapse-tags-tooltip style="width: 220px">
        <el-option v-for="m in metricOptions" :key="m.value" :label="m.label" :value="m.value" />
      </el-select>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="chart-grid">
      <div v-for="i in 4" :key="i" class="chart-card-item">
        <el-card>
          <template #header>
            <el-skeleton :rows="1" animated />
          </template>
          <div class="chart-placeholder">
            <el-skeleton :rows="8" animated />
          </div>
        </el-card>
      </div>
    </div>

    <!-- Empty states -->
    <el-empty v-else-if="!selectedGreenhouseId" description="请选择温室查看实时数据" />
    <el-empty v-else-if="devices.length === 0" description="所选范围内暂无传感器设备" />
    <el-empty v-else-if="channels.length === 0 && !loading" description="所选设备暂无通道配置" />

    <!-- Time range selector + Chart grid -->
    <template v-else>
      <div class="time-range-bar">
        <el-radio-group v-model="timeRangePreset" @change="fetchInitialHistory">
          <el-radio-button value="1h">1小时</el-radio-button>
          <el-radio-button value="6h">6小时</el-radio-button>
          <el-radio-button value="24h">24小时</el-radio-button>
          <el-radio-button value="3d">3天</el-radio-button>
          <el-radio-button value="7d">7天</el-radio-button>
        </el-radio-group>
      </div>

      <div class="chart-grid">
        <div v-for="item in metricCharts" :key="item.metricCode" class="chart-card-item">
          <el-card shadow="hover">
            <template #header>
              <div class="chart-card-header">
                <span class="chart-card-title">{{ item.metricName }}</span>
                <el-tag size="small" type="info">{{ item.seriesCount }} / {{ item.channelCount }} 通道</el-tag>
              </div>
            </template>
            <div v-if="chartLoading" class="chart-placeholder">
              <el-skeleton :rows="8" animated />
            </div>
            <el-empty v-else-if="item.series.length === 0" description="暂无数据" />
            <MetricTrendChart v-else :series="item.series" :y-axis-name="item.unit" />
          </el-card>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { deviceApi, greenhouseApi, telemetryApi } from '@/api'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { getMetricName } from '@/utils/format'
import { useTelemetrySSE } from '@/composables/useTelemetrySSE'
import { ElMessage } from 'element-plus'
import type { SensorDevice, SensorChannel, Greenhouse, GrowingZone } from '@/types'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'
import { appendTrendPoint, normalizeTrendPoints } from './trendBuffer'

const greenhouses = ref<Greenhouse[]>([])
const zones = ref<GrowingZone[]>([])
const devices = ref<SensorDevice[]>([])
const channels = ref<SensorChannel[]>([])
const selectedGreenhouseId = ref<number | null>(null)
const selectedZoneId = ref<number | null>(null)
const selectedDeviceIds = ref<number[]>([])
const selectedMetricCodes = ref<string[]>([])
const loading = ref(false)
const chartLoading = ref(false)
const timeRangePreset = ref('1h')
const pageError = ref('')
let loadSeq = 0
let historySeq = 0

// SSE
const selectedDeviceCodes = computed(() => {
  if (selectedDeviceIds.value.length === 0) return []
  const idSet = new Set(selectedDeviceIds.value)
  return devices.value.filter(d => idSet.has(d.id)).map(d => d.device_code)
})
const {
  status: sseStatus,
  lastError: sseLastError,
  channelValues,
  connect,
  disconnect
} = useTelemetrySSE({ deviceCodes: selectedDeviceCodes, metricCodes: selectedMetricCodes })

const sseStatusText = computed(() => {
  if (sseStatus.value === 'connected') return '● 实时已连接'
  if (sseStatus.value === 'connecting') return '○ 正在连接'
  if (sseStatus.value === 'error') return '○ 连接失败'
  return '○ 未连接'
})
const sseStatusClass = computed(() => {
  if (sseStatus.value === 'connected') return 'connected'
  if (sseStatus.value === 'connecting') return 'connecting'
  if (sseStatus.value === 'error') return 'error'
  return 'disconnected'
})

function reconnectSSE() {
  disconnect()
  connect()
}

// Trend buffer: channel_id -> [time, value]
const trendBuffer = ref<Record<number, Array<{ time: string; value: number }>>>({})
const MAX_BUFFER = 360

// Lookup maps for building chart series labels
const deviceMap = ref<Map<number, SensorDevice>>(new Map())
const channelMap = ref<Map<number, SensorChannel>>(new Map())

const metricOptions = computed(() => {
  const seen = new Set<string>()
  for (const ch of channels.value) {
    if (!seen.has(ch.metric_code)) {
      seen.add(ch.metric_code)
    }
  }
  return [...seen].map((code) => ({ value: code, label: getMetricName(code) }))
})

// Group trendBuffer by metric_code → one chart card per metric, one series per channel
interface MetricChartItem {
  metricCode: string
  metricName: string
  channelCount: number
  seriesCount: number
  series: Array<{ name: string; data: Array<{ time: string; value: number }> }>
  unit: string
}

const metricCharts = computed<MetricChartItem[]>(() => {
  // Group channels by metric_code
  const metricGroups = new Map<string, SensorChannel[]>()
  for (const ch of channels.value) {
    if (!selectedMetricCodes.value.includes(ch.metric_code)) continue
    if (!metricGroups.has(ch.metric_code)) metricGroups.set(ch.metric_code, [])
    metricGroups.get(ch.metric_code)!.push(ch)
  }

  return [...metricGroups.entries()].map(([code, chs]) => {
    const series: Array<{ name: string; data: Array<{ time: string; value: number }> }> = []
    for (const ch of chs) {
      const buf = trendBuffer.value[ch.id]
      if (buf && buf.length > 0) {
        const dev = deviceMap.value.get(ch.sensor_device_id)
        series.push({
          name: `${dev?.name || '?'} / ${ch.channel_code}`,
          data: buf
        })
      }
    }

    return {
      metricCode: code,
      metricName: getMetricName(code),
      channelCount: chs.length,
      seriesCount: series.length,
      series,
      unit: chs[0]?.unit || ''
    }
  })
})

// SSE watch: update trendBuffer in real-time
watch(
  () => channelValues.value,
  (map) => {
    if (map.size === 0) return
    for (const [chIdStr, evt] of map) {
      const chId = Number(chIdStr)
      // Only buffer if this channel belongs to selected devices
      if (!channelMap.value.has(chId)) continue
      const buf = trendBuffer.value[chId] || []
      trendBuffer.value[chId] = appendTrendPoint(
        buf,
        { time: evt.collected_at, value: evt.value },
        MAX_BUFFER
      )
    }
  }
)

function getTimeRange(): { start: string; end: string } {
  const now = new Date()
  const end = now.toISOString()
  let start: Date
  switch (timeRangePreset.value) {
    case '1h': start = new Date(now.getTime() - 60 * 60 * 1000); break
    case '6h': start = new Date(now.getTime() - 6 * 60 * 60 * 1000); break
    case '24h': start = new Date(now.getTime() - 24 * 60 * 60 * 1000); break
    case '3d': start = new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000); break
    case '7d': start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000); break
    default: start = new Date(now.getTime() - 60 * 60 * 1000)
  }
  return { start: start.toISOString(), end }
}

async function fetchInitialHistory() {
  const allChannels = channels.value
  if (allChannels.length === 0) return

  chartLoading.value = true
  const { start, end } = getTimeRange()
  const seq = ++historySeq

  try {
    const results = await Promise.all(
      allChannels.map((ch) =>
        telemetryApi.getChannelHistory(ch.id, {
          start_time: start,
          end_time: end,
          page: 1,
          page_size: MAX_BUFFER
        }).then(history => ({ chId: ch.id, items: history.items }))
          .catch(() => ({ chId: ch.id, items: [] }))
      )
    )
    if (seq !== historySeq) return

    const newBuffer: Record<number, Array<{ time: string; value: number }>> = {}
    for (const { chId, items } of results) {
      if (items.length > 0) {
        newBuffer[chId] = normalizeTrendPoints(
          items.map(r => ({ time: r.collected_at, value: r.value })),
          MAX_BUFFER
        )
      }
    }
    trendBuffer.value = newBuffer
  } finally {
    if (seq === historySeq) chartLoading.value = false
  }
}

async function loadGreenhouses() {
  const seq = ++loadSeq
  pageError.value = ''
  try {
    const result = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    if (seq !== loadSeq) return
    greenhouses.value = result.items
  } catch {
    if (seq !== loadSeq) return
    pageError.value = '温室列表加载失败'
    ElMessage.error(pageError.value)
  }
}

async function onGreenhouseChange() {
  const seq = ++loadSeq
  pageError.value = ''
  selectedZoneId.value = null
  selectedDeviceIds.value = []
  devices.value = []
  channels.value = []

  if (!selectedGreenhouseId.value) return

  try {
    const result = await greenhouseApi.getGreenhouseZones(selectedGreenhouseId.value)
    if (seq !== loadSeq) return
    zones.value = result.items
  } catch {
    if (seq !== loadSeq) return
    pageError.value = '种植区列表加载失败'
    ElMessage.error(pageError.value)
  }

  // Load all sensor devices for this greenhouse
  try {
    const result = await deviceApi.getSensorDevices({
      greenhouse_id: selectedGreenhouseId.value,
      page_size: LARGE_PAGE_SIZE
    })
    if (seq !== loadSeq) return
    devices.value = result.items
  } catch {
    if (seq !== loadSeq) return
    pageError.value = '设备列表加载失败'
    ElMessage.error(pageError.value)
  }
}

async function onZoneChange() {
  const seq = ++loadSeq
  pageError.value = ''
  selectedDeviceIds.value = []
  channels.value = []

  if (!selectedGreenhouseId.value) return

  const params: Record<string, unknown> = {
    greenhouse_id: selectedGreenhouseId.value,
    page_size: LARGE_PAGE_SIZE
  }
  if (selectedZoneId.value) {
    params.growing_zone_id = selectedZoneId.value
  }

  try {
    const result = await deviceApi.getSensorDevices(params)
    if (seq !== loadSeq) return
    devices.value = result.items
  } catch {
    if (seq !== loadSeq) return
    pageError.value = '设备列表加载失败'
    ElMessage.error(pageError.value)
  }
}

async function onDeviceChange() {
  const seq = ++loadSeq
  pageError.value = ''
  channels.value = []
  trendBuffer.value = {}

  if (selectedDeviceIds.value.length === 0) return

  // Load channels for all selected devices (parallel)
  loading.value = true
  try {
    const results = await Promise.all(
      selectedDeviceIds.value.map((deviceId) =>
        deviceApi.getSensorChannels({
          sensor_device_id: deviceId,
          page_size: LARGE_PAGE_SIZE,
          enabled: 1
        }).catch(() => ({ items: [] as SensorChannel[] }))
      )
    )
    if (seq !== loadSeq) return
    const allChannels = results.flatMap((r) => r.items)
    channels.value = allChannels

    // Build device and channel lookup
    deviceMap.value = new Map(devices.value.map((d) => [d.id, d]))
    channelMap.value = new Map(allChannels.map((c) => [c.id, c]))

    // Auto-select all available metrics
    const codes = [...new Set(allChannels.map(ch => ch.metric_code))]
    selectedMetricCodes.value = codes

    // Fetch initial history
    if (allChannels.length > 0) {
      await fetchInitialHistory()
    }
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

async function retryLoad() {
  if (selectedDeviceIds.value.length > 0) {
    await onDeviceChange()
    return
  }
  if (selectedGreenhouseId.value) {
    if (selectedZoneId.value) {
      await onZoneChange()
      return
    }
    await onGreenhouseChange()
    return
  }
  await loadGreenhouses()
}

onMounted(() => {
  loadGreenhouses()
  connect()
})

onBeforeUnmount(() => {
  disconnect()
})
</script>

<style scoped lang="scss">
.overview-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .sse-error {
    font-size: 12px;
    color: #f56c6c;
    margin-bottom: 8px;
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
      color: #67c23a;
      background: #f0f9eb;
    }
    &.connecting {
      color: #e6a23c;
      background: #fdf6ec;
    }
    &.error {
      color: #f56c6c;
      background: #fef0f0;
    }
    &.disconnected {
      color: #909399;
      background: #f4f4f5;
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

  .time-range-bar {
    margin: 12px 0;
  }

  .chart-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
    margin-top: 12px;
  }

  .chart-card-item {
    min-width: 0;
  }

  .chart-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .chart-card-title {
    font-size: 15px;
    font-weight: 600;
  }

  .chart-placeholder {
    padding: 12px 0;
  }
}
</style>
