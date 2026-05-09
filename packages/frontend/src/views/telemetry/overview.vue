<template>
  <div class="overview-page">
    <div class="page-header">
      <h1 class="page-title">实时总览</h1>
      <span v-if="sseConnected" class="sse-status connected">&#x25CF; 实时连接中</span>
      <span v-else class="sse-status disconnected">&#x25CB; 连接断开</span>
    </div>

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
    <div v-if="loading" class="card-grid">
      <el-skeleton v-for="i in 6" :key="i" animated>
        <template #template>
          <el-card class="sensor-card"><div style="height:100px" /></el-card>
        </template>
      </el-skeleton>
    </div>

    <!-- Empty states -->
    <el-empty v-else-if="!selectedGreenhouseId" description="请选择温室查看实时数据" />
    <el-empty v-else-if="devices.length === 0" description="所选范围内暂无传感器设备" />
    <el-empty v-else-if="snapshots.length === 0 && !loading" description="所选设备暂无通道配置" />

    <!-- Card grid -->
    <div v-else class="card-grid">
      <el-card
        v-for="snap in snapshots"
        :key="snap.channel_id"
        class="sensor-card"
        :class="{
          updated: updatedChannels.has(snap.channel_id),
          'border-online': snap.status === 'ONLINE',
          'border-offline': snap.status === 'OFFLINE',
          'border-fault': snap.status === 'FAULT'
        }"
      >
        <div class="card-header-row">
          <span class="card-device">{{ snap.device_name }}</span>
          <el-tag size="small" :type="qualityTagType(snap.quality_flag)">
            {{ snap.quality_flag || '-' }}
          </el-tag>
        </div>
        <div class="card-channel">{{ snap.channel_code }}</div>
        <div class="card-metric">{{ getMetricName(snap.metric_code) }}</div>
        <div class="card-value">
          {{ snap.latest_value !== null ? formatNumber(snap.latest_value) : '暂无数据' }}
          <span v-if="snap.latest_value !== null" class="card-unit">{{ snap.unit }}</span>
        </div>
        <div class="card-time">{{ snap.collected_at ? formatDate(snap.collected_at) : '-' }}</div>
      </el-card>
    </div>

    <!-- Trend chart (collapsible) -->
    <el-card v-if="snapshots.length > 0 && selectedMetricCodes.length > 0" class="chart-card">
      <template #header>
        <div class="chart-header">
          <span>趋势图（最近1小时）</span>
          <el-button text type="primary" @click="chartExpanded = !chartExpanded">
            {{ chartExpanded ? '收起' : '展开' }}
          </el-button>
        </div>
      </template>
      <div v-show="chartExpanded">
        <metric-trend-chart :series="chartSeries" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { deviceApi, greenhouseApi, telemetryApi } from '@/api'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { formatDate, formatNumber, getMetricName } from '@/utils/format'
import { useTelemetrySSE } from '@/composables/useTelemetrySSE'
import type { SensorDevice, SensorChannel, Greenhouse, GrowingZone, ChannelSnapshot } from '@/types'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'

const greenhouses = ref<Greenhouse[]>([])
const zones = ref<GrowingZone[]>([])
const devices = ref<SensorDevice[]>([])
const channels = ref<SensorChannel[]>([])
const selectedGreenhouseId = ref<number | null>(null)
const selectedZoneId = ref<number | null>(null)
const selectedDeviceIds = ref<number[]>([])
const selectedMetricCodes = ref<string[]>([])
const snapshots = ref<ChannelSnapshot[]>([])
const loading = ref(false)
const updatedChannels = ref(new Set<number>())
const chartExpanded = ref(true)

// SSE
const { connected: sseConnected, channelValues } = useTelemetrySSE()

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

const chartSeries = computed(() => {
  return Object.entries(trendBuffer.value)
    .filter(([, data]) => data.length > 0)
    .filter(([chIdStr]) => {
      const ch = channelMap.value.get(Number(chIdStr))
      return ch && selectedMetricCodes.value.includes(ch.metric_code)
    })
    .map(([chIdStr, data]) => {
      const chId = Number(chIdStr)
      const ch = channelMap.value.get(chId)
      const dev = ch ? deviceMap.value.get(ch.sensor_device_id) : undefined
      const label = ch
        ? `${dev?.name || '?'} / ${ch.channel_code} - ${getMetricName(ch.metric_code)}`
        : `CH#${chId}`
      return { name: label, data }
    })
})

// Watch SSE channelValues and update snapshots
watch(
  () => channelValues.value,
  (map) => {
    if (map.size === 0) return
    const flash = new Set<number>()

    snapshots.value = snapshots.value.map((snap) => {
      const evt = map.get(snap.channel_id)
      if (!evt) return snap
      flash.add(snap.channel_id)

      // Update trend buffer keyed by channel_id
      if (!trendBuffer.value[evt.sensor_channel_id]) {
        trendBuffer.value[evt.sensor_channel_id] = []
      }
      const buf = trendBuffer.value[evt.sensor_channel_id]
      buf.push({ time: evt.collected_at, value: evt.value })
      if (buf.length > MAX_BUFFER) buf.shift()

      return {
        ...snap,
        latest_value: evt.value,
        quality_flag: 'normal',
        collected_at: evt.collected_at
      }
    })

    if (flash.size > 0) {
      updatedChannels.value = flash
      setTimeout(() => {
        updatedChannels.value = new Set()
      }, 1500)
    }
  },
  { deep: true }
)

