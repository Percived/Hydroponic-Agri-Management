<template>
  <div class="trends-page">
    <div class="page-header">
      <h1 class="page-title">趋势分析</h1>
    </div>

    <!-- Filter bar -->
    <div class="filter-section">
      <div class="filter-row">
        <el-select v-model="selectedGreenhouseId" placeholder="选择温室" filterable style="width: 180px" @change="onGreenhouseChange">
          <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
        </el-select>
        <el-select v-model="selectedZoneId" placeholder="选择种植区" filterable style="width: 160px" :disabled="!selectedGreenhouseId" @change="onZoneChange">
          <el-option v-for="zone in zones" :key="zone.id" :label="zone.name" :value="zone.id" />
          <el-option v-if="zones.length === 0 && selectedGreenhouseId" :value="0" label="(全部种植区)" />
        </el-select>
        <el-select
          v-model="selectedChannelIds"
          placeholder="选择通道（可多选）"
          filterable
          multiple
          collapse-tags
          collapse-tags-tooltip
          style="width: 320px"
          :disabled="!selectedGreenhouseId"
        >
          <el-option
            v-for="ch in channelOptions"
            :key="ch.value"
            :label="ch.label"
            :value="ch.value"
          />
        </el-select>
      </div>
      <div class="filter-row">
        <el-select v-model="selectedMetricCodes" placeholder="指标（可多选）" filterable multiple collapse-tags collapse-tags-tooltip style="width: 200px">
          <el-option v-for="m in metricList" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
        </el-select>
        <el-select v-model="timeRangePreset" placeholder="时间范围" style="width: 140px" @change="onPresetChange">
          <el-option label="最近1小时" value="1h" />
          <el-option label="最近3小时" value="3h" />
          <el-option label="最近6小时" value="6h" />
          <el-option label="最近12小时" value="12h" />
          <el-option label="最近24小时" value="24h" />
          <el-option label="最近3天" value="3d" />
          <el-option label="最近7天" value="7d" />
          <el-option label="自定义" value="custom" />
        </el-select>
        <el-date-picker
          v-if="timeRangePreset === 'custom'"
          v-model="customTimeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          style="width: 360px"
        />
        <el-select v-model="selectedBatchId" placeholder="批次（可选）" filterable clearable style="width: 180px">
          <el-option v-for="b in batches" :key="b.id" :label="b.batch_no || `批次#${b.id}`" :value="b.id" />
        </el-select>
        <el-select v-model="qualityFilter" placeholder="质量标识" clearable style="width: 120px">
          <el-option label="正常" value="normal" />
          <el-option label="缺失" value="missing" />
          <el-option label="超出范围" value="out_of_range" />
          <el-option label="设备离线" value="device_offline" />
        </el-select>
        <el-button type="primary" :loading="querying" @click="doQuery">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>
    </div>

    <!-- Empty state: no query executed -->
    <el-empty v-if="!hasQueried" description="请选择通道和指标后查询" />

    <!-- Loading -->
    <template v-if="querying">
      <el-skeleton :rows="3" animated style="margin-top: 16px" />
      <el-skeleton :rows="8" animated style="margin-top: 12px" />
    </template>

    <!-- Error state -->
    <el-alert v-if="queryError" :title="queryError" type="error" show-icon closable style="margin-top: 16px" @close="queryError = ''">
      <template #default>
        <el-button text type="primary" @click="doQuery">重试</el-button>
      </template>
    </el-alert>

    <!-- No data -->
    <el-empty v-if="hasQueried && !querying && chartSeries.length === 0 && tableData.length === 0" description="当前筛选条件下暂无数据" />

    <!-- Results -->
    <template v-if="hasQueried && !querying && (chartSeries.length > 0 || tableData.length > 0)">
      <!-- Stats Summary -->
      <el-row :gutter="12" class="stats-row" v-if="statsByMetric.size > 0">
        <el-col v-for="[code, stat] in statsByMetric" :key="code" :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-title">{{ getMetricName(code) }}</div>
            <div class="stat-values">
              <span class="stat-item">均值 <strong>{{ formatNumber(stat.avg) }}</strong></span>
              <span class="stat-item">最大 <strong>{{ formatNumber(stat.max) }}</strong></span>
              <span class="stat-item">最小 <strong>{{ formatNumber(stat.min) }}</strong></span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- Chart area -->
      <el-row :gutter="12" class="chart-area">
        <el-col :span="selectedBatchId ? 16 : 24">
          <el-card>
            <template #header><span>多指标对比</span></template>
            <metric-trend-chart :series="chartSeries" :events="eventPoints" />
          </el-card>
        </el-col>
        <el-col v-if="selectedBatchId" :span="8">
          <el-card>
            <template #header><span>批次事件</span></template>
            <batch-event-overlay :events="timelineEvents" />
          </el-card>
        </el-col>
      </el-row>

      <!-- Data table -->
      <el-card class="table-card">
        <template #header><span>数据明细</span></template>
        <el-table :data="pagedTableData" v-loading="querying" stripe>
          <el-table-column prop="collected_at" label="采集时间" width="170" :formatter="(_r: any, _c: any, val: string) => formatDate(val)" />
          <el-table-column prop="device_name" label="设备名" width="140" />
          <el-table-column prop="channel_code" label="通道码" width="100" />
          <el-table-column prop="metric_code" label="指标" width="100" :formatter="(_r: any, _c: any, val: string) => getMetricName(val)" />
          <el-table-column prop="value" label="数值" width="120" :formatter="(_r: any, _c: any, val: number) => formatNumber(val)" />
          <el-table-column prop="quality_flag" label="质量标识" width="100">
            <template #default="{ row }">
              <el-tag :type="row.quality_flag === 'normal' ? 'success' : 'danger'" size="small">
                {{ row.quality_flag }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-pagination">
          <el-pagination
            v-model:current-page="tablePage"
            :page-size="tablePageSize"
            :total="tableData.length"
            layout="total, prev, pager, next"
          />
        </div>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { deviceApi, greenhouseApi, telemetryApi, metricApi, cropApi } from '@/api'
import { LARGE_PAGE_SIZE, EXTRA_LARGE_PAGE_SIZE } from '@/utils/constants'
import { formatDate, formatNumber, getMetricName, populateMetricNames } from '@/utils/format'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'
import BatchEventOverlay from '@/components/charts/BatchEventOverlay.vue'
import type {
  Greenhouse, GrowingZone, SensorDevice, SensorChannel,
  MetricDefinition, TelemetryRecord
} from '@/types'

interface TableRow {
  collected_at: string
  device_name: string
  channel_code: string
  metric_code: string
  value: number
  quality_flag: string
}

// Cascade selects
const greenhouses = ref<Greenhouse[]>([])
const zones = ref<GrowingZone[]>([])
const devices = ref<SensorDevice[]>([])
const channels = ref<SensorChannel[]>([])
const selectedGreenhouseId = ref<number | null>(null)
const selectedZoneId = ref<number | null>(null)
const selectedChannelIds = ref<number[]>([])

// Filters
const selectedMetricCodes = ref<string[]>([])
const metricList = ref<MetricDefinition[]>([])
const timeRangePreset = ref('1h')
const customTimeRange = ref<[Date, Date] | null>(null)
const selectedBatchId = ref<number | null>(null)
const batches = ref<any[]>([])
const qualityFilter = ref('')

// Query state
const querying = ref(false)
const queryError = ref('')
const hasQueried = ref(false)

// Results
const rawData = ref<TelemetryRecord[]>([])
const deviceMap = ref<Map<number, SensorDevice>>(new Map())
const channelMap = ref<Map<number, SensorChannel>>(new Map())

const channelOptions = computed(() => {
  return channels.value.map((ch) => {
    const dev = deviceMap.value.get(ch.sensor_device_id)
    const prefix = dev ? `${dev.name} / ` : ''
    return {
      value: ch.id,
      label: `${prefix}${ch.channel_code} - ${getMetricName(ch.metric_code)}`
    }
  })
})

const tableData = computed<TableRow[]>(() => {
  return rawData.value.map((r) => {
    const ch = channelMap.value.get(r.sensor_channel_id)
    const dev = ch ? deviceMap.value.get(ch.sensor_device_id) : undefined
    return {
      collected_at: r.collected_at,
      device_name: dev?.name || '-',
      channel_code: ch?.channel_code || '-',
      metric_code: r.metric_code,
      value: r.value,
      quality_flag: r.quality_flag
    }
  })
})

const tablePage = ref(1)
const tablePageSize = ref(20)
const pagedTableData = computed(() => {
  const start = (tablePage.value - 1) * tablePageSize.value
  return tableData.value.slice(start, start + tablePageSize.value)
})

// Chart series grouped by channel_id + metric_code (each channel gets its own line)
const chartSeries = computed(() => {
  const key = (chId: number, code: string) => `${chId}:${code}`
  const grouped: Record<string, Array<{ time: string; value: number }>> = {}
  const meta: Record<string, { chId: number; code: string }> = {}

  for (const r of rawData.value) {
    const k = key(r.sensor_channel_id, r.metric_code)
    if (!grouped[k]) {
      grouped[k] = []
      meta[k] = { chId: r.sensor_channel_id, code: r.metric_code }
    }
    grouped[k].push({ time: r.collected_at, value: r.value })
  }

  return Object.entries(grouped)
    .filter(([k]) => {
      const m = meta[k]
      return m && selectedMetricCodes.value.includes(m.code)
    })
    .map(([k, data]) => {
      const m = meta[k]
      const ch = channelMap.value.get(m.chId)
      const dev = ch ? deviceMap.value.get(ch.sensor_device_id) : undefined
      const label = dev
        ? `${dev.name} / ${ch!.channel_code} - ${getMetricName(m.code)}`
        : `CH#${m.chId} - ${getMetricName(m.code)}`
      return { name: label, data }
    })
})

// Stats
const statsByMetric = computed(() => {
  const grouped: Record<string, number[]> = {}
  for (const r of rawData.value) {
    if (!grouped[r.metric_code]) grouped[r.metric_code] = []
    grouped[r.metric_code].push(r.value)
  }
  const result = new Map<string, { avg: number; max: number; min: number }>()
  for (const [code, values] of Object.entries(grouped)) {
    if (!selectedMetricCodes.value.includes(code)) continue
    const sum = values.reduce((a, b) => a + b, 0)
    result.set(code, {
      avg: sum / values.length,
      max: Math.max(...values),
      min: Math.min(...values)
    })
  }
  return result
})

// Event points for chart (placeholder - populated if batch events API available)
const eventPoints = ref<any[]>([])
const timelineEvents = ref<any[]>([])

function getTimeRange(): { start: string; end: string } {
  const end = new Date()
  let start: Date
  switch (timeRangePreset.value) {
    case '1h': start = new Date(end.getTime() - 60 * 60 * 1000); break
    case '3h': start = new Date(end.getTime() - 3 * 60 * 60 * 1000); break
    case '6h': start = new Date(end.getTime() - 6 * 60 * 60 * 1000); break
    case '12h': start = new Date(end.getTime() - 12 * 60 * 60 * 1000); break
    case '24h': start = new Date(end.getTime() - 24 * 60 * 60 * 1000); break
    case '3d': start = new Date(end.getTime() - 3 * 24 * 60 * 60 * 1000); break
    case '7d': start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000); break
    case 'custom':
      if (customTimeRange.value) {
        return {
          start: customTimeRange.value[0].toISOString(),
          end: customTimeRange.value[1].toISOString()
        }
      }
      start = new Date(end.getTime() - 60 * 60 * 1000)
      break
    default: start = new Date(end.getTime() - 60 * 60 * 1000)
  }
  return { start: start.toISOString(), end: end.toISOString() }
}

