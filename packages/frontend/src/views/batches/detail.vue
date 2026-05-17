<template>
  <div class="batch-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/batches/ledger')" text>
          <el-icon><ArrowLeft /></el-icon>返回批次台账
        </el-button>
        <h1 class="page-title">{{ dashboard?.batch.batch_no || '批次详情' }}</h1>
        <el-tag :type="statusTagType" size="large">{{ dashboard?.batch.status }}</el-tag>
      </div>
      <div class="header-right" v-if="dashboard && canControl">
        <el-button type="default" @click="openEditDialog">
          <el-icon><Edit /></el-icon>编辑
        </el-button>
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
            <div class="info-value">
              <el-button v-if="dashboard.batch.greenhouse_id" type="primary" link size="small" @click="router.push('/assets/greenhouses')">
                {{ dashboard.greenhouse_name || '-' }}
              </el-button>
              <span v-else>{{ dashboard.greenhouse_name || '-' }}</span>
              <span> / </span>
              <el-button v-if="dashboard.batch.growing_zone_id" type="primary" link size="small" @click="router.push(`/assets/growing-zones?greenhouse_id=${dashboard.batch.greenhouse_id}`)">
                {{ dashboard.zone_name || '-' }}
              </el-button>
              <span v-else>{{ dashboard.zone_name || '-' }}</span>
            </div>
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

      <!-- Stage Plans Card -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-title">阶段计划 ({{ stagePlans.length }})</span>
            <el-button v-if="canControl" size="small" type="primary" @click="openCreateStageDialog">新增阶段</el-button>
          </div>
        </template>
        <el-alert
          v-if="stageConflictMessage"
          type="warning" :closable="false" show-icon
          :title="stageConflictMessage"
          class="conflict-alert"
        />
        <el-table v-if="stagePlans.length" :data="stagePlans" stripe v-loading="stageLoading" :row-class-name="stageRowClass" size="small">
          <el-table-column label="生长阶段" min-width="140">
            <template #default="{ row }">{{ growthStageLabelById[row.growth_stage_id] || fallbackIdLabel('阶段', row.growth_stage_id) }}</template>
          </el-table-column>
          <el-table-column label="配方" min-width="140">
            <template #default="{ row }">{{ recipeLabelById[row.recipe_id] || fallbackIdLabel('配方', row.recipe_id) }}</template>
          </el-table-column>
          <el-table-column label="策略" min-width="140">
            <template #default="{ row }">{{ policyLabelById[row.policy_id] || fallbackIdLabel('策略', row.policy_id) }}</template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="stageStatusTag(row)" size="small">{{ stageStatusText(row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="开始时间" min-width="160">
            <template #default="{ row }">{{ formatDateTime(row.stage_start_at) }}</template>
          </el-table-column>
          <el-table-column label="结束时间" min-width="160">
            <template #default="{ row }">{{ formatDateTime(row.stage_end_at) }}</template>
          </el-table-column>
          <el-table-column label="EC目标" width="110">
            <template #default="{ row }">{{ stageRange(row.target_ec_min, row.target_ec_max) }}</template>
          </el-table-column>
          <el-table-column label="pH目标" width="110">
            <template #default="{ row }">{{ stageRange(row.target_ph_min, row.target_ph_max) }}</template>
          </el-table-column>
          <el-table-column v-if="canControl" label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openEditStageDialog(row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="removeStage(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else-if="!stageLoading" description="暂无阶段计划" :image-size="60" />
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
                <el-button type="primary" link size="small" class="device-name" @click="router.push(`/devices/${d.device_id}?type=${d.device_type}`)">
                  {{ d.device_name || d.device_code || `#${d.device_id}` }}
                </el-button>
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
                class="alert-item clickable"
                @click="router.push(`/alerts/timeline?alertId=${a.id}`)"
                title="查看告警时间线"
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

      <!-- Edit Batch Dialog -->
      <el-dialog v-model="editDialogVisible" title="编辑批次" width="520px" @open="loadEditOptions">
        <el-form :model="editForm" label-width="100px">
          <el-form-item label="批次编号">
            <el-input v-model="editForm.batch_no" placeholder="批次编号" maxlength="64" />
          </el-form-item>
          <el-form-item label="温室">
            <el-select
              v-model="editForm.greenhouse_id"
              placeholder="选择温室"
              style="width: 100%"
              @change="onEditGreenhouseChange"
            >
              <el-option
                v-for="g in greenhouseOptions"
                :key="g.id"
                :label="g.name"
                :value="g.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="种植区">
            <el-select
              v-model="editForm.growing_zone_id"
              placeholder="选择种植区"
              clearable
              style="width: 100%"
            >
              <el-option
                v-for="z in zoneOptions"
                :key="z.id"
                :label="z.name"
                :value="z.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="作物品种">
            <el-select
              v-model="editForm.crop_variety_id"
              placeholder="选择品种"
              style="width: 100%"
            >
              <el-option
                v-for="v in varietyOptions"
                :key="v.id"
                :label="`${v.name} (${v.code})`"
                :value="v.id"
              />
            </el-select>
          </el-form-item>
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="定植密度" label-width="80px">
                <el-input-number v-model="editForm.planting_density" :min="0" :precision="1" style="width: 100%" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="总株数" label-width="70px">
                <el-input-number v-model="editForm.total_plants" :min="0" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item label="开始时间" label-width="80px">
                <el-date-picker
                  v-model="editForm.started_at"
                  type="date"
                  placeholder="选择日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="预计采收" label-width="70px">
                <el-date-picker
                  v-model="editForm.expected_harvest_at"
                  type="date"
                  placeholder="选择日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="备注">
            <el-input v-model="editForm.note" type="textarea" placeholder="备注信息" maxlength="500" show-word-limit />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="editing" @click="doEdit">保存</el-button>
        </template>
      </el-dialog>

      <!-- Stage Plan Editor Dialog -->
      <el-dialog v-model="stageEditorVisible" :title="editingStageId ? '编辑阶段计划' : '新增阶段计划'" width="760px">
        <stage-plan-editor v-model="stageEditorData" />
        <template #footer>
          <el-button @click="stageEditorVisible = false">取消</el-button>
          <el-button type="primary" :loading="stageSubmitLoading" @click="submitStage">保存</el-button>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, ArrowLeft, Edit } from '@element-plus/icons-vue'
