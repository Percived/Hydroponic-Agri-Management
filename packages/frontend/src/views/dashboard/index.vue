<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h1 class="page-title">系统概览 (Command Center)</h1>
      <span class="current-date">{{ currentDate }}</span>
    </div>

    <!-- 关键指标 (Quick Stats) -->
    <div class="stats-grid">
      <div class="stat-card active-batches">
        <div class="stat-icon">
          <el-icon size="32"><Monitor /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ overview.stats?.active_batches_count ?? 0 }}</div>
          <div class="stat-label">活跃批次</div>
        </div>
      </div>
      <div class="stat-card alert">
        <div class="stat-icon">
          <el-icon size="32"><Bell /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ overview.stats?.unresolved_alerts ?? 0 }}</div>
          <div class="stat-label">活跃告警</div>
        </div>
      </div>
      <div class="stat-card online">
        <div class="stat-icon">
          <el-icon size="32"><DataLine /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">
            {{ overview.stats?.devices_online ?? 0 }}
            <span class="stat-sub">/ {{ (overview.stats?.devices_online ?? 0) + (overview.stats?.devices_offline ?? 0) }}</span>
          </div>
          <div class="stat-label">在线设备</div>
        </div>
      </div>
      <div class="stat-card energy">
        <div class="stat-icon">
          <el-icon size="32"><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ overview.stats?.energy_kwh_today ?? 0 }} <span class="stat-sub">kWh</span></div>
          <div class="stat-label">今日能耗</div>
        </div>
      </div>
    </div>

    <div class="main-content-grid">
      <!-- 左侧：温室环境与水培核心参数 -->
      <div class="left-column">
        <div class="section-header">
          <h2 class="section-title">温室实时监控</h2>
        </div>
        <div class="greenhouse-list">
          <div v-for="gh in overview.greenhouses" :key="gh.id" class="greenhouse-card">
            <div class="gh-header">
              <span class="gh-name">{{ gh.name }}</span>
              <el-tag :type="gh.health_score === 'good' ? 'success' : 'warning'" size="small">
                {{ gh.health_score === 'good' ? '健康 🟢' : '异常 🔴' }}
              </el-tag>
            </div>
            <div class="gh-metrics-grid">
              <div class="metric-item">
                <div class="metric-label">温度</div>
                <div class="metric-value">{{ gh.metrics.temperature.toFixed(1) }} °C</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">湿度</div>
                <div class="metric-value">{{ gh.metrics.humidity.toFixed(1) }} %</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">EC值</div>
                <div class="metric-value highlight">{{ gh.metrics.ec.toFixed(2) }}</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">pH值</div>
                <div class="metric-value highlight">{{ gh.metrics.ph.toFixed(2) }}</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">溶氧量</div>
                <div class="metric-value">{{ gh.metrics.do.toFixed(1) }} mg/L</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">光照度</div>
                <div class="metric-value">{{ gh.metrics.lux.toFixed(0) }} Lux</div>
              </div>
            </div>
            <div class="gh-footer">
              <span class="strategy-label">运行策略：</span>
              <span class="strategy-val">{{ gh.active_strategies?.join(', ') || '暂无' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：业务趋势与策略调度 -->
      <div class="right-column">
        <!-- 24小时趋势图 -->
        <div class="chart-card">
          <h2 class="chart-title">24小时水质趋势 (EC / pH)</h2>
          <div ref="trendChartRef" class="chart-container" role="img"></div>
        </div>

        <!-- 当前活跃批次 -->
        <div class="section-card">
          <div class="section-header">
            <h2 class="section-title">活跃批次进度</h2>
          </div>
          <el-table :data="overview.active_batches" stripe size="small" style="width: 100%">
            <el-table-column prop="batch_id" label="批次" width="80" />
            <el-table-column prop="crop_name" label="作物" />
            <el-table-column prop="stage" label="阶段" width="100" />
            <el-table-column label="已运行" width="80">
              <template #default="{ row }">{{ row.day }} 天</template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </div>

    <!-- 底部：快捷操作与设备交互 -->
    <div class="bottom-grid">
      <!-- 最近未处理告警 -->
      <div class="section-card">
        <div class="section-header">
          <h2 class="section-title">最近未处理告警</h2>
          <el-button type="primary" link @click="router.push('/alerts/overview')">查看全部</el-button>
        </div>
        <el-table :data="overview.recent_alerts" stripe size="small" style="width: 100%">
          <el-table-column prop="timestamp" label="时间" width="160">
            <template #default="{ row }">{{ formatDateTime(row.timestamp) }}</template>
          </el-table-column>
          <el-table-column prop="greenhouse_name" label="温室" width="120" />
          <el-table-column prop="severity" label="级别" width="100">
            <template #default="{ row }">
              <el-tag :type="row.severity === 'CRITICAL' ? 'danger' : 'warning'" size="small">
                {{ row.severity }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="内容" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="handleAck(row.alert_id)">确认</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 最近下发指令 -->
      <div class="section-card">
        <div class="section-header">
          <h2 class="section-title">最近下发指令</h2>
          <el-button type="primary" link @click="router.push('/controls/commands')">查看全部</el-button>
        </div>
        <el-table :data="overview.recent_commands" stripe size="small" style="width: 100%">
          <el-table-column prop="created_at" label="时间" width="160">
            <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
          </el-table-column>
          <el-table-column prop="device_name" label="设备" width="140" />
          <el-table-column prop="command_type" label="类型" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="commandStatusType(row.status)" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { Monitor, Bell, DataLine, Clock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { dashboardApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { DashboardData } from '@/types'

const router = useRouter()
const loading = ref(false)
const overview = ref<DashboardData>({
  stats: {
    active_batches_count: 0,
    unresolved_alerts: 0,
    devices_online: 0,
    devices_offline: 0,
    energy_kwh_today: 0,
    water_l_today: 0
  },
  greenhouses: [],
  trends: { timestamps: [], ec_avg: [], ph_avg: [] },
  active_batches: [],
  recent_alerts: [],
  recent_commands: []
})

const trendChartRef = ref<HTMLElement>()
const trendChart = shallowRef<echarts.ECharts | null>(null)

const currentDate = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日`
})

function commandStatusType(status: string): string {
  switch (status) {
    case 'EXECUTED': return 'success'
    case 'SENT': return 'primary'
    case 'PENDING': return 'info'
    case 'FAILED': return 'danger'
    default: return 'info'
  }
}

function handleAck(alertId: string) {
  ElMessage.success(`已确认告警: ${alertId}`)
  // TODO: Call API to acknowledge alert
}

async function fetchData() {
  loading.value = true
  try {
    const data = await dashboardApi.getDashboardData()
    overview.value = data
    updateTrendChart()
  } catch (error) {
    console.error('[Dashboard] Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

function initCharts() {
  if (trendChartRef.value) {
    trendChart.value = echarts.init(trendChartRef.value)
  }
}

function updateTrendChart() {
  if (!trendChart.value) return
  const trends = overview.value.trends
  if (!trends || !trends.timestamps || trends.timestamps.length === 0) return

  trendChart.value.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['EC值', 'pH值'], bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: trends.timestamps
    },
    yAxis: [
      { type: 'value', name: 'EC', position: 'left' },
      { type: 'value', name: 'pH', position: 'right', min: 0, max: 14 }
    ],
    series: [
      {
        name: 'EC值',
        type: 'line',
        smooth: true,
        data: trends.ec_avg,
        itemStyle: { color: '#409EFF' }
      },
      {
        name: 'pH值',
        type: 'line',
        smooth: true,
        yAxisIndex: 1,
        data: trends.ph_avg,
        itemStyle: { color: '#67C23A' }
      }
    ]
  })
}

function handleResize() {
  trendChart.value?.resize()
}

// SSE Integration
let deviceEventSource: EventSource | null = null
let commandEventSource: EventSource | null = null

function setupSSE() {
  // Using direct EventSource for now, could be replaced with robust utility later
  const token = localStorage.getItem('hydroponic_token') || ''
  
  // Connect to devices/subscribe for telemetry updates
  deviceEventSource = new EventSource(`/api/devices/subscribe?token=${token}`)
  deviceEventSource.addEventListener('telemetry_update', () => {
    try {
      // For simplicity in this demo, we'll just refetch if there's significant change
      // or implement direct reactivity:
    } catch (err) {
      console.error('SSE parsing error', err)
    }
  })

  // Listen to new alerts
  deviceEventSource.addEventListener('new_alert', () => {
    // A new alert arrives, we might want to reload the dashboard or push it to recent_alerts
    fetchData()
  })

  // Connect to commands/subscribe
  commandEventSource = new EventSource(`/api/commands/subscribe?token=${token}`)
  commandEventSource.addEventListener('command_dispatched', () => fetchData())
  commandEventSource.addEventListener('command_acked', () => fetchData())
}

onMounted(() => {
  fetchData()
  initCharts()
  setupSSE()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  trendChart.value?.dispose()
  if (deviceEventSource) deviceEventSource.close()
  if (commandEventSource) commandEventSource.close()
})
</script>

<style scoped lang="scss">
.dashboard-page {
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

  .current-date {
    color: var(--color-text-secondary);
    font-size: 14px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 20px;
  }

  .stat-card {
    background: var(--bg-card);
    border-radius: var(--radius-lg);
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    box-shadow: var(--shadow-card);
    transition: transform var(--transition-fast);

    &:hover { transform: translateY(-2px); }

    .stat-icon {
      width: 56px;
      height: 56px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    .stat-info { flex: 1; }

    .stat-value {
      font-size: 24px;
      font-weight: 600;
      .stat-sub { font-size: 14px; font-weight: normal; color: var(--color-text-secondary); }
    }

    .stat-label {
      font-size: 13px;
      color: var(--color-text-secondary);
      margin-top: 4px;
    }

    &.active-batches .stat-icon { background: linear-gradient(135deg, #67C23A, #95D475); }
    &.alert .stat-icon { background: linear-gradient(135deg, #F56C6C, #FAB6B6); }
    &.online .stat-icon { background: linear-gradient(135deg, #409EFF, #A0CFFF); }
    &.energy .stat-icon { background: linear-gradient(135deg, #E6A23C, #F3D19E); }
  }

  .main-content-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
    margin-bottom: 20px;

    @media (max-width: 1200px) {
      grid-template-columns: 1fr;
    }
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0;
  }

  .greenhouse-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .greenhouse-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 16px;
    box-shadow: var(--shadow-card);
    border: 1px solid var(--border-color);

    .gh-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 12px;
      .gh-name { font-weight: 600; font-size: 15px; }
    }

    .gh-metrics-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
      margin-bottom: 12px;
      background: var(--bg-body);
      padding: 12px;
      border-radius: 8px;

      .metric-item {
        .metric-label { font-size: 12px; color: var(--color-text-secondary); }
        .metric-value { 
          font-size: 14px; font-weight: 500; 
          &.highlight { color: var(--color-primary); font-weight: 600; }
        }
      }
    }

    .gh-footer {
      font-size: 13px;
      .strategy-label { color: var(--color-text-secondary); }
      .strategy-val { color: var(--color-success); font-weight: 500; }
    }
  }

  .chart-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 20px;
    box-shadow: var(--shadow-card);
    margin-bottom: 20px;

    .chart-title { font-size: 16px; font-weight: 600; margin: 0 0 16px 0; }
    .chart-container { height: 260px; }
  }

  .section-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 20px;
    box-shadow: var(--shadow-card);
  }

  .bottom-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;

    @media (max-width: 1200px) {
      grid-template-columns: 1fr;
    }
  }
}
</style>