async function loadGreenhouses() {
  try {
    const result = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = result.items
  } catch { /* ignore */ }
}

async function loadMetrics() {
  try {
    const result = await metricApi.getMetrics({ page_size: EXTRA_LARGE_PAGE_SIZE })
    metricList.value = result.items
    populateMetricNames(result.items)
  } catch { /* ignore */ }
}

async function loadBatches() {
  try {
    const result = await cropApi.getBatches({ page_size: LARGE_PAGE_SIZE })
    batches.value = result.items || []
  } catch { /* ignore */ }
}

async function onGreenhouseChange() {
  selectedZoneId.value = null
  selectedChannelIds.value = []
  devices.value = []
  channels.value = []

  if (!selectedGreenhouseId.value) return

  try {
    const result = await greenhouseApi.getGreenhouseZones(selectedGreenhouseId.value)
    zones.value = result.items
  } catch { /* ignore */ }

  try {
    const result = await deviceApi.getSensorDevices({
      greenhouse_id: selectedGreenhouseId.value,
      page_size: LARGE_PAGE_SIZE
    })
    devices.value = result.items
    deviceMap.value = new Map(result.items.map((d) => [d.id, d]))
  } catch { /* ignore */ }

  // Load all channels for all devices
  await loadAllChannels()
}

async function onZoneChange() {
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
    deviceMap.value = new Map(result.items.map((d) => [d.id, d]))
  } catch { /* ignore */ }

  await loadAllChannels()
}