import { cropApi, deviceApi, greenhouseApi, policyApi, recipeApi } from '@/api'
import StagePlanEditor from '@/components/batch/StagePlanEditor.vue'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, fallbackIdLabel, growthStageLabel } from '@/utils/labels'
import { usePermission } from '@/composables'
import type { BatchDashboard, BatchStagePlan, CreateBatchStagePlanRequest, GrowthStage, NutrientRecipe, ControlPolicy, SensorDevice, ActuatorDevice } from '@/types'

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
  const startStr = dashboard.value?.batch.started_at || dashboard.value?.batch.created_at
  if (!startStr) return 0
  const start = new Date(startStr).getTime()
  const end = dashboard.value?.batch.ended_at
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

// Batch edit
const editDialogVisible = ref(false)
const editing = ref(false)
const editForm = ref({
  batch_no: '',
  greenhouse_id: 0,
  growing_zone_id: undefined as number | undefined,
  crop_variety_id: 0,
  planting_density: undefined as number | undefined,
  total_plants: undefined as number | undefined,
  started_at: '',
  expected_harvest_at: '',
  note: ''
})
const greenhouseOptions = ref<{ id: number; name: string }[]>([])
const zoneOptions = ref<{ id: number; name: string }[]>([])
const varietyOptions = ref<{ id: number; name: string; code: string }[]>([])

function openEditDialog() {
  if (!dashboard.value) return
  const b = dashboard.value.batch
  editForm.value = {
    batch_no: b.batch_no || '',
    greenhouse_id: b.greenhouse_id || 0,
    growing_zone_id: b.growing_zone_id,
    crop_variety_id: b.crop_variety_id || 0,
    planting_density: b.planting_density,
    total_plants: b.total_plants,
    started_at: b.started_at?.slice(0, 10) || '',
    expected_harvest_at: b.expected_harvest_at?.slice(0, 10) || '',
    note: b.note || ''
  }
  editDialogVisible.value = true
}

async function loadEditOptions() {
  try {
    const [ghRes, varRes] = await Promise.all([
      greenhouseApi.getGreenhouses(),
      cropApi.getCropVarieties()
    ])
    greenhouseOptions.value = ghRes.items || []
    varietyOptions.value = varRes.items || []
  } catch { /* ignore */ }
  onEditGreenhouseChange(editForm.value.greenhouse_id)
}

