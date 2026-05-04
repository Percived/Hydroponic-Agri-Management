<template>
    <div class="history-page">
      <div class="page-header">
        <h1 class="page-title">历史数据查询</h1>
      </div>

      <!-- 查询条件 -->
      <div class="filter-section">
        <el-form :inline="true" :model="queryParams">
          <el-form-item label="设备">
            <el-select v-model="queryParams.device_id" placeholder="选择设备" filterable style="width: 250px">
              <el-option-group label="传感器">
                <el-option v-for="device in sensorDevices" :key="device.id" :label="`${device.name} (${device.device_code})`" :value="device.id" />
              </el-option-group>
            </el-select>
          </el-form-item>
          <el-form-item label="指标">
            <el-select v-model="queryParams.metric_code" placeholder="选择指标" style="width: 150px">
              <el-option v-for="metric in metricOptions" :key="metric.value" :label="metric.label" :value="metric.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <div class="time-range-wrapper">
              <div class="time-presets">
                <el-button
                  v-for="preset in timePresets"
                  :key="preset.label"
                  :type="activePreset === preset.label ? 'primary' : ''"
                  size="small"
                  @click="applyPreset(preset)"
                >{{ preset.label }}</el-button>
              </div>
              <el-date-picker
                v-model="timeRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
                value-format="YYYY-MM-DDTHH:mm:ss"
                style="width: 360px"
              />
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchData" :loading="loading">查询</el-button>
            <el-button @click="resetQuery">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 数据展示 -->
      <template v-if="hasData">
        <!-- 统计摘要 -->
        <el-card class="stats-card">
          <div class="stats-grid">
            <div class="stat-item">
              <div class="stat-label">平均值</div>
              <div class="stat-value">{{ formatNumber(stats?.avg) }} {{ currentUnit }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">最大值</div>
              <div class="stat-value">{{ formatNumber(stats?.max) }} {{ currentUnit }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">最小值</div>
              <div class="stat-value">{{ formatNumber(stats?.min) }} {{ currentUnit }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">数据点数</div>
              <div class="stat-value">{{ chartData.length }}</div>
            </div>
          </div>
        </el-card>

        <!-- 展示方式切换 -->
        <div class="view-toggle">
          <el-radio-group v-model="viewMode">
            <el-radio-button value="chart">图表</el-radio-button>
            <el-radio-button value="table">表格</el-radio-button>
          </el-radio-group>
        </div>

        <!-- 图表 -->
        <el-card v-show="viewMode === 'chart'" class="chart-card">
          <div ref="chartRef" class="chart-container"></div>
        </el-card>

        <!-- 表格 -->
        <el-card v-show="viewMode === 'table'" class="table-card">
          <el-table :data="tableData" stripe max-height="500">
            <el-table-column prop="collected_at" label="采集时间" width="180">
              <template #default="{ row }">
                {{ formatDate(row.collected_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="value" label="数值" width="150">
              <template #default="{ row }">
                {{ formatNumber(row.value) }} {{ currentUnit }}
              </template>
            </el-table-column>
            <el-table-column prop="raw_value" label="原始值" width="150">
              <template #default="{ row }">
                {{ formatNumber(row.raw_value) }} {{ currentUnit }}
              </template>
            </el-table-column>
            <el-table-column prop="quality" label="质量" width="100">
              <template #default="{ row }">
                <el-tag :type="row.quality === 0 ? 'success' : 'danger'" size="small">
                  {{ row.quality === 0 ? '正常' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :total="pagination.total"
              :page-sizes="[20, 50, 100, 200]"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchTableData"
              @current-change="fetchTableData"
            />
          </div>
        </el-card>
      </template>

      <el-empty v-else :description="emptyDescription" />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, shallowRef, nextTick } from 'vue'
import { getDevices } from '@/api/device'
import { getHistoryTelemetry, getTelemetryStats } from '@/api/telemetry'
import { formatDate, formatNumber } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { Device, TelemetryPoint, TelemetryStats, MetricNames, MetricUnits, DeviceType } from '@/types'
import * as echarts from 'echarts'

// ---- 工具函数 ----

/** 将本地 Date 转为 RFC3339 格式（带时区偏移） */
function toLocalRFC3339(d: Date): string {
  const pad = (n: number) => n.toString().padStart(2, '0')
  const tzOff = -d.getTimezoneOffset()
  const tzSign = tzOff >= 0 ? '+' : '-'
  const tzH = pad(Math.floor(Math.abs(tzOff) / 60))
  const tzM = pad(Math.abs(tzOff) % 60)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}${tzSign}${tzH}:${tzM}`
}

// ---- 状态 ----

const devices = ref<Device[]>([])
const loading = ref(false)

const queryParams = reactive({
  device_id: null as number | null,
  metric_code: ''
})

const timeRange = ref<[string, string] | null>(null)
const activePreset = ref<string>('24小时')

// 图表数据（全量，不分页）
const chartData = ref<TelemetryPoint[]>([])
// 表格数据（分页）
const tableData = ref<TelemetryPoint[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 100,
  total: 0
})

const stats = ref<TelemetryStats | null>(null)
const hasData = ref(false)
const emptyDescription = ref('请选择设备和指标后点击查询')
const viewMode = ref<'chart' | 'table'>('chart')

const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)
let resizeObserver: ResizeObserver | null = null

// ---- 计算属性 ----

const sensorDevices = computed(() => devices.value.filter((d) => d.type === 'SENSOR'))

const metricOptions = [
  { label: '温度', value: 'TEMP' },
  { label: '湿度', value: 'HUMIDITY' },
  { label: 'pH值', value: 'PH' },
  { label: '电导率', value: 'EC' },
  { label: 'CO2', value: 'CO2' },
  { label: '光照', value: 'LIGHT' }
]

const currentUnit = computed(() => MetricUnits[queryParams.metric_code] || '')

const timePresets = [
  { label: '1小时', hours: 1 },
  { label: '6小时', hours: 6 },
  { label: '24小时', hours: 24 },
  { label: '7天',   hours: 24 * 7 }
]

// ---- 时间预设 ----

function applyPreset(preset: { label: string; hours: number }) {
  activePreset.value = preset.label
  const end = new Date()
  const start = new Date(end.getTime() - preset.hours * 3600 * 1000)
  timeRange.value = [toLocalRFC3339(start), toLocalRFC3339(end)]
}

// ---- 数据获取 ----

async function fetchDevices() {
  try {
    const result = await getDevices({ type: DeviceType.SENSOR, page_size: LARGE_PAGE_SIZE })
    devices.value = result.items
  } catch {
    // 错误在拦截器中统一处理
  }
}

function getQueryTime(): [string, string] | null {
  if (!timeRange.value) return null
  return [timeRange.value[0], timeRange.value[1]]
}

async function fetchData() {
  if (!queryParams.device_id || !queryParams.metric_code || !timeRange.value) {
    return
  }

  const [startTime, endTime] = getQueryTime()!
  loading.value = true
  try {
    // P0-1: chart 全量数据（page_size=2000）
    // P2-8: 三个请求并行
    const [chartResult, statsResult, tableResult] = await Promise.all([
      getHistoryTelemetry({
        device_id: queryParams.device_id,
        metric_code: queryParams.metric_code,
        start_time: startTime,
        end_time: endTime,
        page: 1,
        page_size: 2000
      }),
      getTelemetryStats({
        device_id: queryParams.device_id,
        metric_code: queryParams.metric_code,
        start_time: startTime,
        end_time: endTime
      }),
      getHistoryTelemetry({
        device_id: queryParams.device_id,
        metric_code: queryParams.metric_code,
        start_time: startTime,
        end_time: endTime,
        page: pagination.page,
        page_size: pagination.pageSize
      })
    ])

    chartData.value = chartResult.items
    stats.value = statsResult
    tableData.value = tableResult.items
    pagination.total = tableResult.total
    hasData.value = true

    await nextTick()
    drawChart()
  } catch {
    // 错误在拦截器中统一处理
  } finally {
    loading.value = false
  }
}

async function fetchTableData() {
  const time = getQueryTime()
  if (!queryParams.device_id || !queryParams.metric_code || !time) return

  try {
    const result = await getHistoryTelemetry({
      device_id: queryParams.device_id,
      metric_code: queryParams.metric_code,
      start_time: time[0],
      end_time: time[1],
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = result.items
    pagination.total = result.total
  } catch {
    // 已处理
  }
}

function resetQuery() {
  queryParams.device_id = null
  queryParams.metric_code = ''
  timeRange.value = null
  activePreset.value = ''
  pagination.page = 1
  hasData.value = false
  chartData.value = []
  tableData.value = []
  stats.value = null
  emptyDescription.value = '请选择设备和指标后点击查询'
}

// ---- 图表 ----

function drawChart() {
  if (!chartRef.value || chartData.value.length === 0) return

  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartRef.value)
  }

  // P0-1: 图标用全量 chartData（升序排列）
  const data = [...chartData.value].sort(
    (a, b) => new Date(a.collected_at).getTime() - new Date(b.collected_at).getTime()
  )
  const anomalyPoints = data.filter((d) => d.quality !== 0)

  const timeLabels = data.map((d) => formatDate(d.collected_at, 'MM-DD HH:mm:ss'))

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params
        if (!p) return ''
        const idx = p.dataIndex
        const point = data[idx]
        const lines = [
          `${p.name}`,
          `${p.seriesName}: ${formatNumber(point?.value)} ${currentUnit.value}`
        ]
        if (point && point.quality !== 0) {
          lines.push(`⚠ 异常数据 (quality=${point.quality})`)
        }
        return lines.join('<br/>')
      }
    },
    legend: {
      data: [MetricNames[queryParams.metric_code] || queryParams.metric_code, '异常值'],
      selected: { '异常值': true }
    },
    grid: {
      left: 60,
      right: 40,
      top: 40,
      bottom: data.length > 200 ? 80 : 60
    },
    dataZoom: [
      { type: 'inside', start: 0, end: 100 },
      { start: 0, end: 100, height: 20, bottom: data.length > 200 ? 40 : 10 }
    ],
    xAxis: {
      type: 'category',
      data: timeLabels,
      axisLabel: {
        rotate: 45,
        interval: Math.max(0, Math.floor(data.length / 15) - 1)  // auto interval
      }
    },
    yAxis: {
      type: 'value',
      name: currentUnit.value
    },
    series: [
      {
        name: MetricNames[queryParams.metric_code] || queryParams.metric_code,
        type: 'line',
        data: data.map((d) => d.value),
        smooth: data.length < 500,
        symbol: 'none',
        lineStyle: { width: 2 },
        itemStyle: { color: '#0ea882' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(14, 168, 130, 0.25)' },
            { offset: 1, color: 'rgba(14, 168, 130, 0.02)' }
          ])
        }
      },
      // P1-4: 异常点 scatter 叠加层
      {
        name: '异常值',
        type: 'scatter',
        data: anomalyPoints.map((d) => {
          const idx = data.indexOf(d)
          return [timeLabels[idx], d.value]
        }),
        symbolSize: 10,
        symbol: 'circle',
        itemStyle: {
          color: '#f56c6c',
          borderColor: '#fff',
          borderWidth: 1
        },
        emphasis: {
          scale: 1.5
        }
      }
    ]
  }

  chartInstance.value.setOption(option, { notMerge: true })
}

// ---- 生命周期 ----

watch(viewMode, async () => {
  if (viewMode.value === 'chart') {
    await nextTick()
    drawChart()
    chartInstance.value?.resize()
  }
})

onMounted(() => {
  fetchDevices()
  // P0-2: 默认最近 24 小时（本地时间）
  applyPreset({ label: '24小时', hours: 24 })
  // P2-7: ResizeObserver 监听容器变化
  if (chartRef.value) {
    resizeObserver = new ResizeObserver(() => {
      chartInstance.value?.resize()
    })
    resizeObserver.observe(chartRef.value)
  }
})

onUnmounted(() => {
  chartInstance.value?.dispose()
  resizeObserver?.disconnect()
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
    color: var(--color-text-primary);
    margin: 0;
    text-wrap: balance;
  }

  .filter-section {
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }

  .time-range-wrapper {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .time-presets {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .stats-card {
    margin-bottom: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
  }

  .stats-grid {
    display: flex;
    gap: 40px;
    flex-wrap: wrap;
  }

  .stat-item {
    .stat-label {
      font-size: 14px;
      color: var(--color-text-secondary);
      margin-bottom: 4px;
    }

    .stat-value {
      font-size: 24px;
      font-weight: 600;
      color: var(--color-text-primary);
    }
  }

  .view-toggle {
    margin-bottom: 16px;
  }

  .chart-card {
    .chart-container {
      height: 400px;
    }
  }

  .table-card {
    .pagination-container {
      display: flex;
      justify-content: flex-end;
      margin-top: var(--spacing-md);
      padding-top: var(--spacing-md);
      border-top: 1px solid var(--border-color);
    }
  }
}
</style>