async function loadAllChannels() {
  channels.value = []
  const results = await Promise.all(
    devices.value.map((dev) =>
      deviceApi.getSensorChannels({
        sensor_device_id: dev.id,
        page_size: LARGE_PAGE_SIZE
      }).catch(() => ({ items: [] as SensorChannel[] }))
    )
  )
  const allCh = results.flatMap((r) => r.items)
  channels.value = allCh
  channelMap.value = new Map(allCh.map((c) => [c.id, c]))
}

function onPresetChange() {
  if (timeRangePreset.value !== 'custom') {
    customTimeRange.value = null
  }
}

async function doQuery() {
  if (selectedChannelIds.value.length === 0 || selectedMetricCodes.value.length === 0) return

  querying.value = true
  queryError.value = ''
  rawData.value = []
  eventPoints.value = []
  timelineEvents.value = []
  tablePage.value = 1

  const { start, end } = getTimeRange()

  try {
    const result = await telemetryApi.queryTelemetry({
      sensor_channel_id: selectedChannelIds.value.join(','),
      metric_code: selectedMetricCodes.value.join(','),
      start_time: start,
      end_time: end,
      batch_id: selectedBatchId.value || undefined,
      quality_flag: qualityFilter.value || undefined,
      page: 1,
      page_size: EXTRA_LARGE_PAGE_SIZE
    })

    // Sort by time ascending for correct chart rendering
    rawData.value = result.items.sort((a, b) =>
      new Date(a.collected_at).getTime() - new Date(b.collected_at).getTime()
    )
    hasQueried.value = true
  } catch (e: any) {
    queryError.value = e?.message || '查询失败'
  } finally {
    querying.value = false
  }
}

