<template>
  <AppLayout>
    <div class="device-detail-page">
      <div class="page-header">
        <el-button @click="goBack" :icon="ArrowLeft">返回列表</el-button>
        <h1 class="page-title">设备详情 - {{ device?.device_code || '加载中...' }}</h1>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="8" animated />
      </div>

      <!-- 错误状态 -->
      <div v-else-if="errorMsg" class="error-container">
        <el-result icon="error" :title="errorMsg" sub-title="请检查网络连接或返回列表重试">
          <template #extra>
            <el-button type="primary" @click="loadData">重新加载</el-button>
            <el-button @click="goBack">返回列表</el-button>
          </template>
        </el-result>
      </div>

      <!-- 设备数据 -->
      <template v-else-if="device">
        <el-tabs v-model="activeTab" type="border-card">
          <!-- Tab 1: 基本信息 -->
          <el-tab-pane label="基本信息" name="info">
            <el-card class="info-card">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="设备编码">{{ device.device_code }}</el-descriptions-item>
                <el-descriptions-item label="设备名称">{{ device.name }}</el-descriptions-item>
                <el-descriptions-item label="设备类型">
                  {{ device.type === 'SENSOR' ? '传感器' : '执行器' }}
                </el-descriptions-item>
                <el-descriptions-item label="设备分类">{{ getCategoryName(device.category) }}</el-descriptions-item>
                <el-descriptions-item label="通信协议">{{ device.protocol }}</el-descriptions-item>
                <el-descriptions-item label="采样间隔">{{ device.sampling_interval_sec || 60 }} 秒</el-descriptions-item>
                <el-descriptions-item label="所属分组">{{ getGroupName(device.group_id) }}</el-descriptions-item>
                <el-descriptions-item label="启用状态">
                  <el-tag :type="device.status === 'ENABLED' ? 'success' : 'danger'">
                    {{ device.status === 'ENABLED' ? '启用' : '禁用' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="在线状态">
                  <el-tag :type="health?.online ? 'success' : 'danger'">
                    {{ health?.online ? '在线' : '离线' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="最后上报">
                  {{ device.last_seen_at ? formatDate(device.last_seen_at) : '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ formatDate(device.created_at) }}</el-descriptions-item>
                <el-descriptions-item label="更新时间">{{ formatDate(device.updated_at) }}</el-descriptions-item>
              </el-descriptions>

              <div class="action-buttons" v-if="canEdit">
                <el-button type="primary" @click="openEditDialog">编辑设备</el-button>
                <el-button
                  :type="device.status === 'ENABLED' ? 'danger' : 'success'"
                  @click="toggleStatus"
                >
                  {{ device.status === 'ENABLED' ? '禁用设备' : '启用设备' }}
                </el-button>
              </div>
            </el-card>

            <!-- 最新遥测数据 -->
            <el-card class="telemetry-card" v-if="device.type === 'SENSOR'">
              <template #header><span>最新遥测数据</span></template>
              <div v-if="telemetryLoading" class="loading-placeholder">
                <el-skeleton :rows="3" animated />
              </div>
              <div v-else-if="telemetryData.length === 0" class="empty-placeholder">
                暂无遥测数据
              </div>
              <div v-else class="telemetry-grid">
                <div v-for="item in telemetryData" :key="item.metric_code" class="telemetry-item">
                  <div class="metric-name">{{ MetricNames[item.metric_code] || item.metric_code }}</div>
                  <div class="metric-value">
                    {{ formatNumber(item.value) }}
                    <span class="metric-unit">{{ MetricUnits[item.metric_code] || '' }}</span>
                  </div>
                  <div class="metric-meta">
                    <el-tag :type="item.quality === 0 ? 'success' : 'danger'" size="small">
                      {{ item.quality === 0 ? '正常' : '异常' }}
                    </el-tag>
                    <span class="metric-time">{{ formatDate(item.collected_at, 'HH:mm:ss') }}</span>
                  </div>
                </div>
              </div>
            </el-card>
          </el-tab-pane>

          <!-- Tab 2: 数据看板 -->
          <el-tab-pane label="数据看板" name="dashboard" v-if="device.type === 'SENSOR'">
            <div class="dashboard-tab">
              <!-- 时间范围选择 -->
              <div class="dashboard-controls">
                <el-date-picker
                  v-model="timeRange"
                  type="datetimerange"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
                  format="YYYY-MM-DD HH:mm"
                  value-format="YYYY-MM-DDTHH:mm:ss.SSSZ"
                  :shortcuts="timeShortcuts"
                  style="width: 400px"
                />
                <el-select v-model="selectedMetrics" placeholder="选择指标" multiple style="width: 300px">
                  <el-option
                    v-for="code in availableMetrics"
                    :key="code"
                    :label="MetricNames[code] || code"
                    :value="code"
                  />
                </el-select>
                <el-button type="primary" @click="fetchTelemetrySummary" :loading="summaryLoading">查询</el-button>
              </div>

              <!-- 在线率 -->
              <div class="online-rate" v-if="telemetrySummary">
                <span>在线率：</span>
                <el-progress
                  :percentage="Math.round((telemetrySummary.online_rate || 0) * 100)"
                  :color="telemetrySummary.online_rate > 0.8 ? '#67c23a' : telemetrySummary.online_rate > 0.5 ? '#e6a23c' : '#f56c6c'"
                  style="width: 200px"
                />
              </div>

              <!-- 指标概览卡片 -->
              <div class="metric-cards" v-if="telemetrySummary">
                <el-card v-for="code in selectedMetrics" :key="code" class="metric-card" shadow="hover">
                  <template #header>
                    <span>{{ MetricNames[code] || code }}</span>
                    <span class="metric-unit-text">{{ MetricUnits[code] || '' }}</span>
                  </template>
                  <div class="metric-stats">
                    <div class="stat-item">
                      <span class="stat-label">平均</span>
                      <span class="stat-value">{{ telemetrySummary.metrics[code]?.avg?.toFixed(2) ?? '-' }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">最大</span>
                      <span class="stat-value">{{ telemetrySummary.metrics[code]?.max?.toFixed(2) ?? '-' }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">最小</span>
                      <span class="stat-value">{{ telemetrySummary.metrics[code]?.min?.toFixed(2) ?? '-' }}</span>
                    </div>
                    <div class="stat-item">
                      <span class="stat-label">告警</span>
                      <span class="stat-value alert-count">{{ telemetrySummary.metrics[code]?.alerts ?? 0 }}</span>
                    </div>
                  </div>
                  <!-- 迷你图表 -->
                  <div :ref="(el) => setChartRef(code, el as HTMLElement)" class="mini-chart" style="height: 150px" />
                </el-card>
              </div>

              <!-- 历史告警 -->
              <el-card class="alert-events-card" v-if="telemetrySummary?.alert_events?.length">
                <template #header><span>历史告警事件</span></template>
                <el-table :data="telemetrySummary.alert_events" size="small">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="level" label="级别" width="80">
                    <template #default="{ row }">
                      <el-tag :type="row.level === 'CRITICAL' ? 'danger' : row.level === 'WARN' ? 'warning' : 'info'" size="small">
                        {{ row.level }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
                  <el-table-column prop="status" label="状态" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.status }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="triggered_at" label="触发时间" width="160">
                    <template #default="{ row }">{{ formatDate(row.triggered_at) }}</template>
                  </el-table-column>
                </el-table>
              </el-card>
            </div>
          </el-tab-pane>
        </el-tabs>
      </template>

      <!-- 编辑弹窗 -->
      <el-dialog v-model="editDialogVisible" title="编辑设备" width="500px">
        <el-form ref="formRef" :model="editForm" :rules="formRules" label-width="100px">
          <el-form-item label="设备名称" prop="name">
            <el-input v-model="editForm.name" placeholder="请输入设备名称" autocomplete="off" name="name" />
          </el-form-item>
          <el-form-item label="设备分类" prop="category">
            <el-select v-model="editForm.category" placeholder="请选择设备分类">
              <el-option v-for="cat in categoryOptions" :key="cat.value" :label="cat.label" :value="cat.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="所属分组" prop="group_id">
            <el-select v-model="editForm.group_id" placeholder="请选择分组" clearable>
              <el-option v-for="group in deviceGroups" :key="group.id" :label="group.name" :value="group.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="采样间隔" prop="sampling_interval_sec">
            <el-input-number v-model="editForm.sampling_interval_sec" :min="5" :max="3600" />
            <span class="ml-sm">秒</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleEdit">确定</el-button>
        </template>
      </el-dialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { AppLayout } from '@/components/layout'
import { useDeviceStore } from '@/stores/device'
import { usePermission } from '@/composables/usePermission'
import { getCategoryName, formatDate, formatNumber } from '@/utils/format'
import { getLatestTelemetry } from '@/api/telemetry'
import { getTelemetrySummary, TelemetrySummary } from '@/api/device'
import { Device, DeviceHealth, TelemetryPoint, DeviceGroup, MetricNames, MetricUnits } from '@/types'

const route = useRoute()
const router = useRouter()
const deviceStore = useDeviceStore()
const { canEditDevice } = usePermission()

const canEdit = computed(() => canEditDevice())
const device = computed(() => deviceStore.currentDevice as Device | null)
const deviceGroups = computed(() => deviceStore.deviceGroups as DeviceGroup[])

const loading = ref(false)
const errorMsg = ref('')
const health = ref<DeviceHealth | null>(null)
const telemetryData = ref<TelemetryPoint[]>([])
const telemetryLoading = ref(false)
const activeTab = ref('info')

const deviceId = computed(() => Number(route.params.id))

// Dashboard state
const timeRange = ref<[string, string]>()
const selectedMetrics = ref<string[]>([])
const availableMetrics = ref<string[]>([])
const summaryLoading = ref(false)
const telemetrySummary = ref<TelemetrySummary | null>(null)
const chartRefs: Record<string, HTMLElement | null> = {}
const chartInstances: Record<string, echarts.ECharts> = {}

const timeShortcuts = [
  { text: '最近1小时', value: () => {
    const end = new Date(); const start = new Date(); start.setHours(start.getHours() - 1)
    return [start, end] as [Date, Date]
  }},
  { text: '最近6小时', value: () => {
    const end = new Date(); const start = new Date(); start.setHours(start.getHours() - 6)
    return [start, end] as [Date, Date]
  }},
  { text: '最近24小时', value: () => {
    const end = new Date(); const start = new Date(); start.setHours(start.getHours() - 24)
    return [start, end] as [Date, Date]
  }},
  { text: '最近7天', value: () => {
    const end = new Date(); const start = new Date(); start.setDate(start.getDate() - 7)
    return [start, end] as [Date, Date]
  }}
]

function setChartRef(code: string, el: HTMLElement | null) {
  chartRefs[code] = el
}

async function fetchTelemetrySummary() {
  if (!deviceId.value) return
  summaryLoading.value = true
  try {
    const params: any = {}
    if (timeRange.value && timeRange.value.length === 2) {
      params.from = timeRange.value[0]
      params.to = timeRange.value[1]
    }
    const data = await getTelemetrySummary(deviceId.value, params.from, params.to)
    telemetrySummary.value = data
    availableMetrics.value = Object.keys(data.metrics)
    if (selectedMetrics.value.length === 0) {
      selectedMetrics.value = availableMetrics.value.slice(0, 4)
    }
    await nextTick()
    renderCharts()
  } catch (e) {
    console.error('[Device Detail] Failed to fetch telemetry summary:', e)
  } finally {
    summaryLoading.value = false
  }
}

function renderCharts() {
  if (!telemetrySummary.value) return
  for (const code of selectedMetrics.value) {
    const el = chartRefs[code]
    if (!el) continue
    if (chartInstances[code]) chartInstances[code].dispose()
    const metric = telemetrySummary.value.metrics[code]
    if (!metric?.hourly?.length) continue

    const chart = echarts.init(el)
    chartInstances[code] = chart

    const hours = metric.hourly.map((h: any) => h.hour)
    const values = metric.hourly.map((h: any) => h.avg)

    chart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 10, top: 10, bottom: 25 },
      xAxis: {
        type: 'category',
        data: hours,
        axisLabel: { fontSize: 10, rotate: 30 }
      },
      yAxis: {
        type: 'value',
        axisLabel: { fontSize: 10 }
      },
      series: [{
        data: values,
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#409eff', width: 2 },
        areaStyle: { color: 'rgba(64,158,255,0.1)' }
      }]
    })
  }
}

function disposeCharts() {
  for (const code of Object.keys(chartInstances)) {
    chartInstances[code]?.dispose()
    delete chartInstances[code]
  }
}

onBeforeUnmount(() => {
  disposeCharts()
})

// 编辑弹窗
const editDialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)

const editForm = reactive({
  name: '',
  category: '',
  group_id: null as number | null,
  sampling_interval_sec: 60
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入设备名称', trigger: 'blur' },
    { min: 1, max: 64, message: '设备名称长度为 1-64 个字符', trigger: 'blur' }
  ],
  category: [{ required: true, message: '请选择设备分类', trigger: 'change' }]
}

const categoryOptions = [
  { label: '温度', value: 'TEMP' },
  { label: '湿度', value: 'HUMIDITY' },
  { label: 'pH值', value: 'PH' },
  { label: '电导率', value: 'EC' },
  { label: 'CO2', value: 'CO2' },
  { label: '光照', value: 'LIGHT' },
  { label: '风机', value: 'FAN' },
  { label: '水泵', value: 'PUMP' },
  { label: '阀门', value: 'VALVE' }
]

function getGroupName(groupId: number | null): string {
  if (!groupId) return '-'
  const group = deviceGroups.value.find((g) => g.id === groupId)
  return group?.name || '-'
}

function goBack() { router.push('/devices') }

async function loadData() {
  if (!deviceId.value) { errorMsg.value = '无效的设备 ID'; return }
  loading.value = true
  errorMsg.value = ''
  try { await deviceStore.fetchDevice(deviceId.value) }
  catch (e: any) { errorMsg.value = e?.message || '加载设备信息失败'; return }
  finally { loading.value = false }
  deviceStore.fetchDeviceGroups().catch((e) => console.error(e))
  fetchHealth()
}

async function fetchHealth() {
  if (!deviceId.value) return
  try { health.value = await deviceStore.fetchDeviceHealth(deviceId.value) }
  catch (e) { console.error(e) }
}

async function fetchTelemetry() {
  if (!deviceId.value || !device.value || device.value.type !== 'SENSOR') return
  telemetryLoading.value = true
  try {
    const result = await getLatestTelemetry({ device_id: deviceId.value })
    telemetryData.value = result.items
  } catch (e) { console.error(e) }
  finally { telemetryLoading.value = false }
}

function openEditDialog() {
  if (!device.value) return
  editForm.name = device.value.name
  editForm.category = device.value.category
  editForm.group_id = device.value.group_id
  editForm.sampling_interval_sec = device.value.sampling_interval_sec || 60
  editDialogVisible.value = true
}

async function handleEdit() {
  if (!formRef.value || !deviceId.value) return
  try { await formRef.value.validate() } catch { return }
  submitLoading.value = true
  try {
    await deviceStore.editDevice(deviceId.value, {
      name: editForm.name, category: editForm.category,
      group_id: editForm.group_id, sampling_interval_sec: editForm.sampling_interval_sec
    })
    ElMessage.success('设备更新成功')
    editDialogVisible.value = false
  } catch { }
  finally { submitLoading.value = false }
}

async function toggleStatus() {
  if (!device.value) return
  const newStatus = device.value.status === 'ENABLED' ? 'DISABLED' : 'ENABLED'
  const action = newStatus === 'ENABLED' ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确定要${action}该设备吗？`, '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    await deviceStore.setDeviceStatus(deviceId.value, newStatus)
    ElMessage.success(`设备已${action}`)
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

watch(device, (val) => { if (val && val.type === 'SENSOR') fetchTelemetry() })
watch(selectedMetrics, () => { nextTick().then(() => renderCharts()) })

onMounted(() => { loadData() })
</script>

<style scoped lang="scss">
.device-detail-page {
  .page-header {
    display: flex; align-items: center; gap: 16px; margin-bottom: 16px;
  }
  .page-title {
    font-size: 18px; font-weight: 600; margin: 0; text-wrap: balance;
  }
  .loading-container, .error-container {
    padding: 40px; background: #fff; border-radius: 4px;
  }
  .info-card, .telemetry-card {
    margin-bottom: 16px;
  }
  .action-buttons {
    margin-top: 16px; display: flex; gap: 8px;
  }
  .loading-placeholder, .empty-placeholder {
    padding: 20px; text-align: center; color: #909399;
  }
  .telemetry-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 16px;
  }
  .telemetry-item {
    padding: 16px; background: #f5f7fa; border-radius: 4px;
  }
  .metric-name { font-size: 14px; color: #606266; margin-bottom: 8px; }
  .metric-value { font-size: 24px; font-weight: 600; color: #303133; margin-bottom: 8px; }
  .metric-unit { font-size: 14px; font-weight: normal; color: #909399; }
  .metric-meta { display: flex; align-items: center; gap: 8px; }
  .metric-time { font-size: 12px; color: #909399; }

  .dashboard-tab {
    .dashboard-controls {
      display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap;
    }
    .online-rate {
      display: flex; align-items: center; gap: 8px; margin-bottom: 16px;
    }
    .metric-cards {
      display: grid; grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 16px; margin-bottom: 16px;
    }
    .metric-card {
      .metric-unit-text { font-size: 12px; color: #909399; margin-left: 4px; }
      .metric-stats {
        display: flex; gap: 16px; flex-wrap: wrap;
      }
      .stat-item {
        text-align: center;
        .stat-label { display: block; font-size: 12px; color: #909399; }
        .stat-value { display: block; font-size: 18px; font-weight: 600; color: #303133; }
        .alert-count { color: #f56c6c; }
      }
    }
    .alert-events-card {
      margin-bottom: 16px;
    }
  }
}
</style>
