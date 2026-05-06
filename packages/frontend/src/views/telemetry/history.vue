<template>
  <div class="history-page">
    <div class="page-header">
      <h1 class="page-title">历史趋势</h1>
    </div>

    <div class="filter-section">
      <el-form :inline="true">
        <el-form-item label="传感器设备">
          <el-select v-model="query.device_id" placeholder="选择设备" filterable style="width: 220px" @change="onDeviceChange">
            <el-option v-for="device in sensorDevices" :key="device.id" :label="`${device.name} (${device.device_code})`" :value="device.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="通道">
          <el-select v-model="query.channel_id" placeholder="选择通道" filterable style="width: 220px">
            <el-option v-for="ch in deviceChannels" :key="ch.id" :label="`${ch.channel_code} - ${getMetricName(ch.metric_code)}`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="指标">
          <el-select v-model="query.metric_codes" multiple collapse-tags clearable placeholder="多指标对比" style="width: 280px">
            <el-option v-for="m in metricOptions" :key="m.value" :label="m.label" :value="m.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="质量标识">
          <el-select v-model="query.quality_flag" clearable style="width: 150px">
            <el-option label="normal" value="normal" />
            <el-option label="outlier" value="outlier" />
            <el-option label="missing" value="missing" />
            <el-option label="interpolated" value="interpolated" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 360px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="fetchData">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-row :gutter="16" class="summary-row" v-if="statsMapKeys.length > 0">
      <el-col v-for="key in statsMapKeys" :key="key" :span="8">
        <el-card>
          <div class="summary-title">{{ getMetricName(key) }}</div>
          <div>avg: {{ formatNumber(statsMap[key]?.avg) }}</div>
          <div>max: {{ formatNumber(statsMap[key]?.max) }}</div>
          <div>min: {{ formatNumber(statsMap[key]?.min) }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="chart-card">
      <metric-trend-chart :series="chartSeries" />
    </el-card>

    <el-card class="table-card">
      <el-table :data="pagedTableData" stripe max-height="420" v-loading="loading">
        <el-table-column prop="collected_at" label="采集时间" width="180">
          <template #default="{ row }">{{ formatDate(row.collected_at) }}</template>
        </el-table-column>
        <el-table-column prop="metric_code" label="指标" width="120">
          <template #default="{ row }">{{ getMetricName(row.metric_code) }}</template>
        </el-table-column>
        <el-table-column prop="value" label="数值" width="120">
          <template #default="{ row }">{{ formatNumber(row.value) }}</template>
        </el-table-column>
        <el-table-column prop="quality_flag" label="质量标识" width="120" />
        <el-table-column prop="sensor_channel_id" label="通道ID" width="90" />
      </el-table>
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="tableData.length"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { deviceApi, telemetryApi } from '@/api'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { formatDate, formatNumber, getMetricName } from '@/utils/format'
import type { SensorDevice, SensorChannel, TelemetryRecord } from '@/types'

const sensorDevices = ref<SensorDevice[]>([])
const deviceChannels = ref<SensorChannel[]>([])
const loading = ref(false)
const timeRange = ref<[string, string] | null>(null)

const query = reactive({
  device_id: null as number | null,
  channel_id: null as number | null,
  metric_codes: ['TEMP'] as string[],
  quality_flag: undefined as string | undefined
})

const tableData = ref<TelemetryRecord[]>([])
const telemetryByMetric = ref<Record<string, TelemetryRecord[]>>({})
const statsMap = ref<Record<string, { avg: number; max: number; min: number }>>({})

const pagination = reactive({ page: 1, pageSize: 50 })

const metricOptions = [
  { value: 'TEMP', label: '温度' },
  { value: 'HUMIDITY', label: '湿度' },
  { value: 'PH', label: 'pH值' },
  { value: 'EC', label: '电导率' },
  { value: 'CO2', label: 'CO2' },
  { value: 'LIGHT', label: '光照' }
]

const chartSeries = computed(() =>
  query.metric_codes.map((code) => ({
    name: getMetricName(code),
    data: (telemetryByMetric.value[code] || []).map((i) => ({ time: i.collected_at, value: i.value }))
  }))
)

const statsMapKeys = computed(() => Object.keys(statsMap.value))

const pagedTableData = computed(() => {
  const start = (pagination.page - 1) * pagination.pageSize
  return tableData.value.slice(start, start + pagination.pageSize)
})

function computeStats(records: TelemetryRecord[]): { avg: number; max: number; min: number } {
  if (records.length === 0) return { avg: 0, max: 0, min: 0 }
  let sum = 0
  let max = -Infinity
  let min = Infinity
  for (const r of records) {
    sum += r.value
    if (r.value > max) max = r.value
    if (r.value < min) min = r.value
  }
  return { avg: sum / records.length, max, min }
}

async function fetchData() {
  if (!query.channel_id || !timeRange.value) return
  loading.value = true
  try {
    const [startTime, endTime] = timeRange.value

    const promises = query.metric_codes.map((metricCode) =>
      telemetryApi.getChannelHistory(query.channel_id!, {
        metric_code: metricCode,
        start_time: startTime,
        end_time: endTime,
        quality_flag: query.quality_flag,
        page: 1,
        page_size: 2000
      })
    )

    const results = await Promise.all(promises)

    const nextTelemetryMap: Record<string, TelemetryRecord[]> = {}
    const merged: TelemetryRecord[] = []
    results.forEach((res, idx) => {
      const metricCode = query.metric_codes[idx]
      nextTelemetryMap[metricCode] = res.items
      merged.push(...res.items)
    })
    telemetryByMetric.value = nextTelemetryMap
    tableData.value = merged.sort((a, b) => new Date(b.collected_at).getTime() - new Date(a.collected_at).getTime())

    // Compute stats client-side
    const nextStatsMap: Record<string, { avg: number; max: number; min: number }> = {}
    for (const [code, records] of Object.entries(nextTelemetryMap)) {
      nextStatsMap[code] = computeStats(records)
    }
    statsMap.value = nextStatsMap
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.device_id = null
  query.channel_id = null
  query.metric_codes = ['TEMP']
  query.quality_flag = undefined
  tableData.value = []
  telemetryByMetric.value = {}
  statsMap.value = {}
  deviceChannels.value = []
}

async function loadSensorDevices() {
  try {
    const data = await deviceApi.getSensorDevices({ page_size: LARGE_PAGE_SIZE })
    sensorDevices.value = data.items
  } catch { /* ignore */ }
}

async function onDeviceChange() {
  query.channel_id = null
  tableData.value = []
  telemetryByMetric.value = {}
  statsMap.value = {}

  if (!query.device_id) {
    deviceChannels.value = []
    return
  }

  try {
    const data = await deviceApi.getSensorChannels({
      sensor_device_id: query.device_id,
      page_size: LARGE_PAGE_SIZE
    })
    deviceChannels.value = data.items
  } catch { /* ignore */ }
}

onMounted(() => {
  loadSensorDevices()
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  timeRange.value = [start.toISOString(), end.toISOString()]
})
</script>

<style scoped lang="scss">
.history-page {
  .page-header {
    margin-bottom: 20px;
  }
  .page-title {
    font-size: 22px;
    font-weight: 700;
    margin: 0;
  }
  .filter-section {
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }
  .summary-row {
    margin-top: 12px;
  }
  .summary-title {
    font-weight: 700;
    margin-bottom: 8px;
  }
  .chart-card {
    margin-top: 12px;
  }
  .table-card {
    margin-top: 12px;
  }
  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
}
</style>