function reset() {
  selectedChannelIds.value = []
  selectedMetricCodes.value = []
  timeRangePreset.value = '1h'
  customTimeRange.value = null
  selectedBatchId.value = null
  qualityFilter.value = ''
  rawData.value = []
  hasQueried.value = false
  queryError.value = ''
}

onMounted(() => {
  loadGreenhouses()
  loadMetrics()
  loadBatches()
})

// When a batch is selected, load its devices and auto-select channels
watch(selectedBatchId, async (batchId) => {
  if (!batchId) return
  try {
    const { items } = await cropApi.getBatchDevices(batchId, 'sensor')
    if (items && items.length > 0) {
      // Find channels for batch sensor devices (parallel)
      const results = await Promise.all(
        items.map((bd: any) =>
          deviceApi.getSensorChannels({
            sensor_device_id: bd.device_id,
            page_size: LARGE_PAGE_SIZE
          }).catch(() => ({ items: [] as SensorChannel[] }))
        )
      )
      const allCh = results.flatMap((r) => r.items)
      channels.value = allCh
      channelMap.value = new Map(allCh.map((c) => [c.id, c]))
      selectedChannelIds.value = allCh.map((c) => c.id)
    }
  } catch { /* ignore */ }
})
</script>

<style scoped lang="scss">
.trends-page {
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

  .filter-section {
    margin: 12px 0;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .filter-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
  }

  .stats-row {
    margin-top: 16px;
  }

  .stat-card {
    .stat-title {
      font-size: 13px;
      color: var(--color-text-secondary);
      margin-bottom: 8px;
    }
    .stat-values {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }
    .stat-item {
      font-size: 13px;
      strong {
        color: var(--color-primary);
      }
    }
  }

  .chart-area {
    margin-top: 16px;
  }

  .table-card {
    margin-top: 16px;
  }

  .table-pagination {
    margin-top: 12px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
