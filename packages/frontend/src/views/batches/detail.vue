<template>
  <div class="batch-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.back()" :icon="'ArrowLeft'" text>返回</el-button>
        <h1 class="page-title">{{ dashboard?.batch.batch_no || '批次详情' }}</h1>
        <el-tag :type="statusTagType" size="large">{{ dashboard?.batch.status }}</el-tag>
      </div>
      <div class="header-right" v-if="dashboard && canControl">
        <el-dropdown @command="handleStatusTransition" v-if="allowedTransitions.length > 0">
          <el-button type="primary">
            状态转换 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="s in allowedTransitions"
                :key="s"
                :command="s"
              >{{ s }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <template v-if="dashboard">
      <!-- Info row -->
      <el-row :gutter="16" class="info-row">
        <el-col :span="6">
          <div class="info-card">
            <div class="info-label">品种</div>
            <div class="info-value">{{ dashboard.variety?.name || '-' }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="info-label">温室 / 种植区</div>
            <div class="info-value">{{ dashboard.greenhouse_name || '-' }} / {{ dashboard.zone_name || '-' }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="info-label">定植密度</div>
            <div class="info-value">{{ dashboard.batch.planting_density ?? '-' }} 株/㎡</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="info-label">运行天数</div>
            <div class="info-value">{{ runningDays }} 天</div>
          </div>
        </el-col>
      </el-row>

      <!-- Stage Progress Card -->
      <el-card class="section-card" v-if="dashboard.stage_progress">
        <template #header>
          <span class="card-title">阶段进度</span>
        </template>
        <div class="stage-info" v-if="dashboard.stage_progress.current_stage_name">
          <div class="stage-name">
            {{ dashboard.stage_progress.current_stage_name }}
            <el-tag size="small">{{ dashboard.stage_progress.current_stage_code }}</el-tag>
          </div>
          <el-progress
            :percentage="dashboard.stage_progress.progress_percent"
            :stroke-width="18"
            :color="progressColor"
          />
          <div class="stage-days">
            <span>已过 {{ dashboard.stage_progress.days_elapsed }} 天</span>
            <span>剩余 {{ dashboard.stage_progress.days_remaining }} 天</span>
          </div>
          <el-row :gutter="16" class="target-row">
            <el-col :span="12">
              <span class="target-label">目标 EC:</span>
              <span class="target-value">
                {{ dashboard.stage_progress.target_ec_min ?? '-' }} ~ {{ dashboard.stage_progress.target_ec_max ?? '-' }}
              </span>
            </el-col>
            <el-col :span="12">
              <span class="target-label">目标 pH:</span>
              <span class="target-value">
                {{ dashboard.stage_progress.target_ph_min ?? '-' }} ~ {{ dashboard.stage_progress.target_ph_max ?? '-' }}
              </span>
            </el-col>
          </el-row>
        </div>
        <el-empty v-else description="暂无阶段计划" :image-size="60" />
      </el-card>

      <el-row :gutter="16">
        <!-- Devices Card -->
        <el-col :span="12">
          <el-card class="section-card">
            <template #header>
              <div class="card-header-row">
                <span class="card-title">绑定设备 ({{ dashboard.devices?.length || 0 }})</span>
                <el-button v-if="canControl" size="small" type="primary" @click="openBindDialog">添加设备</el-button>
              </div>
            </template>
            <div v-if="dashboard.devices?.length">
              <div
                v-for="d in dashboard.devices"
                :key="d.id"
                class="device-item"
              >
                <el-tag :type="d.device_type === 'sensor' ? 'success' : 'warning'" size="small" effect="plain">
                  {{ d.device_type === 'sensor' ? '传感器' : '执行器' }}
                </el-tag>
                <span class="device-name">{{ d.device_name || d.device_code || `#${d.device_id}` }}</span>
                <span class="device-code">{{ d.device_code }}</span>
                <el-button
                  v-if="canControl"
                  size="small"
                  type="danger"
                  text
                  :loading="unbindingId === d.device_id"
                  @click="handleUnbind(d.device_type, d.device_id)"
                >解绑</el-button>
              </div>
            </div>
            <el-empty v-else description="未绑定设备" :image-size="60" />
          </el-card>
        </el-col>

        <!-- Latest Telemetry Card -->
        <el-col :span="12">
          <el-card class="section-card">
            <template #header>
              <span class="card-title">最新遥测</span>
            </template>
            <div v-if="dashboard.latest_telemetry?.length">
              <div
                v-for="t in dashboard.latest_telemetry.slice(0, 5)"
                :key="t.collected_at"
                class="telemetry-item"
              >
                <span class="metric-name">{{ t.metric_name || t.metric_code }}</span>
                <span class="metric-value">{{ t.value }} {{ t.unit }}</span>
                <span class="metric-time">{{ formatDateTime(t.collected_at) }}</span>
              </div>
            </div>
            <el-empty v-else description="暂无遥测数据" :image-size="60" />
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" style="margin-top: 16px">
        <!-- Recent Alerts Card -->
        <el-col :span="12">
          <el-card class="section-card">
            <template #header>
              <span class="card-title">待处理告警</span>
            </template>
            <div v-if="dashboard.recent_alerts?.length">
              <div
                v-for="a in dashboard.recent_alerts"
                :key="a.id"
                class="alert-item"
              >
                <el-tag :type="alertLevelTag(a.level)" size="small">{{ a.level }}</el-tag>
                <span class="alert-msg">{{ a.message }}</span>
                <span class="alert-time">{{ formatDateTime(a.triggered_at) }}</span>
              </div>
            </div>
            <el-empty v-else description="暂无告警" :image-size="60" />
          </el-card>
        </el-col>

        <!-- Recent Commands Card -->
        <el-col :span="12">
          <el-card class="section-card">
            <template #header>
              <span class="card-title">最近指令</span>
            </template>
            <div v-if="dashboard.recent_commands?.length">
              <div
                v-for="cmd in dashboard.recent_commands"
                :key="cmd.id"
                class="command-item"
              >
                <span class="cmd-type">{{ cmd.command_type }}</span>
                <el-tag :type="cmd.status === 'ACKED' ? 'success' : 'info'" size="small">{{ cmd.status }}</el-tag>
                <span class="cmd-time">{{ formatDateTime(cmd.created_at) }}</span>
              </div>
            </div>
            <el-empty v-else description="暂无指令记录" :image-size="60" />
          </el-card>
        </el-col>
      </el-row>

      <!-- Harvest Summary Card -->
      <el-card class="section-card" v-if="dashboard.harvest_summary">
        <template #header>
          <span class="card-title">采收汇总</span>
        </template>
        <div class="harvest-total">
          总产量: <strong>{{ dashboard.harvest_summary.total_weight_kg }} kg</strong>
        </div>
        <el-row :gutter="16">
          <el-col :span="6" v-for="g in dashboard.harvest_summary.grades" :key="g.grade">
            <div class="grade-card">
              <div class="grade-label">等级 {{ g.grade }}</div>
              <div class="grade-weight">{{ g.weight_kg }} kg</div>
              <div class="grade-count">{{ g.count }} 次</div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- Bind Device Dialog -->
      <el-dialog v-model="bindDialogVisible" title="绑定设备到批次" width="420px">
        <el-form :model="bindForm" label-width="80px">
          <el-form-item label="设备类型">
            <el-radio-group v-model="bindForm.device_type" @change="onBindTypeChange">
              <el-radio value="sensor">传感器</el-radio>
              <el-radio value="actuator">执行器</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="选择设备">
            <el-select
              v-model="bindForm.device_id"
              placeholder="选择设备"
              filterable
              style="width: 100%"
              :loading="availableDevicesLoading"
            >
              <el-option
                v-for="dev in availableDevices"
                :key="dev.id"
                :label="`${dev.name} (${dev.device_code})`"
                :value="dev.id"
              />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="bindDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="binding" :disabled="!bindForm.device_id" @click="doBind">确定绑定</el-button>
        </template>
      </el-dialog>

      <!-- Planting Record Card -->
      <el-card class="section-card" v-if="dashboard.planting_record">
        <template #header>
          <span class="card-title">定植记录</span>
        </template>
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="种子来源">{{ dashboard.planting_record.seed_source || '-' }}</el-descriptions-item>
          <el-descriptions-item label="种子批号">{{ dashboard.planting_record.seed_batch_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="苗龄">{{ dashboard.planting_record.seedling_age_days ?? '-' }} 天</el-descriptions-item>
          <el-descriptions-item label="播种时间">{{ formatDateTime(dashboard.planting_record.seeded_at) }}</el-descriptions-item>
          <el-descriptions-item label="定植时间">{{ formatDateTime(dashboard.planting_record.planted_at) }}</el-descriptions-item>
          <el-descriptions-item label="实际株数">{{ dashboard.planting_record.actual_plant_count ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="初始EC">{{ dashboard.planting_record.initial_ec ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="初始pH">{{ dashboard.planting_record.initial_ph ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="初始水温">{{ dashboard.planting_record.initial_water_temp ?? '-' }} °C</el-descriptions-item>
        </el-descriptions>
      </el-card>
    </template>

    <el-empty v-else-if="!loading" description="批次不存在" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { cropApi, deviceApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { usePermission } from '@/composables'
import type { BatchDashboard, SensorDevice, ActuatorDevice } from '@/types'

const route = useRoute()
const router = useRouter()
const { canControlDevice } = usePermission()

const loading = ref(true)
const dashboard = ref<BatchDashboard | null>(null)

const canControl = computed(() => canControlDevice())

const statusTagType = computed(() => {
  const map: Record<string, string> = {
    PLANNED: 'info',
    RUNNING: 'success',
    HARVESTING: 'warning',
    COMPLETED: '',
    ABORTED: 'danger'
  }
  return dashboard.value ? (map[dashboard.value.batch.status] || 'info') : 'info'
})

const runningDays = computed(() => {
  if (!dashboard.value?.batch.started_at) return 0
  const start = new Date(dashboard.value.batch.started_at).getTime()
  const end = dashboard.value.batch.ended_at
    ? new Date(dashboard.value.batch.ended_at).getTime()
    : Date.now()
  return Math.max(0, Math.floor((end - start) / (1000 * 60 * 60 * 24)))
})

// Legal transitions based on current status
const allowedTransitions = computed(() => {
  if (!dashboard.value) return []
  const transitions: Record<string, string[]> = {
    PLANNED: ['RUNNING', 'ABORTED'],
    RUNNING: ['HARVESTING', 'ABORTED'],
    HARVESTING: ['COMPLETED', 'ABORTED'],
    COMPLETED: [],
    ABORTED: []
  }
  return transitions[dashboard.value.batch.status] || []
})

const progressColor = computed(() => {
  const p = dashboard.value?.stage_progress?.progress_percent || 0
  if (p < 30) return '#409EFF'
  if (p < 70) return '#E6A23C'
  return '#67C23A'
})

// Device binding
const bindDialogVisible = ref(false)
const binding = ref(false)
const unbindingId = ref<number | null>(null)
const availableDevices = ref<(SensorDevice | ActuatorDevice)[]>([])
const availableDevicesLoading = ref(false)
const bindForm = ref({ device_type: 'sensor' as 'sensor' | 'actuator', device_id: null as number | null })

function openBindDialog() {
  bindForm.value = { device_type: 'sensor', device_id: null }
  bindDialogVisible.value = true
  loadAvailableDevices()
}

function onBindTypeChange() {
  bindForm.value.device_id = null
  loadAvailableDevices()
}

async function loadAvailableDevices() {
  if (!dashboard.value) return
  availableDevicesLoading.value = true
  try {
    const params: Record<string, unknown> = {
      greenhouse_id: dashboard.value.batch.greenhouse_id,
      page_size: 200
    }
    if (bindForm.value.device_type === 'sensor') {
      const res = await deviceApi.getSensorDevices(params)
      availableDevices.value = res.items
    } else {
      const res = await deviceApi.getActuatorDevices(params)
      availableDevices.value = res.items
    }
  } catch {
    availableDevices.value = []
  } finally {
    availableDevicesLoading.value = false
  }
}

async function doBind() {
  if (!dashboard.value || !bindForm.value.device_id) return
  binding.value = true
  try {
    await cropApi.bindDevice(dashboard.value.batch.id, {
      device_type: bindForm.value.device_type,
      device_id: bindForm.value.device_id
    })
    ElMessage.success('设备已绑定')
    bindDialogVisible.value = false
    await fetchDashboard()
  } catch {
    ElMessage.error('绑定失败，设备可能已绑定')
  } finally {
    binding.value = false
  }
}

async function handleUnbind(deviceType: string, deviceId: number) {
  if (!dashboard.value) return
  unbindingId.value = deviceId
  try {
    await cropApi.unbindDevice(dashboard.value.batch.id, deviceId, deviceType)
    ElMessage.success('已解绑')
    await fetchDashboard()
  } catch {
    ElMessage.error('解绑失败')
  } finally {
    unbindingId.value = null
  }
}

async function fetchDashboard() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    if (!id) {
      router.replace('/batches/ledger')
      return
    }
    dashboard.value = await cropApi.getBatchDashboard(id)
  } catch {
    ElMessage.error('加载批次详情失败')
  } finally {
    loading.value = false
  }
}

async function handleStatusTransition(newStatus: string) {
  if (!dashboard.value) return
  try {
    await cropApi.transitionBatch(dashboard.value.batch.id, { status: newStatus })
    ElMessage.success(`状态已变更为 ${newStatus}`)
    await fetchDashboard()
  } catch {
    ElMessage.error('状态转换失败')
  }
}

function alertLevelTag(level: string) {
  const map: Record<string, string> = { INFO: 'info', WARN: 'warning', CRITICAL: 'danger' }
  return map[level] || 'info'
}

onMounted(() => {
  fetchDashboard()
})
</script>

<style scoped lang="scss">
.batch-detail-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .page-title {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
    }
  }

  .info-row {
    margin-bottom: 16px;
  }
  .info-card {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
    padding: 16px;
    text-align: center;
    .info-label {
      font-size: 13px;
      color: var(--text-secondary);
      margin-bottom: 4px;
    }
    .info-value {
      font-size: 16px;
      font-weight: 600;
    }
  }

  .section-card {
    margin-bottom: 16px;
    .card-title {
      font-weight: 600;
    }
    .card-header-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .stage-info {
    .stage-name {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .stage-days {
      display: flex;
      justify-content: space-between;
      margin-top: 8px;
      font-size: 13px;
      color: var(--text-secondary);
    }
    .target-row {
      margin-top: 12px;
      font-size: 14px;
      .target-label {
        color: var(--text-secondary);
      }
      .target-value {
        font-weight: 600;
        margin-left: 4px;
      }
    }
  }

  .device-item, .telemetry-item, .alert-item, .command-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
    border-bottom: 1px solid var(--border-light);
    &:last-child { border-bottom: none; }

    .device-name, .metric-name, .alert-msg, .cmd-type {
      flex: 1;
      font-weight: 500;
    }
    .device-code, .metric-value, .alert-time, .cmd-time {
      font-size: 13px;
      color: var(--text-secondary);
    }
    .metric-value {
      font-weight: 600;
      color: var(--text-primary);
    }
  }

  .harvest-total {
    font-size: 18px;
    margin-bottom: 12px;
    strong { color: var(--color-primary); }
  }
  .grade-card {
    text-align: center;
    padding: 12px;
    background: var(--bg-page);
    border-radius: var(--radius-md);
    .grade-label { font-size: 14px; font-weight: 600; }
    .grade-weight { font-size: 18px; font-weight: 700; color: var(--color-primary); }
    .grade-count { font-size: 12px; color: var(--text-secondary); }
  }
}
</style>