async function onEditGreenhouseChange(greenhouseId: number | undefined) {
  if (!greenhouseId) {
    zoneOptions.value = []
    return
  }
  try {
    const res = await greenhouseApi.getGrowingZones({ greenhouse_id: greenhouseId })
    zoneOptions.value = res.items || []
  } catch {
    zoneOptions.value = []
  }
}

async function doEdit() {
  if (!dashboard.value) return
  editing.value = true
  try {
    const payload: Record<string, unknown> = {}
    const f = editForm.value
    if (f.batch_no) payload.batch_no = f.batch_no
    if (f.greenhouse_id) payload.greenhouse_id = f.greenhouse_id
    if (f.growing_zone_id !== undefined) payload.growing_zone_id = f.growing_zone_id
    if (f.crop_variety_id) payload.crop_variety_id = f.crop_variety_id
    if (f.planting_density !== undefined) payload.planting_density = f.planting_density
    if (f.total_plants !== undefined) payload.total_plants = f.total_plants
    if (f.started_at) payload.started_at = f.started_at
    if (f.expected_harvest_at) payload.expected_harvest_at = f.expected_harvest_at
    if (f.note !== undefined) payload.note = f.note

    await cropApi.updateBatch(dashboard.value.batch.id, payload as any)
    ElMessage.success('批次更新成功')
    editDialogVisible.value = false
    await fetchDashboard()
  } catch {
    ElMessage.error('更新失败')
  } finally {
    editing.value = false
  }
}

// Stage Plans
const stagePlans = ref<BatchStagePlan[]>([])
const stageLoading = ref(false)
const stageEditorVisible = ref(false)
const stageSubmitLoading = ref(false)
const editingStageId = ref<number>()
const stageEditorData = ref<CreateBatchStagePlanRequest>({
  batch_id: 0,
  growth_stage_id: 0,
  stage_start_at: new Date().toISOString(),
  stage_end_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
  target_ec_min: 1,
  target_ec_max: 2,
  target_ph_min: 5.5,
  target_ph_max: 6.5,
  recipe_id: undefined,
  policy_id: undefined,
  climate_profile_id: undefined
})

// Reference data for labels
const growthStages = ref<GrowthStage[]>([])
const recipes = ref<NutrientRecipe[]>([])
const policies = ref<ControlPolicy[]>([])

const growthStageLabelById = computed(() =>
  buildIdLabelMap(growthStages.value, s => s.id, growthStageLabel, '阶段')
)
const recipeLabelById = computed(() =>
  buildIdLabelMap(recipes.value, r => r.id, r => `${r.name} (${r.recipe_code})`, '配方')
)
const policyLabelById = computed(() =>
  buildIdLabelMap(policies.value, p => p.id, p => `${p.name} (${p.policy_code})`, '策略')
)

const nowForStage = computed(() => new Date())

const stageConflictMessage = computed(() => {
  const sorted = [...stagePlans.value].sort((a, b) => new Date(a.stage_start_at).getTime() - new Date(b.stage_start_at).getTime())
  for (let i = 1; i < sorted.length; i++) {
    const prevEnd = new Date(sorted[i - 1].stage_end_at).getTime()
    const currStart = new Date(sorted[i].stage_start_at).getTime()
    if (currStart < prevEnd) {
      return `阶段时间窗冲突：阶段之间存在重叠。`
    }
  }
  return ''
})

function stageStatusTag(stage: BatchStagePlan) {
  const start = new Date(stage.stage_start_at)
  const end = new Date(stage.stage_end_at)
  if (nowForStage.value < start) return 'info'
  if (nowForStage.value > end) return 'success'
  return ''
}

function stageStatusText(stage: BatchStagePlan) {
  const start = new Date(stage.stage_start_at)
  const end = new Date(stage.stage_end_at)
  if (nowForStage.value < start) return '未开始'
  if (nowForStage.value > end) return '已完成'
  return '进行中'
}

function stageRowClass({ row }: { row: BatchStagePlan }) {
  const start = new Date(row.stage_start_at)
  const end = new Date(row.stage_end_at)
  if (nowForStage.value < start) return 'stage-pending'
  if (nowForStage.value > end) return 'stage-completed'
  return 'stage-active'
}

function stageRange(min?: number | null, max?: number | null) {
  if (min == null && max == null) return '-'
  return `${min ?? '-'} ~ ${max ?? '-'}`
}

