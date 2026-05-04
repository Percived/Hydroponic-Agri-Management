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
        <!-- 基本信息 -->
        <el-card class="info-card">
          <template #header>
            <span>基本信息</span>
          </template>
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
          <template #header>
            <span>最新遥测数据</span>
          </template>
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { AppLayout } from '@/components/layout'
import { useDeviceStore } from '@/stores/device'
import { usePermission } from '@/composables/usePermission'
import { getCategoryName, formatDate, formatNumber } from '@/utils/format'
import { getLatestTelemetry } from '@/api/telemetry'
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

const deviceId = computed(() => Number(route.params.id))

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

// 获取分组名称
function getGroupName(groupId: number | null): string {
  if (!groupId) return '-'
  const group = deviceGroups.value.find((g) => g.id === groupId)
  return group?.name || '-'
}

// 返回列表
function goBack() {
  router.push('/devices')
}

// 加载所有数据
async function loadData() {
  if (!deviceId.value) {
    errorMsg.value = '无效的设备 ID'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    await deviceStore.fetchDevice(deviceId.value)
  } catch (e: any) {
    console.error('[Device Detail] Failed to fetch device:', e)
    errorMsg.value = e?.message || '加载设备信息失败'
    return
  } finally {
    loading.value = false
  }

  // 加载辅助数据（不阻塞页面渲染）
  deviceStore.fetchDeviceGroups().catch((e) => {
    console.error('[Device Detail] Failed to fetch groups:', e)
  })
  fetchHealth()
}

// 获取设备健康状态
async function fetchHealth() {
  if (!deviceId.value) return
  try {
    health.value = await deviceStore.fetchDeviceHealth(deviceId.value)
  } catch (error) {
    console.error('[Device Detail] Failed to fetch health:', error)
  }
}

// 获取遥测数据
async function fetchTelemetry() {
  if (!deviceId.value || !device.value || device.value.type !== 'SENSOR') return
  telemetryLoading.value = true
  try {
    const result = await getLatestTelemetry({ device_id: deviceId.value })
    telemetryData.value = result.items
  } catch (error) {
    console.error('[Device Detail] Failed to fetch telemetry:', error)
  } finally {
    telemetryLoading.value = false
  }
}

// 打开编辑弹窗
function openEditDialog() {
  if (!device.value) return
  editForm.name = device.value.name
  editForm.category = device.value.category
  editForm.group_id = device.value.group_id
  editForm.sampling_interval_sec = device.value.sampling_interval_sec || 60
  editDialogVisible.value = true
}

// 提交编辑
async function handleEdit() {
  if (!formRef.value || !deviceId.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    await deviceStore.editDevice(deviceId.value, {
      name: editForm.name,
      category: editForm.category,
      group_id: editForm.group_id,
      sampling_interval_sec: editForm.sampling_interval_sec
    })
    ElMessage.success('设备更新成功')
    editDialogVisible.value = false
  } catch (error) {
    console.error('[Device Detail] Failed to update device:', error)
  } finally {
    submitLoading.value = false
  }
}

// 切换状态
async function toggleStatus() {
  if (!device.value) return
  const newStatus = device.value.status === 'ENABLED' ? 'DISABLED' : 'ENABLED'
  const action = newStatus === 'ENABLED' ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(`确定要${action}该设备吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deviceStore.setDeviceStatus(deviceId.value, newStatus)
    ElMessage.success(`设备已${action}`)
  } catch (error) {
    // 用户取消操作不记录错误
    if (error !== 'cancel') {
      console.error('[Device Detail] Failed to toggle status:', error)
    }
  }
}

// 监听设备变化获取遥测数据
watch(device, (val) => {
  if (val && val.type === 'SENSOR') {
    fetchTelemetry()
  }
})

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.device-detail-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
  }

  .loading-container,
  .error-container {
    padding: 40px;
    background: #fff;
    border-radius: 4px;
  }

  .info-card,
  .telemetry-card {
    margin-bottom: 16px;
  }

  .action-buttons {
    margin-top: 16px;
    display: flex;
    gap: 8px;
  }

  .loading-placeholder,
  .empty-placeholder {
    padding: 20px;
    text-align: center;
    color: #909399;
  }

  .telemetry-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }

  .telemetry-item {
    padding: 16px;
    background: #f5f7fa;
    border-radius: 4px;
  }

  .metric-name {
    font-size: 14px;
    color: #606266;
    margin-bottom: 8px;
  }

  .metric-value {
    font-size: 24px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
  }

  .metric-unit {
    font-size: 14px;
    font-weight: normal;
    color: #909399;
  }

  .metric-meta {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .metric-time {
    font-size: 12px;
    color: #909399;
  }
}
</style>