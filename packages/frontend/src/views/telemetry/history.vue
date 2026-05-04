<template>
  <AppLayout>
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
            <el-date-picker
              v-model="timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DDTHH:mm:ss"
              style="width: 360px"
            />
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
              @size-change="fetchData"
              @current-change="fetchData"
            />
          </div>
        </el-card>
      </template>

      <el-empty v-else description="请设置查询条件后点击查询" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, shallowRef } from 'vue'
import { AppLayout } from '@/components/layout'
import { getDevices } from '@/api/device'
import { getHistoryTelemetry, getTelemetryStats } from '@/api/telemetry'
import { formatDate, formatNumber } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { Device, TelemetryPoint, TelemetryStats, MetricNames, MetricUnits, DeviceType } from '@/types'
import * as echarts from 'echarts'

const devices = ref<Device[]>([])
const loading = ref(false)

// 查询参数
const queryParams = reactive({
  device_id: null as number | null,
  metric_code: ''
})

const timeRange = ref<[string, string] | null>(null)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 100,
  total: 0
})

// 数据
const tableData = ref<TelemetryPoint[]>([])
const stats = ref<TelemetryStats | null>(null)
const hasData = ref(false)

// 展示模式
const viewMode = ref<'chart' | 'table'>('chart')

// 图表 - 使用 shallowRef 避免 ECharts 实例的深度响应式代理
const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)

// 传感器设备
const sensorDevices = computed(() => devices.value.filter((d) => d.type === 'SENSOR'))

// 指标选项
const metricOptions = [
  { label: '温度', value: 'TEMP' },
  { label: '湿度', value: 'HUMIDITY' },
  { label: 'pH值', value: 'PH' },
  { label: '电导率', value: 'EC' },
  { label: 'CO2', value: 'CO2' },
  { label: '光照', value: 'LIGHT' }
]

// 当前单位
const currentUnit = computed(() => MetricUnits[queryParams.metric_code] || '')

// 获取设备列表
async function fetchDevices() {
  try {
    const result = await getDevices({ type: DeviceType.SENSOR, page_size: LARGE_PAGE_SIZE })
    devices.value = result.items
  } catch {
    // 错误已处理
  }
}

// 获取数据
async function fetchData() {
  if (!queryParams.device_id || !queryParams.metric_code || !timeRange.value) {
    return
  }

  loading.value = true
  try {
    // 获取历史数据
    const historyResult = await getHistoryTelemetry({
      device_id: queryParams.device_id,
      metric_code: queryParams.metric_code,
      start_time: timeRange.value[0],
      end_time: timeRange.value[1],
      page: pagination.page,
      page_size: pagination.pageSize
    })
    tableData.value = historyResult.items
    pagination.total = historyResult.total
    hasData.value = true

    // 获取统计数据
    const statsResult = await getTelemetryStats({
      device_id: queryParams.device_id,
      metric_code: queryParams.metric_code,
      start_time: timeRange.value[0],
      end_time: timeRange.value[1]
    })
    stats.value = statsResult

    // 绘制图表
    drawChart()
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 重置查询
function resetQuery() {
  queryParams.device_id = null
  queryParams.metric_code = ''
  timeRange.value = null
  pagination.page = 1
  hasData.value = false
  tableData.value = []
  stats.value = null
}

// 绘制图表
function drawChart() {
  if (!chartRef.value || tableData.value.length === 0) return

  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartRef.value)
  }

  const data = [...tableData.value].sort((a, b) => new Date(a.collected_at).getTime() - new Date(b.collected_at).getTime())

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis'
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        start: 0,
        end: 100
      }
    ],
    xAxis: {
      type: 'category',
      data: data.map((d) => formatDate(d.collected_at, 'MM-DD HH:mm:ss')),
      axisLabel: {
        rotate: 45
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
        smooth: true,
        itemStyle: {
          color: '#409eff'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        }
      }
    ]
  }

  chartInstance.value.setOption(option)
}

// 窗口大小变化时重绘图表
function handleResize() {
  chartInstance.value?.resize()
}

// 监听展示模式变化
watch(viewMode, () => {
  if (viewMode.value === 'chart') {
    setTimeout(() => {
      drawChart()
      chartInstance.value?.resize()
    }, 100)
  }
})

onMounted(() => {
  fetchDevices()
  // 设置默认时间范围（最近24小时）
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  timeRange.value = [
    start.toISOString().slice(0, 19),
    end.toISOString().slice(0, 19)
  ]
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  chartInstance.value?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.history-page {
  .page-header {
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
  }

  .filter-section {
    margin-bottom: 16px;
    padding: 16px;
    background: #fff;
    border-radius: 4px;
  }

  .stats-card {
    margin-bottom: 16px;
  }

  .stats-grid {
    display: flex;
    gap: 40px;
  }

  .stat-item {
    .stat-label {
      font-size: 14px;
      color: #909399;
      margin-bottom: 4px;
    }

    .stat-value {
      font-size: 24px;
      font-weight: 600;
      color: #303133;
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
      margin-top: 16px;
    }
  }
}
</style>