async function loadStagePlans() {
  if (!dashboard.value) return
  stageLoading.value = true
  try {
    const result = await cropApi.getBatchStagePlans({ batch_id: dashboard.value.batch.id })
    stagePlans.value = (result.items || []).sort(
      (a, b) => new Date(a.stage_start_at).getTime() - new Date(b.stage_start_at).getTime()
    )
  } catch {
    stagePlans.value = []
  } finally {
    stageLoading.value = false
  }
}

async function loadStageRefData() {
  try {
    const [stageRes, recipeRes, policyRes] = await Promise.all([
      cropApi.getGrowthStages({ page_size: 200 }),
      recipeApi.getRecipes({ page_size: 200 }),
      policyApi.getPolicies({ page_size: 200 })
    ])
    growthStages.value = stageRes.items
    recipes.value = recipeRes.items
    policies.value = policyRes.items
  } catch { /* ignore */ }
}

function openCreateStageDialog() {
  if (!dashboard.value) return
  editingStageId.value = undefined
  stageEditorData.value = {
    batch_id: dashboard.value.batch.id,
    growth_stage_id: growthStages.value[0]?.id || 0,
    stage_start_at: new Date().toISOString(),
    stage_end_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    target_ec_min: 1,
    target_ec_max: 2,
    target_ph_min: 5.5,
    target_ph_max: 6.5,
    recipe_id: undefined,
    policy_id: undefined,
    climate_profile_id: undefined
  }
  stageEditorVisible.value = true
}

function openEditStageDialog(stage: BatchStagePlan) {
  editingStageId.value = stage.id
  stageEditorData.value = {
    batch_id: stage.batch_id,
    growth_stage_id: stage.growth_stage_id,
    recipe_id: stage.recipe_id ?? undefined,
    policy_id: stage.policy_id ?? undefined,
    climate_profile_id: stage.climate_profile_id ?? undefined,
    stage_start_at: stage.stage_start_at,
    stage_end_at: stage.stage_end_at,
    target_ec_min: stage.target_ec_min ?? undefined,
    target_ec_max: stage.target_ec_max ?? undefined,
    target_ph_min: stage.target_ph_min ?? undefined,
    target_ph_max: stage.target_ph_max ?? undefined
  }
  stageEditorVisible.value = true
}

function validateStageInput(payload: CreateBatchStagePlanRequest) {
  if (!payload.growth_stage_id || !payload.stage_start_at || !payload.stage_end_at) return '请填写阶段和时间窗'
  const start = new Date(payload.stage_start_at).getTime()
  const end = new Date(payload.stage_end_at).getTime()
  if (start >= end) return '阶段结束时间必须晚于开始时间'
  const overlap = stagePlans.value.some((s) => {
    if (editingStageId.value && s.id === editingStageId.value) return false
    const sStart = new Date(s.stage_start_at).getTime()
    const sEnd = new Date(s.stage_end_at).getTime()
    return Math.max(start, sStart) < Math.min(end, sEnd)
  })
  if (overlap) return '阶段时间窗与现有阶段冲突'
  return ''
}

async function submitStage() {
  if (!dashboard.value) return
  const validation = validateStageInput(stageEditorData.value)
  if (validation) {
    ElMessage.error(validation)
    return
  }
  stageSubmitLoading.value = true
  try {
    if (editingStageId.value) {
      await cropApi.updateBatchStagePlan(editingStageId.value, stageEditorData.value)
      ElMessage.success('阶段计划已更新')
    } else {
      await cropApi.createBatchStagePlan(stageEditorData.value)
      ElMessage.success('阶段计划已创建')
    }
    stageEditorVisible.value = false
    await loadStagePlans()
  } finally {
    stageSubmitLoading.value = false
  }
}

async function removeStage(stageId: number) {
  try {
    await ElMessageBox.confirm('确认删除该阶段计划？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await cropApi.deleteBatchStagePlan(stageId)
    ElMessage.success('已删除')
    await loadStagePlans()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  fetchDashboard()
    .then(() => {
      loadStagePlans()
      loadStageRefData()
    })
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

    &.clickable {
      cursor: pointer;
      &:hover { background: var(--bg-page); }
    }

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
  .conflict-alert {
    margin-bottom: 10px;
  }
  :deep(.stage-completed) {
    background-color: rgba(103, 194, 58, 0.06);
  }
  :deep(.stage-active) {
    background-color: rgba(64, 158, 255, 0.06);
  }
  :deep(.stage-pending) {
    background-color: rgba(144, 147, 153, 0.04);
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
