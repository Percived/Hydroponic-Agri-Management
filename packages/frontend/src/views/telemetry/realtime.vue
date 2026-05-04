<template>
  <AppLayout>
    <div class="realtime-page">
      <div class="page-header">
        <h1 class="page-title">实时数据监控</h1>
        <span v-if="sseConnected" class="sse-status connected">● 实时连接中</span>
        <span v-else class="sse-status disconnected">○ 连接断开</span>
      </div>

      <!-- 设备选择 -->
      <div class="filter-section">
        <el-select v-model="selectedDeviceId" placeholder="选择设备" filterable style="width: 300px" @change="onDeviceChange">
          <el-option-group label="传感器">
            <el-option v-for="device in sensorDevices" :key="device.id" :label="`${device.name} (${device.device_code})`" :value="device.id" />
          </el-option-group>
        </el-select>
        <el-button type="primary" @click="fetchInitialData" :loading="loading">刷新数据</el-button>
      </div>

      <!-- 数据卡片 -->
      <div v-if="selectedDeviceId" class="data-section">
        <div v-if="loading && telemetryData.length === 0" class="loading-placeholder">
          <el-skeleton :rows="5" animated />
        </div>
        <template v-else>
          <div v-if="telemetryData.length === 0" class="empty-placeholder">
            <el-empty description="暂无遥测数据" />
          </div>
          <template v-else>
            <!-- 指标卡片 -->
            <div class="metrics-grid" aria-live="polite" aria-label="实时遥测指标">
              <el-card v-for="item in telemetryData" :key="item.metric_code" class="metric-card" :class="{ updated: updatedMetrics.has(item.metric_code) }">
                <div class="metric-header">
                  <span class="metric-name">{{ MetricNames[item.metric_code] || item.metric_code }}</span>
                  <el-tag :type="item.quality === 0 ? 'success' : 'danger'" size="small">
                    {{ item.quality === 0 ? '正常' : '异常' }}
                  </el-tag>
                </div>
                <div class="metric-value">
                  {{ formatNumber(item.value) }}
                  <span class="metric-unit">{{ MetricUnits[item.metric_code] || '' }}</span>
                </div>
                <div class="metric-time">
                  {{ formatDate(item.collected_at) }}
                </div>
              </el-card>
            </div>

            <!-- 趋势图 -->
            <el-card class="chart-card">
              <template #header>
                <div class="chart-header">
                  <span>趋势图 (最近1小时)</span>
                  <el-select v-model="selectedMetric" placeholder="选择指标" style="width: 150px">
                    <el-option v-for="item in telemetryData" :key="item.metric_code" :label="MetricNames[item.metric_code] || item.metric_code" :value="item.metric_code" />
                  </el-select>
                </div>
              </template>
              <div ref="chartRef" class="chart-container" role="img" aria-label="遥测数据趋势图"></div>
            </el-card>
          </template>
        </template>
      </div>

      <el-empty v-else description="请选择设备查看实时数据" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, shallowRef } from 'vue'
import { AppLayout } from '@/components/layout'
import { getDevices } from '@/api/device'
import { getLatestTelemetry, getHistoryTelemetry } from '@/api/telemetry'
import { useTelemetrySSE } from '@/composables'
import { formatDate, formatNumber } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { Device, TelemetryPoint, MetricNames, MetricUnits, DeviceType } from '@/types'
import * as echarts from 'echarts'

const devices = ref<Device[]>([])
const selectedDeviceId = ref<number | null>(null)
const telemetryData = ref<TelemetryPoint[]>([])
const loading = ref(false)
const updatedMetrics = ref(new Set<string>())

// SSE
const deviceCodes = ref<string[]>([])
const { connected: sseConnected, latestUpdate, connect: connectSSE, disconnect: disconnectSSE } = useTelemetrySSE({ deviceCodes })

// 图表
const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)
const selectedMetric = ref('')
const historyData = ref<TelemetryPoint[]>([])

// 传感器设备
const sensorDevices = computed(() => devices.value.filter((d) => d.type === 'SENSOR'))

