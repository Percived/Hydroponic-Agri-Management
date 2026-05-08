<template>
    <div class="dashboard-page">
      <div class="page-header">
        <h1 class="page-title">系统概览</h1>
        <span class="current-date">{{ currentDate }}</span>
      </div>

      <!-- 关键指标 -->
      <div class="stats-grid">
        <div class="stat-card online">
          <div class="stat-icon">
            <el-icon size="32"><Monitor /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview.devices_online }}</div>
            <div class="stat-label">在线设备</div>
          </div>
        </div>
        <div class="stat-card offline">
          <div class="stat-icon">
            <el-icon size="32"><Warning /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview.devices_offline }}</div>
            <div class="stat-label">离线设备</div>
          </div>
        </div>
        <div class="stat-card alert">
          <div class="stat-icon">
            <el-icon size="32"><Bell /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">
              {{ overview.alerts_open }}
              <span v-if="overview.alerts_critical" class="stat-sub">({{ overview.alerts_critical }} 严重)</span>
            </div>
            <div class="stat-label">活跃告警</div>
          </div>
        </div>
        <div class="stat-card total">
          <div class="stat-icon">
            <el-icon size="32"><DataLine /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview.devices_total ?? (overview.devices_online + overview.devices_offline) }}</div>
            <div class="stat-label">设备总数</div>
          </div>
        </div>
        <div class="stat-card today-alerts">
          <div class="stat-icon">
            <el-icon size="32"><Clock /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ overview.alerts_today ?? 0 }}</div>
            <div class="stat-label">今日告警</div>
          </div>
        </div>
      </div>

      <!-- 温室概览 -->
      <div v-if="overview.greenhouse_summary?.length" class="section-card">
        <div class="section-header">
          <h2 class="section-title">温室概览</h2>
        </div>
        <div class="greenhouse-grid">
          <div
            v-for="gh in overview.greenhouse_summary"
            :key="gh.greenhouse_id"
            class="greenhouse-card"
          >
            <div class="gh-name">{{ gh.name }}</div>
            <div class="gh-stats">
              <div class="gh-stat">
                <span class="gh-stat-label">设备数</span>
                <span class="gh-stat-value">{{ gh.sensor_count + gh.actuator_count }}</span>
              </div>
              <div class="gh-stat">
                <span class="gh-stat-label">平均温度</span>
                <span class="gh-stat-value">{{ gh.avg_temp != null ? gh.avg_temp.toFixed(1) + '°C' : '--' }}</span>
              </div>
              <div class="gh-stat">
                <span class="gh-stat-label">平均湿度</span>
                <span class="gh-stat-value">{{ gh.avg_humidity != null ? gh.avg_humidity.toFixed(1) + '%' : '--' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 设备分布 + 告警列表 -->
      <div class="charts-grid">
        <div class="chart-card">
          <h2 class="chart-title">设备类型分布</h2>
          <div ref="typeChartRef" class="chart-container" role="img" aria-label="设备类型分布饼图：显示传感器和执行器的比例"></div>
        </div>
        <div class="section-card" style="margin-bottom: 0">
          <div class="section-header">
            <h2 class="section-title">最近命令</h2>
            <el-button type="primary" link @click="goCommands">查看全部</el-button>
          </div>
          <el-table v-if="overview.recent_commands?.length" :data="overview.recent_commands" stripe size="small">
            <el-table-column prop="command_type" label="类型" width="100" />
            <el-table-column prop="device_name" label="设备" width="140" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="commandStatusType(row.status)" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" min-width="160">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
          <div v-else class="empty-alert">暂无命令记录</div>
        </div>
      </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { Monitor, Warning, Bell, DataLine, Clock } from '@element-plus/icons-vue'
import { dashboardApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { DashboardOverview } from '@/types'

const router = useRouter()

// 数据
const loading = ref(false)
const overview = ref<DashboardOverview>({
  sensors_online: 0,
  sensors_offline: 0,
  sensors_total: 0,
  actuators_online: 0,
  actuators_offline: 0,
  actuators_total: 0,
  devices_online: 0,
  devices_offline: 0,
  devices_total: 0,
  alerts_open: 0,
  alerts_critical: 0,
  alerts_today: 0,
  device_type_distribution: [],
  greenhouse_summary: [],
  recent_commands: []
})

// 图表
const typeChartRef = ref<HTMLElement>()
const typeChart = shallowRef<echarts.ECharts | null>(null)

// 当前日期
const currentDate = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日`
})

// 命令状态标签类型
function commandStatusType(status: string): string {
  switch (status) {
    case 'EXECUTED': return 'success'
    case 'SENT': return 'primary'
    case 'PENDING': return 'info'
    case 'FAILED': return 'danger'
    default: return 'info'
  }
}

// 跳转
function goCommands() {
  router.push('/controls/commands')
}

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await dashboardApi.getDashboardData()
    overview.value = {
      sensors_online: data.sensors_online ?? 0,
      sensors_offline: data.sensors_offline ?? 0,
      sensors_total: data.sensors_total ?? 0,
      actuators_online: data.actuators_online ?? 0,
      actuators_offline: data.actuators_offline ?? 0,
      actuators_total: data.actuators_total ?? 0,
      devices_online: data.devices_online ?? 0,
      devices_offline: data.devices_offline ?? 0,
      devices_total: data.devices_total ?? 0,
      alerts_open: data.alerts_open ?? 0,
      alerts_critical: data.alerts_critical ?? 0,
      alerts_today: data.alerts_today ?? 0,
      device_type_distribution: data.device_type_distribution ?? [],
      greenhouse_summary: data.greenhouse_summary ?? [],
      recent_commands: data.recent_commands ?? []
    }
    updateTypeChart()
  } catch (error) {
    console.error('[Dashboard] Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

// 初始化图表
function initCharts() {
  if (typeChartRef.value) {
    typeChart.value = echarts.init(typeChartRef.value)
  }
}

// 设备类型分布饼图
function updateTypeChart() {
  if (!typeChart.value) return
  const dist = overview.value.device_type_distribution
  if (dist.length === 0) return

  const typeNames: Record<string, string> = {
    SENSOR: '传感器',
    ACTUATOR: '执行器'
  }

  typeChart.value.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: 0,
      left: 'center'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' }
        },
        data: dist.map((item) => ({
          name: typeNames[item.type] || item.type,
          value: item.count
        }))
      }
    ]
  })
}

// 窗口大小变化时重绘图表
function handleResize() {
  typeChart.value?.resize()
}

onMounted(() => {
  fetchData()
  initCharts()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  typeChart.value?.dispose()
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
    text-wrap: balance;
  }

  .current-date {
    color: var(--color-text-secondary);
    font-size: 14px;
  }

  // 指标卡片
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 16px;
    margin-bottom: 20px;

    @media (max-width: 1400px) {
      grid-template-columns: repeat(3, 1fr);
    }

    @media (max-width: 992px) {
      grid-template-columns: repeat(2, 1fr);
    }

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }

  .stat-card {
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border-radius: var(--radius-lg);
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    box-shadow: var(--shadow-card);
    transition: transform var(--transition-fast), box-shadow var(--transition-fast);

    &:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-card-hover);
    }

    .stat-icon {
      width: 64px;
      height: 64px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    .stat-info {
      flex: 1;
    }

    .stat-value {
      font-size: 28px;
      font-weight: 600;
      line-height: 1.2;

      .stat-sub {
        font-size: 13px;
        font-weight: normal;
        color: var(--color-danger);
      }
    }

    .stat-label {
      font-size: 14px;
      color: var(--color-text-secondary);
      margin-top: 4px;
    }

    &.online .stat-icon {
      background: linear-gradient(135deg, var(--color-primary), var(--color-primary-lighter));
    }

    &.offline .stat-icon {
      background: linear-gradient(135deg, #f56c6c, #fab6b6);
    }

    &.alert .stat-icon {
      background: linear-gradient(135deg, #e6a23c, #f3d19e);
    }

    &.total .stat-icon {
      background: linear-gradient(135deg, #409eff, #a0cfff);
    }

    &.today-alerts .stat-icon {
      background: linear-gradient(135deg, #8b5cf6, #c4b5fd);
    }
  }

  // 温室概览
  .greenhouse-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 16px;
  }

  .greenhouse-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 16px;
    border: 1px solid var(--border-color);
    transition: border-color var(--transition-fast), box-shadow var(--transition-fast);

    &:hover {
      border-color: var(--color-primary-light);
      box-shadow: var(--shadow-card);
    }

    .gh-name {
      font-size: 15px;
      font-weight: 600;
      color: var(--color-text-primary);
      margin-bottom: 12px;
    }

    .gh-stats {
      display: flex;
      gap: 16px;
    }

    .gh-stat {
      text-align: center;

      .gh-stat-label {
        display: block;
        font-size: 12px;
        color: var(--color-text-secondary);
        margin-bottom: 4px;
      }

      .gh-stat-value {
        font-size: 16px;
        font-weight: 600;
        color: var(--color-text-primary);
      }
    }
  }

  // 区块卡片
  .section-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 20px;
    margin-bottom: 20px;
    box-shadow: var(--shadow-card);
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

  .empty-alert {
    text-align: center;
    color: var(--color-text-secondary);
    padding: 40px 0;
  }

  // 图表区域
  .charts-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;

    @media (max-width: 992px) {
      grid-template-columns: 1fr;
    }
  }

  .chart-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 20px;
    box-shadow: var(--shadow-card);
    transition: box-shadow var(--transition-normal);

    &:hover {
      box-shadow: var(--shadow-card-hover);
    }
  }

  .chart-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 16px 0;
  }

  .chart-container {
    height: 280px;
  }
}
</style>
