<template>
  <AppLayout>
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
            <div class="stat-value">{{ overview.alerts_open }}</div>
            <div class="stat-label">活跃告警</div>
          </div>
        </div>
        <div class="stat-card data">
          <div class="stat-icon">
            <el-icon size="32"><DataLine /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ formatBigNumber(overview.today_data_points ?? 0) }}</div>
            <div class="stat-label">今日数据</div>
          </div>
        </div>
      </div>

      <!-- 告警列表 -->
      <div class="section-card">
        <div class="section-header">
          <h2 class="section-title">告警列表（最新5条）</h2>
          <el-button type="primary" link @click="goAlerts">查看全部告警</el-button>
        </div>
        <el-table :data="recentAlerts" v-loading="loading" stripe>
          <el-table-column label="级别" width="100">
            <template #default="{ row }">
              <el-tag :type="getAlertLevelType(row.level)">
                {{ getAlertLevelName(row.level) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              {{ getAlertTypeName(row.type) }}
            </template>
          </el-table-column>
          <el-table-column prop="message" label="消息" min-width="200" />
          <el-table-column prop="device_name" label="设备" width="120" />
          <el-table-column prop="triggered_at" label="时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.triggered_at || row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loading && recentAlerts.length === 0" class="empty-alert">
          暂无告警
        </div>
      </div>

      <!-- 设备分布图表 -->
      <div class="charts-grid">
        <div class="chart-card">
          <h2 class="chart-title">设备类型分布</h2>
          <div ref="typeChartRef" class="chart-container" role="img" aria-label="设备类型分布饼图：显示传感器和执行器的比例"></div>
        </div>
        <div class="chart-card">
          <h2 class="chart-title">设备分组分布</h2>
          <div ref="groupChartRef" class="chart-container" role="img" aria-label="设备分组分布饼图：显示各分组设备数量"></div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { Monitor, Warning, Bell, DataLine } from '@element-plus/icons-vue'
import { AppLayout } from '@/components/layout'
import { dashboardApi, alertApi, deviceApi, deviceGroupApi } from '@/api'
import { formatDateTime, getAlertLevelType, getAlertLevelName, getAlertTypeName } from '@/utils/format'
import type { DashboardOverview } from '@/types'

const router = useRouter()

// 数据
const loading = ref(false)
const overview = ref<DashboardOverview>({
  devices_online: 0,
  devices_offline: 0,
  alerts_open: 0
})
const recentAlerts = ref<any[]>([])
const typeDistribution = ref<any[]>([])
const groupDistribution = ref<any[]>([])

// 图表 - 使用 shallowRef 避免 ECharts 实例的深度响应式代理
const typeChartRef = ref<HTMLElement>()
const groupChartRef = ref<HTMLElement>()
const typeChart = shallowRef<echarts.ECharts | null>(null)
const groupChart = shallowRef<echarts.ECharts | null>(null)

// 当前日期
const currentDate = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日`
})

// 格式化大数字（带万单位）
function formatBigNumber(num: number): string {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toLocaleString()
}

// 跳转告警页
function goAlerts() {
  router.push('/alerts')
}

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    // 并行请求多个接口
    const [dashboardData, alertsData, devicesData, groupsData] = await Promise.all([
      dashboardApi.getDashboardData(),
      alertApi.getAlerts({ page: 1, page_size: 5 }).catch(() => ({ items: [], page: 1, page_size: 5, total: 0 })),
      deviceApi.getDevices({ page: 1, page_size: 1000 }).catch(() => ({ items: [], page: 1, page_size: 1000, total: 0 })),
      deviceGroupApi.getDeviceGroups().catch(() => ({ items: [] }))
    ])

    // 概览数据
    overview.value = {
      devices_online: dashboardData.devices_online ?? 0,
      devices_offline: dashboardData.devices_offline ?? 0,
      alerts_open: dashboardData.alerts_open ?? 0,
      today_data_points: dashboardData.today_data_points ?? 0
    }

    // 最近告警 - 构建设备ID到名称的映射
    const deviceNameMap = new Map<number, string>()
    for (const device of devicesData.items || []) {
      deviceNameMap.set(device.id, device.name)
    }
    recentAlerts.value = (alertsData.items || []).map((alert: any) => ({
      ...alert,
      device_name: alert.device_name || deviceNameMap.get(alert.device_id) || `设备${alert.device_id}`
    }))

    // 设备类型分布统计
    const typeMap = new Map<string, number>()
    for (const device of devicesData.items || []) {
      const type = device.type || 'UNKNOWN'
      typeMap.set(type, (typeMap.get(type) || 0) + 1)
    }
    typeDistribution.value = Array.from(typeMap.entries()).map(([type, count]) => ({ type, count }))

    // 设备分组分布统计
    const groupMap = new Map<string, number>()
    const groupNames = new Map<number, string>()
    for (const group of groupsData.items || []) {
      groupNames.set(group.id, group.name)
    }
    for (const device of devicesData.items || []) {
      if (device.group_id) {
        const groupName = groupNames.get(device.group_id) || `分组${device.group_id}`
        groupMap.set(groupName, (groupMap.get(groupName) || 0) + 1)
      }
    }
    groupDistribution.value = Array.from(groupMap.entries()).map(([group_name, count]) => ({ group_name, count }))

    // 更新图表
    updateCharts()
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
  if (groupChartRef.value) {
    groupChart.value = echarts.init(groupChartRef.value)
  }
}

// 更新图表
function updateCharts() {
  // 设备类型分布饼图
  if (typeChart.value && typeDistribution.value.length > 0) {
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
          label: {
            show: false
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 14,
              fontWeight: 'bold'
            }
          },
          data: typeDistribution.value.map((item) => ({
            name: item.type === 'SENSOR' ? '传感器' : '执行器',
            value: item.count
          }))
        }
      ]
    })
  }

  // 设备分组分布饼图
  if (groupChart.value && groupDistribution.value.length > 0) {
    groupChart.value.setOption({
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
          label: {
            show: false
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 14,
              fontWeight: 'bold'
            }
          },
          data: groupDistribution.value.map((item) => ({
            name: item.group_name,
            value: item.count
          }))
        }
      ]
    })
  }
}

// 窗口大小变化时重绘图表
function handleResize() {
  typeChart.value?.resize()
  groupChart.value?.resize()
}

onMounted(() => {
  fetchData()
  initCharts()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  typeChart.value?.dispose()
  groupChart.value?.dispose()
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
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
  }

  .current-date {
    color: #909399;
    font-size: 14px;
  }

  // 指标卡片
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 20px;

    @media (max-width: 1200px) {
      grid-template-columns: repeat(2, 1fr);
    }

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }

  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .stat-icon {
      width: 64px;
      height: 64px;
      border-radius: 12px;
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
    }

    .stat-label {
      font-size: 14px;
      color: #909399;
      margin-top: 4px;
    }

    &.online .stat-icon {
      background: linear-gradient(135deg, #67c23a, #85ce61);
    }

    &.offline .stat-icon {
      background: linear-gradient(135deg, #f56c6c, #fab6b6);
    }

    &.alert .stat-icon {
      background: linear-gradient(135deg, #e6a23c, #f3d19e);
    }

    &.data .stat-icon {
      background: linear-gradient(135deg, #409eff, #a0cfff);
    }
  }

  // 区块卡片
  .section-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    margin-bottom: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
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
    color: #909399;
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
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
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