// SSE 数据到达时更新遥测数据
watch(latestUpdate, (update) => {
  if (!update || !selectedDeviceId.value) return

  const device = devices.value.find((d) => d.device_code === update.device_code)
  if (!device || device.id !== selectedDeviceId.value) return

  for (const m of update.metrics) {
    const idx = telemetryData.value.findIndex((t) => t.metric_code === m.code)
    const point: TelemetryPoint = {
      device_id: selectedDeviceId.value,
      metric_code: m.code,
      value: m.value,
      raw_value: m.value,
      quality: 0,
      collected_at: update.collected_at
    }
    if (idx >= 0) {
      telemetryData.value[idx] = point
    } else {
      telemetryData.value.push(point)
    }
    // Flash animation
    updatedMetrics.value = new Set(updatedMetrics.value).add(m.code)
    setTimeout(() => {
      const s = new Set(updatedMetrics.value)
      s.delete(m.code)
      updatedMetrics.value = s
    }, 1500)
  }

  // If chart metric matches, refetch history
  if (selectedMetric.value && update.metrics.some((m) => m.code === selectedMetric.value)) {
    fetchHistoryData()
  }
})

// 获取设备列表
async function fetchDevices() {
  try {
    const result = await getDevices({ type: DeviceType.SENSOR, page_size: LARGE_PAGE_SIZE })
    devices.value = result.items
  } catch {
    // 错误已处理
  }
}

// 获取最新遥测（初始加载）
async function fetchInitialData() {
  if (!selectedDeviceId.value) return
  loading.value = true
  try {
    const result = await getLatestTelemetry({ device_id: selectedDeviceId.value })
    telemetryData.value = result.items
    if (result.items.length > 0 && !selectedMetric.value) {
      selectedMetric.value = result.items[0].metric_code
    }
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 获取历史数据并绘制图表
async function fetchHistoryData() {
  if (!selectedDeviceId.value || !selectedMetric.value) return

  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - 60 * 60 * 1000)

  try {
    const result = await getHistoryTelemetry({
      device_id: selectedDeviceId.value,
      metric_code: selectedMetric.value,
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      page_size: LARGE_PAGE_SIZE
    })
    historyData.value = result.items
    drawChart()
  } catch {
    // 错误已处理
  }
}

// 绘制图表
function drawChart() {
  if (!chartRef.value || !selectedMetric.value) return

  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartRef.value)
  }

  const data = historyData.value.sort((a, b) => new Date(a.collected_at).getTime() - new Date(b.collected_at).getTime())

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: data.map((d) => formatDate(d.collected_at, 'HH:mm:ss')),
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      name: MetricUnits[selectedMetric.value] || ''
    },
    series: [
      {
        name: MetricNames[selectedMetric.value] || selectedMetric.value,
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

// 设备变化
function onDeviceChange() {
  selectedMetric.value = ''
  const device = devices.value.find((d) => d.id === selectedDeviceId.value)
  if (device) {
    deviceCodes.value = [device.device_code]
    connectSSE()
  } else {
    disconnectSSE()
    deviceCodes.value = []
  }
  fetchInitialData()
}

// 监听指标选择
watch(selectedMetric, () => {
  fetchHistoryData()
})

// 窗口大小变化时重绘图表
function handleResize() {
  chartInstance.value?.resize()
}

onMounted(() => {
  fetchDevices()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  disconnectSSE()
  chartInstance.value?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.realtime-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
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
      background: #f5f7fa;
    }
  }

  .filter-section {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;
    padding: 16px;
    background: #fff;
    border-radius: 4px;
  }

  .data-section {
    .loading-placeholder,
    .empty-placeholder {
      padding: 40px;
      text-align: center;
    }
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
    margin-bottom: 16px;
  }

  .metric-card {
    transition: box-shadow 0.3s ease;

    &.updated {
      box-shadow: 0 0 0 2px #67c23a;
    }

    .metric-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }

    .metric-name {
      font-size: 14px;
      color: #606266;
    }

    .metric-value {
      font-size: 32px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 8px;
    }

    .metric-unit {
      font-size: 14px;
      font-weight: normal;
      color: #909399;
    }

    .metric-time {
      font-size: 12px;
      color: #909399;
    }
  }

  .chart-card {
    .chart-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .chart-container {
      height: 400px;
    }
  }
}
</style>
