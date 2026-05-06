<template>
  <div class="batch-trends-page">
    <div class="page-header">
      <h1 class="page-title">批次趋势</h1>
    </div>

    <div class="filter-section">
      <el-form :inline="true">
        <el-form-item label="传感器通道">
          <el-select v-model="query.sensor_channel_id" placeholder="选择通道" filterable style="width: 220px">
            <el-option v-for="ch in sensorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.metric_code})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="批次ID">
          <el-input-number v-model="query.batch_id" :min="1" />
        </el-form-item>
        <el-form-item label="指标">
          <el-select v-model="query.metric_codes" multiple collapse-tags style="width: 240px">
            <el-option v-for="m in metricOptions" :key="m.value" :label="m.label" :value="m.value" />
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
        </el-form-item>
      </el-form>
    </div>

    <el-row :gutter="16">
      <el-col :span="16">
        <el-card>
          <template #header>环境曲线 + 告警事件 + 控制动作</template>
          <metric-trend-chart :series="chartSeries" :events="eventPoints" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>事件时间线</template>
          <batch-event-overlay :events="timelineEvents" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { alertApi, commandApi, deviceApi, telemetryApi } from '@/api'
import BatchEventOverlay from '@/components/charts/BatchEventOverlay.vue'
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'
import { getMetricName } from '@/utils/format'
import type { SensorChannel, TelemetryRecord } from '@/types'

const loading = ref(false)
const sensorChannels = ref<SensorChannel[]>([])
const query = reactive({
  sensor_channel_id: null as number | null,
  batch_id: undefined as number | undefined,
  metric_codes: ['TEMP', 'HUMIDITY'] as string[]
})
const timeRange = ref<[string, string] | null>(null)
const telemetryByMetric = ref<Record<string, TelemetryRecord[]>>({})
const eventPoints = ref<Array<{ time: string; value: number; label: string; eventType: 'alert' | 'control' }>>([])
const timelineEvents = ref<Array<{ type: 'alert' | 'control'; time: string; label: string }>>([])

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

async function fetchData() {
  if (!query.sensor_channel_id || query.metric_codes.length === 0 || !timeRange.value) return
  loading.value = true
  try {
    const [startTime, endTime] = timeRange.value
    const [telemetryRes, alertsRes, commandsRes] = await Promise.all([
      telemetryApi.queryTelemetry({
        sensor_channel_id: query.sensor_channel_id,
        start_time: startTime,
        end_time: endTime,
        batch_id: query.batch_id,
        page: 1,
        page_size: 2000
      }),
      alertApi.getAlerts({ page: 1, page_size: 200 }),
      commandApi.getCommands({ page: 1, page_size: 200 })
    ])

    // Group telemetry by metric_code
    const map: Record<string, TelemetryRecord[]> = {}
    for (const item of telemetryRes.items) {
      if (!map[item.metric_code]) map[item.metric_code] = []
      map[item.metric_code].push(item)
    }
    telemetryByMetric.value = map

    const allValues = Object.values(map).flat().map((i) => i.value)
    const baseline = allValues.length ? Math.max(...allValues) : 0
    const alerts = alertsRes.items.filter((a) => isInRange(a.triggered_at, startTime, endTime))
    const controls = commandsRes.items.filter((c) => isInRange(c.created_at, startTime, endTime))

    eventPoints.value = [
      ...alerts.map((a) => ({ time: a.triggered_at, value: baseline, label: `告警:${a.message}`, eventType: 'alert' as const })),
      ...controls.map((c) => ({ time: c.created_at, value: baseline, label: `控制:${c.command_type}`, eventType: 'control' as const }))
    ]
    timelineEvents.value = [
      ...alerts.map((a) => ({ type: 'alert' as const, time: a.triggered_at, label: `${a.level} ${a.message}` })),
      ...controls.map((c) => ({ type: 'control' as const, time: c.created_at, label: `${c.command_type} ${c.status}` }))
    ].sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime())
  } finally {
    loading.value = false
  }
}

function isInRange(time: string, start: string, end: string) {
  const t = new Date(time).getTime()
  return t >= new Date(start).getTime() && t <= new Date(end).getTime()
}

onMounted(async () => {
  const res = await deviceApi.getSensorChannels({ page_size: 200 })
  sensorChannels.value = res.items
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  timeRange.value = [start.toISOString(), end.toISOString()]
})
</script>

<style scoped lang="scss">
.batch-trends-page {
  .page-header {
    margin-bottom: 16px;
  }
  .page-title {
    margin: 0;
    font-size: 22px;
    font-weight: 700;
  }
  .filter-section {
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }
}
</style>