async function loadGreenhouses() {
  try {
    const result = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = result.items
  } catch { /* ignore */ }
}

async function onGreenhouseChange() {
  selectedZoneId.value = null
  selectedDeviceIds.value = []
  devices.value = []
  channels.value = []
  snapshots.value = []

  if (!selectedGreenhouseId.value) return

  try {
    const result = await greenhouseApi.getGreenhouseZones(selectedGreenhouseId.value)
    zones.value = result.items
  } catch { /* ignore */ }

  // Load all sensor devices for this greenhouse
  try {
    const result = await deviceApi.getSensorDevices({
      greenhouse_id: selectedGreenhouseId.value,
      page_size: LARGE_PAGE_SIZE
    })
    devices.value = result.items
  } catch { /* ignore */ }
}

async function onZoneChange() {
  selectedDeviceIds.value = []
  channels.value = []
  snapshots.value = []

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
    devices.value = result.items
  } catch { /* ignore */ }
}

async function onDeviceChange() {
  channels.value = []
  snapshots.value = []

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
    const allChannels = results.flatMap((r) => r.items)
    channels.value = allChannels

    // Build device and channel lookup
    deviceMap.value = new Map(devices.value.map((d) => [d.id, d]))
    channelMap.value = new Map(allChannels.map((c) => [c.id, c]))

    // Reset trend buffer
    trendBuffer.value = {}

    // Fetch latest values
    if (allChannels.length > 0) {
      const channelIds = allChannels.map((ch) => ch.id)
      try {
        const latest = await telemetryApi.getChannelsLatest(channelIds)
        const latestMap = new Map(latest.items.map((it) => [it.sensor_channel_id, it]))

        snapshots.value = allChannels.map((ch) => {
          const dev = deviceMap.value.get(ch.sensor_device_id)
          const lat = latestMap.get(ch.id)
          return {
            channel_id: ch.id,
            device_name: dev?.name || '-',
            device_code: dev?.device_code || '-',
            channel_code: ch.channel_code,
            metric_code: ch.metric_code,
            unit: ch.unit,
            latest_value: lat?.value ?? null,
            quality_flag: lat?.quality_flag || 'normal',
            collected_at: lat?.collected_at || '',
            status: dev?.status || 'OFFLINE'
          } as ChannelSnapshot
        })

        // Seed trend buffer from channel history (keyed by channel_id)
        const historyResults = await Promise.all(
          allChannels.map((ch) => {
            const end = new Date()
            const start = new Date(end.getTime() - 60 * 60 * 1000)
            return telemetryApi.getChannelHistory(ch.id, {
              start_time: start.toISOString(),
              end_time: end.toISOString(),
              page: 1,
              page_size: MAX_BUFFER
            }).then((history) => ({ chId: ch.id, items: history.items }))
              .catch(() => ({ chId: ch.id, items: [] }))
          })
        )
        for (const { chId, items } of historyResults) {
          if (items.length > 0) {
            trendBuffer.value[chId] = items
              .map((r) => ({ time: r.collected_at, value: r.value }))
              .reverse()
          }
        }
      } catch {
        // Even if latest fails, show cards with null values
        snapshots.value = allChannels.map((ch) => {
          const dev = deviceMap.value.get(ch.sensor_device_id)
          return {
            channel_id: ch.id,
            device_name: dev?.name || '-',
            device_code: dev?.device_code || '-',
            channel_code: ch.channel_code,
            metric_code: ch.metric_code,
            unit: ch.unit,
            latest_value: null,
            quality_flag: 'normal',
            collected_at: '',
            status: dev?.status || 'OFFLINE'
          } as ChannelSnapshot
        })
      }
    }
  } finally {
    loading.value = false
  }
}

function qualityTagType(flag: string): string {
  if (flag === 'normal') return 'success'
  if (flag === 'out_of_range' || flag === 'device_offline') return 'danger'
  if (flag === 'missing') return 'warning'
  return 'info'
}

onMounted(() => {
  loadGreenhouses()
})

onUnmounted(() => {
  // SSE disconnect handled by composable's onUnmounted via component lifecycle
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

  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 12px;
    margin-top: 12px;
  }

  .sensor-card {
    transition: box-shadow 0.3s ease, border-color 0.3s ease;
    border-left: 4px solid #dcdfe6;

    &.border-online {
      border-left-color: #67c23a;
    }
    &.border-offline {
      border-left-color: #c0c4cc;
    }
    &.border-fault {
      border-left-color: #f56c6c;
    }

    &.updated {
      box-shadow: 0 0 0 2px #409eff;
    }
  }

  .card-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }
  .card-device {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .card-channel {
    font-size: 12px;
    color: var(--color-text-secondary);
  }
  .card-metric {
    font-size: 12px;
    color: var(--color-text-regular);
    margin-bottom: 8px;
  }
  .card-value {
    font-size: 28px;
    font-weight: 700;
    line-height: 1.2;
  }
  .card-unit {
    font-size: 14px;
    font-weight: 400;
    color: var(--color-text-secondary);
  }
  .card-time {
    margin-top: 6px;
    font-size: 12px;
    color: var(--color-text-secondary);
  }

  .chart-card {
    margin-top: 16px;
  }
  .chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
