<template>
  <div class="climate-profile-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/strategy/climate')" text>
          <el-icon><ArrowLeft /></el-icon>返回气候配置列表
        </el-button>
        <h1 class="page-title">{{ profile?.name || '气候配置详情' }}</h1>
        <el-tag v-if="profile" size="large" :type="profile.enabled ? 'success' : 'info'">
          {{ profile.enabled ? '已启用' : '已停用' }}
        </el-tag>
      </div>
      <div class="header-right" v-if="profile">
        <el-button v-if="canManage" type="primary" @click="openEditProfile">
          <el-icon><Edit /></el-icon>编辑配置
        </el-button>
        <el-button v-if="canManage" type="success" @click="openExecuteDialog">
          <el-icon><VideoPlay /></el-icon>手动执行
        </el-button>
        <el-button v-if="canDelete" type="danger" @click="handleDeleteProfile">
          <el-icon><Delete /></el-icon>删除
        </el-button>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading && !profile" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <!-- Error state -->
    <el-result
      v-else-if="errorMsg"
      icon="error"
      :title="errorMsg"
      sub-title="请检查网络连接或返回列表重试"
    >
      <template #extra>
        <el-button type="primary" @click="loadAllData">重新加载</el-button>
        <el-button @click="router.push('/strategy/climate')">返回列表</el-button>
      </template>
    </el-result>

    <!-- Content -->
    <template v-else-if="profile">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <el-card class="info-card" shadow="never">
            <template #header><span class="card-title">配置信息</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="配置编码">{{ profile.code }}</el-descriptions-item>
              <el-descriptions-item label="配置名称">{{ profile.name }}</el-descriptions-item>
              <el-descriptions-item label="所属温室">{{ greenhouseName(profile.greenhouse_id) }}</el-descriptions-item>
              <el-descriptions-item label="触发指标">{{ getMetricName(profile.trigger_metric_code) }}</el-descriptions-item>
              <el-descriptions-item label="触发采集通道">{{ sensorChannelName(profile.trigger_sensor_channel_id) }}</el-descriptions-item>
              <el-descriptions-item label="启用状态">
                <el-tag :type="profile.enabled ? 'success' : 'info'">{{ profile.enabled ? '已启用' : '已停用' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="描述">{{ profile.description || '-' }}</el-descriptions-item>
              <el-descriptions-item label="阶段数量">{{ stageCount }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDateTime(profile.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatDateTime(profile.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-row :gutter="16" class="stats-row">
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-value">{{ stageCount }}</div>
                <div class="stat-label">阶段总数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-value">{{ totalActions }}</div>
                <div class="stat-label">动作总数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-value">{{ execTotal }}</div>
                <div class="stat-label">执行次数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-value">{{ lastExecutedAt }}</div>
                <div class="stat-label">最近执行</div>
              </div>
            </el-col>
          </el-row>
        </el-tab-pane>

        <!-- Tab 2: 阶段配置 -->
        <el-tab-pane label="阶段配置" name="stages">
          <div class="tab-toolbar">
            <el-button v-if="canManage" type="primary" @click="openCreateStage">
              <el-icon><Plus /></el-icon>新增阶段
            </el-button>
          </div>

          <el-table
            :data="stages"
            v-loading="stagesLoading"
            stripe
            row-key="id"
          >
            <el-table-column type="expand">
              <template #default="{ row: stage }">
                <div class="expanded-section" v-loading="actionsLoading[stage.id]">
                  <div class="expanded-header">
                    <span class="section-subtitle">动作列表</span>
                    <el-button v-if="canManage" type="primary" size="small" @click.stop="openCreateAction(stage)">
                      <el-icon><Plus /></el-icon>新增动作
                    </el-button>
                  </div>
                  <el-table
                    v-if="actionsByStage[stage.id]?.length"
                    :data="actionsByStage[stage.id]"
                    size="small"
                    border
                  >
                    <el-table-column prop="id" label="ID" width="60" />
                    <el-table-column label="执行器通道" min-width="200">
                      <template #default="{ row: action }">{{ actuatorChannelName(action.actuator_channel_id) }}</template>
                    </el-table-column>
                    <el-table-column prop="command_type" label="命令类型" width="120" />
                    <el-table-column label="命令参数" min-width="200">
                      <template #default="{ row: action }">{{ formatCommandPayload(action.command_payload) }}</template>
                    </el-table-column>
                    <el-table-column prop="execution_order" label="顺序" width="70" />
                    <el-table-column label="启用" width="70">
                      <template #default="{ row: action }">
                        <el-tag :type="action.enabled ? 'success' : 'info'" size="small">
                          {{ action.enabled ? '是' : '否' }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column v-if="canManage" label="操作" width="120" fixed="right">
                      <template #default="{ row: action }">
                        <el-button type="primary" link size="small" @click.stop="openEditAction(stage, action)">编辑</el-button>
                        <el-button type="danger" link size="small" @click.stop="removeAction(stage.id, action.id)">删除</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                  <el-empty v-else-if="!actionsLoading[stage.id]" description="暂无动作" :image-size="40" />
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="stage_level" label="级别" width="70" />
            <el-table-column prop="name" label="名称" width="140" />
            <el-table-column label="触发条件" min-width="200">
              <template #default="{ row: stage }">
                {{ stage.trigger_operator }} {{ stage.trigger_threshold }}
                <span v-if="stage.hysteresis" class="hint">(回差: {{ stage.hysteresis }})</span>
              </template>
            </el-table-column>
            <el-table-column label="动作数" width="80">
              <template #default="{ row: stage }">{{ stage.action_count ?? stage.actions?.length ?? 0 }}</template>
            </el-table-column>
            <el-table-column v-if="canManage" label="操作" width="120" fixed="right">
              <template #default="{ row: stage }">
                <el-button type="primary" link size="small" @click.stop="openEditStage(stage)">编辑</el-button>
                <el-button type="danger" link size="small" @click.stop="removeStage(stage.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!stagesLoading && stages.length === 0" description="暂无阶段配置" />
        </el-tab-pane>

        <!-- Tab 3: 执行日志 -->
        <el-tab-pane label="执行日志" name="logs">
          <div class="tab-toolbar">
            <el-date-picker
              v-model="logsFilters.from"
              type="datetime"
              placeholder="开始时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              style="width: 190px"
            />
            <span class="filter-separator">—</span>
            <el-date-picker
              v-model="logsFilters.to"
              type="datetime"
              placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              style="width: 190px"
            />
            <el-button type="primary" @click="fetchExecutionLogs">查询</el-button>
            <el-button @click="resetLogsFilters">重置</el-button>
          </div>

          <el-table :data="execLogs" v-loading="logsLoading" stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="from_stage_level" label="从级别" width="80">
              <template #default="{ row }">{{ row.from_stage_level ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="to_stage_level" label="到级别" width="80" />
            <el-table-column prop="trigger_value" label="触发值" width="100" />
            <el-table-column prop="trigger_sensor_channel_id" label="触发通道" width="120">
              <template #default="{ row }">{{ sensorChannelName(row.trigger_sensor_channel_id) }}</template>
            </el-table-column>
            <el-table-column prop="trigger_metric_code" label="指标" width="90">
              <template #default="{ row }">{{ row.trigger_metric_code ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="executed_actions_count" label="执行动作数" width="110" />
            <el-table-column label="采集时间" width="180">
              <template #default="{ row }">{{ row.collected_at ? formatDateTime(row.collected_at) : '-' }}</template>
            </el-table-column>
            <el-table-column label="执行时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.executed_at) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!logsLoading && execLogs.length === 0" description="暂无执行日志" />

          <div v-if="execTotal > 0" class="pagination-container">
            <el-pagination
              v-model:current-page="logsPagination.page"
              v-model:page-size="logsPagination.pageSize"
              :total="execTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchExecutionLogs"
              @current-change="fetchExecutionLogs"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else description="气候配置不存在" />

    <!-- Dialog: Edit Profile -->
    <el-dialog v-model="profileDialogVisible" title="编辑气候配置" width="500px">
      <el-form ref="profileFormRef" :model="profileForm" :rules="profileFormRules" label-width="120px">
        <el-form-item label="温室" prop="greenhouse_id">
          <el-select v-model="profileForm.greenhouse_id" placeholder="选择温室" filterable disabled style="width: 100%">
            <el-option v-for="gh in greenhouses" :key="gh.id" :label="`${gh.name} (ID:${gh.id})`" :value="gh.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="设备" prop="trigger_sensor_device_id">
          <el-select v-model="profileForm.trigger_sensor_device_id" placeholder="选择采集设备" filterable style="width: 100%" @change="onEditProfileDeviceChange">
            <el-option v-for="d in editSensorDevices" :key="d.id" :label="`${d.name} (${d.device_code})`" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="profileForm.code" maxlength="64" disabled />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="profileForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="profileForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="触发指标" prop="trigger_metric_code">
          <el-select v-model="profileForm.trigger_metric_code" placeholder="选择指标" filterable style="width: 100%" @change="onEditProfileMetricChange">
            <el-option v-for="m in metrics" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="采集通道" prop="trigger_sensor_channel_id">
          <el-select v-model="profileForm.trigger_sensor_channel_id" placeholder="选择采集通道" filterable style="width: 100%">
            <el-option
              v-for="ch in filteredEditSensorChannels"
              :key="ch.id"
              :label="sensorChannelOptionLabel(ch)"
              :value="ch.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="profileForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="profileDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="profileSubmitLoading" @click="handleProfileSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Stage Form -->
    <el-dialog v-model="stageFormVisible" :title="isEditStage ? '编辑阶段' : '新增阶段'" width="500px">
      <el-form ref="stageFormRef" :model="stageForm" :rules="stageFormRules" label-width="120px">
        <el-form-item label="级别" prop="stage_level">
          <el-input-number v-model="stageForm.stage_level" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="stageForm.name" />
        </el-form-item>
        <el-form-item label="触发操作符" prop="trigger_operator">
          <el-select v-model="stageForm.trigger_operator" style="width: 100%">
            <el-option label="&gt;" value="&gt;" />
            <el-option label="&gt;=" value="&gt;=" />
            <el-option label="&lt;" value="&lt;" />
            <el-option label="&lt;=" value="&lt;=" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发阈值" prop="trigger_threshold">
          <el-input-number v-model="stageForm.trigger_threshold" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="回差">
          <el-input-number v-model="stageForm.hysteresis" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stageFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="stageSubmitLoading" @click="handleStageSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Action Form -->
    <el-dialog v-model="actionFormVisible" :title="isEditAction ? '编辑动作' : '新增动作'" width="500px">
      <el-form ref="actionFormRef" :model="actionForm" :rules="actionFormRules" label-width="120px">
        <el-form-item label="执行器通道" prop="actuator_channel_id">
          <el-select v-model="actionForm.actuator_channel_id" placeholder="选择执行器通道" filterable style="width: 100%">
            <el-option v-for="ch in actuatorChannels" :key="ch.id" :label="actuatorChannelName(ch.id)" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令类型" prop="command_type">
          <el-select v-model="actionForm.command_type" style="width: 100%">
            <el-option label="SWITCH" value="SWITCH" />
            <el-option label="SET_VALUE" value="SET_VALUE" />
            <el-option label="CALIBRATE" value="CALIBRATE" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行顺序">
          <el-input-number v-model="actionForm.execution_order" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="actionForm.enabled" />
        </el-form-item>
        <el-form-item label="命令参数" prop="command_payload_str">
          <el-input
            v-model="actionForm.command_payload_str"
            type="textarea"
            :rows="4"
            placeholder='JSON 格式，如 {"value": 1}'
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="actionFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="actionSubmitLoading" @click="handleActionSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- Dialog: Execute -->
    <el-dialog v-model="executeDialogVisible" title="手动执行气候配置" width="450px">
      <el-form ref="executeFormRef" :model="executeForm" :rules="executeFormRules" label-width="120px">
        <el-form-item label="触发值" prop="trigger_value">
          <el-input-number v-model="executeForm.trigger_value" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="目标阶段级别" prop="to_stage_level">
          <el-input-number v-model="executeForm.to_stage_level" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="来源阶段级别">
          <el-input-number v-model="executeForm.from_stage_level" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="executeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="executeLoading" @click="handleExecute">执行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft, Edit, VideoPlay, Delete, Plus } from '@element-plus/icons-vue'
import { climateApi, greenhouseApi, deviceApi, metricApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime } from '@/utils/format'
import {
  buildIdLabelMap, fallbackIdLabel, greenhouseLabel,
  sensorChannelLabel, sensorDeviceLabel,
  actuatorChannelLabel, actuatorDeviceLabel
} from '@/utils/labels'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type {
  ClimateProfile, ClimateStage, ClimateStageAction, ClimateExecutionLog,
  Greenhouse, ActuatorChannel, ActuatorDevice, MetricDefinition,
  SensorDevice, SensorChannel
} from '@/types'
import { Role } from '@/types'

const route = useRoute()
const router = useRouter()

const profileId = computed(() => Number(route.params.id))
const { hasRole, canControlDevice } = usePermission()
const canManage = computed(() => canControlDevice())
const canDelete = computed(() => hasRole(Role.ADMIN))

// ── State ──
const loading = ref(false)
const errorMsg = ref('')
const profile = ref<ClimateProfile | null>(null)
const activeTab = ref('info')

const stages = ref<ClimateStage[]>([])
const stagesLoading = ref(false)
const actionsByStage = reactive<Record<number, ClimateStageAction[]>>({})
const actionsLoading = reactive<Record<number, boolean>>({})

const execLogs = ref<ClimateExecutionLog[]>([])
const execTotal = ref(0)
const logsLoading = ref(false)
const logsPagination = reactive({ page: 1, pageSize: 20 })
const logsFilters = reactive({ from: '', to: '' })

// ── Lookup data ──
const greenhouses = ref<Greenhouse[]>([])
const actuatorDevices = ref<ActuatorDevice[]>([])
const actuatorChannels = ref<ActuatorChannel[]>([])
const metrics = ref<MetricDefinition[]>([])
const sensorDevices = ref<SensorDevice[]>([])
const sensorChannelLabelCache = ref<Record<number, string>>({})
const sensorDeviceLabelCache = ref<Record<number, string>>({})

const greenhouseLabelById = computed(() =>
  buildIdLabelMap(greenhouses.value, gh => gh.id, greenhouseLabel, '温室')
)

function greenhouseName(greenhouseId?: number | null) {
  if (!greenhouseId) return fallbackIdLabel('温室', greenhouseId)
  return greenhouseLabelById.value[greenhouseId] || fallbackIdLabel('温室', greenhouseId)
}

const sensorDeviceLabelById = computed(() =>
  buildIdLabelMap(sensorDevices.value, d => d.id, sensorDeviceLabel, '设备')
)

const actuatorDeviceLabelById = computed(() =>
  buildIdLabelMap(actuatorDevices.value, d => d.id, actuatorDeviceLabel, '设备')
)

const actuatorChannelLabelById = computed(() => {
  const map: Record<number, string> = {}
  for (const ch of actuatorChannels.value) {
    map[ch.id] = actuatorChannelLabel(ch, actuatorDeviceLabelById.value)
  }
  return map
})

function actuatorChannelName(channelId?: number | null) {
  if (!channelId) return '-'
  return actuatorChannelLabelById.value[channelId] || fallbackIdLabel('通道', channelId)
}

function sensorChannelName(channelId?: number | null) {
  if (!channelId) return '-'
  if (!sensorChannelLabelCache.value[channelId]) {
    ensureSensorChannelLabel(channelId)
  }
  return sensorChannelLabelCache.value[channelId] || fallbackIdLabel('通道', channelId)
}

function sensorChannelOptionLabel(ch: SensorChannel) {
  return sensorChannelLabel(ch, sensorDeviceLabelById.value)
}

async function ensureSensorChannelLabel(channelId: number) {
  if (sensorChannelLabelCache.value[channelId]) return
  try {
    const ch = await deviceApi.getSensorChannel(channelId)
    let devLabel = sensorDeviceLabelById.value[ch.sensor_device_id] || sensorDeviceLabelCache.value[ch.sensor_device_id]
    if (!devLabel) {
      try {
        const d = await deviceApi.getSensorDevice(ch.sensor_device_id)
        devLabel = sensorDeviceLabel(d)
        sensorDeviceLabelCache.value[ch.sensor_device_id] = devLabel
      } catch {
        devLabel = fallbackIdLabel('设备', ch.sensor_device_id)
      }
    }
    sensorChannelLabelCache.value[channelId] = sensorChannelLabel(ch, {
      ...sensorDeviceLabelById.value,
      ...sensorDeviceLabelCache.value
    })
  } catch {
    sensorChannelLabelCache.value[channelId] = fallbackIdLabel('通道', channelId)
  }
}

const metricLabelById = computed(() => {
  const map: Record<string, string> = {}
  for (const m of metrics.value) {
    map[m.code] = `${m.name} (${m.code})`
  }
  return map
})

function getMetricName(code: string) {
  return metricLabelById.value[code] || code
}

const stageCount = computed(() => profile.value?.stages?.length ?? profile.value?.stages_count ?? 0)

const totalActions = computed(() => {
  if (profile.value?.stages) {
    return profile.value.stages.reduce((sum, s) => sum + (s.actions?.length ?? s.action_count ?? 0), 0)
  }
  return 0
})

const lastExecutedAt = computed(() => {
  if (execLogs.value.length > 0) {
    return formatDateTime(execLogs.value[0].executed_at)
  }
  return '-'
})

function formatCommandPayload(payload: unknown) {
  if (!payload) return '-'
  return JSON.stringify(payload)
}

// ── Data Loading ──
async function loadReferenceData() {
  await Promise.allSettled([
    loadGreenhouses(),
    loadActuatorDevices(),
    loadActuatorChannels(),
    loadSensorDevices(),
    loadMetrics()
  ])
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { greenhouses.value = [] }
}

async function loadActuatorDevices() {
  try {
    const data = await deviceApi.getActuatorDevices({ page_size: LARGE_PAGE_SIZE })
    actuatorDevices.value = data.items
  } catch { actuatorDevices.value = [] }
}

async function loadActuatorChannels() {
  try {
    const data = await deviceApi.getActuatorChannels({ page_size: LARGE_PAGE_SIZE })
    actuatorChannels.value = data.items
  } catch { actuatorChannels.value = [] }
}

async function loadSensorDevices() {
  try {
    const data = await deviceApi.getSensorDevices({ page_size: LARGE_PAGE_SIZE })
    sensorDevices.value = data.items
  } catch { sensorDevices.value = [] }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    metrics.value = data.items
  } catch { metrics.value = [] }
}

async function loadProfile() {
  if (!profileId.value) {
    errorMsg.value = '无效的气候配置 ID'
    return
  }
  try {
    const data = await climateApi.getClimateProfile(profileId.value)
    profile.value = data
    // Populate stages and their actions
    if (data.stages) {
      stages.value = data.stages
      for (const stage of data.stages) {
        if (stage.actions) {
          actionsByStage[stage.id] = stage.actions
        }
      }
    }
    // Warm sensor channel labels
    if (data.trigger_sensor_channel_id) {
      ensureSensorChannelLabel(data.trigger_sensor_channel_id)
    }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载气候配置失败'
    errorMsg.value = msg
  }
}

async function loadAllData() {
  errorMsg.value = ''
  loading.value = true
  await Promise.allSettled([loadProfile(), loadReferenceData()])
  loading.value = false
}

async function refreshStages() {
  if (!profileId.value) return
  stagesLoading.value = true
  try {
    const data = await climateApi.getClimateProfileStages(profileId.value)
    stages.value = data.items
    // Refresh actions for any expanded stages
    for (const stage of data.items) {
      if (actionsByStage[stage.id] !== undefined) {
        refreshActionsForStage(stage.id)
      }
    }
  } catch {
    stages.value = []
  } finally {
    stagesLoading.value = false
  }
}

async function refreshActionsForStage(stageId: number) {
  if (!profileId.value) return
  actionsLoading[stageId] = true
  try {
    const data = await climateApi.getClimateStageActions(profileId.value, stageId)
    actionsByStage[stageId] = data.items
  } catch {
    actionsByStage[stageId] = []
  } finally {
    actionsLoading[stageId] = false
  }
}

async function fetchExecutionLogs() {
  logsLoading.value = true
  try {
    const params: Record<string, unknown> = {
      page: logsPagination.page,
      page_size: logsPagination.pageSize
    }
    if (logsFilters.from) params.from = logsFilters.from
    if (logsFilters.to) params.to = logsFilters.to
    const data = await climateApi.getClimateProfileExecutionLogs(profileId.value, params)
    execLogs.value = data.items
    execTotal.value = data.total
  } catch {
    execLogs.value = []
    execTotal.value = 0
  } finally {
    logsLoading.value = false
  }
}

function resetLogsFilters() {
  logsFilters.from = ''
  logsFilters.to = ''
  logsPagination.page = 1
  fetchExecutionLogs()
}

// ── Profile Edit Dialog ──
const profileDialogVisible = ref(false)
const profileFormRef = ref<FormInstance>()
const profileSubmitLoading = ref(false)
const editSensorDevices = ref<SensorDevice[]>([])
const editSensorChannels = ref<SensorChannel[]>([])

const profileForm = reactive({
  greenhouse_id: undefined as number | undefined,
  trigger_sensor_device_id: undefined as number | undefined,
  code: '',
  name: '',
  description: '' as string,
  trigger_metric_code: '',
  trigger_sensor_channel_id: undefined as number | undefined,
  enabled: true
})

const profileFormRules: FormRules = {
  trigger_sensor_device_id: [{ required: true, message: '请选择采集设备', trigger: 'change' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  trigger_metric_code: [{ required: true, message: '请选择触发指标', trigger: 'change' }],
  trigger_sensor_channel_id: [{ required: true, message: '请选择采集通道', trigger: 'change' }]
}

const filteredEditSensorChannels = computed(() => {
  const metric = profileForm.trigger_metric_code
  if (!metric) return editSensorChannels.value
  return editSensorChannels.value.filter(ch => ch.metric_code === metric)
})

async function openEditProfile() {
  if (!profile.value) return
  profileForm.greenhouse_id = profile.value.greenhouse_id
  profileForm.trigger_sensor_device_id = undefined
  profileForm.code = profile.value.code
  profileForm.name = profile.value.name
  profileForm.description = profile.value.description || ''
  profileForm.trigger_metric_code = profile.value.trigger_metric_code
  profileForm.trigger_sensor_channel_id = profile.value.trigger_sensor_channel_id
  profileForm.enabled = profile.value.enabled
  profileDialogVisible.value = true

  // Load sensor devices for the profile's greenhouse
  await loadEditSensorDevices()
  if (profile.value.trigger_sensor_channel_id) {
    try {
      const ch = await deviceApi.getSensorChannel(profile.value.trigger_sensor_channel_id)
      profileForm.trigger_sensor_device_id = ch.sensor_device_id
      profileForm.trigger_metric_code = ch.metric_code
      await loadEditSensorChannelsForDevice(ch.sensor_device_id)
      profileForm.trigger_sensor_channel_id = ch.id
    } catch { /* ignore */ }
  }
}

async function loadEditSensorDevices() {
  const greenhouseId = profileForm.greenhouse_id
  if (!greenhouseId) {
    editSensorDevices.value = []
    return
  }
  try {
    const data = await deviceApi.getSensorDevices({ greenhouse_id: greenhouseId, page_size: LARGE_PAGE_SIZE })
    editSensorDevices.value = data.items
  } catch {
    editSensorDevices.value = []
  }
}

async function loadEditSensorChannelsForDevice(sensorDeviceId: number) {
  try {
    const data = await deviceApi.getSensorChannels({ sensor_device_id: sensorDeviceId, enabled: 1, page_size: LARGE_PAGE_SIZE })
    editSensorChannels.value = data.items
  } catch {
    editSensorChannels.value = []
  }
}

async function onEditProfileDeviceChange() {
  const sensorDeviceId = profileForm.trigger_sensor_device_id
  profileForm.trigger_sensor_channel_id = undefined
  editSensorChannels.value = []
  if (!sensorDeviceId) return
  await loadEditSensorChannelsForDevice(sensorDeviceId)
}

function onEditProfileMetricChange() {
  profileForm.trigger_sensor_channel_id = undefined
}

async function handleProfileSubmit() {
  if (!profileFormRef.value) return
  try {
    await profileFormRef.value.validate()
  } catch { return }
  if (!profileId.value) return

  profileSubmitLoading.value = true
  try {
    await climateApi.updateClimateProfile(profileId.value, {
      name: profileForm.name,
      description: profileForm.description || undefined,
      trigger_metric_code: profileForm.trigger_metric_code,
      trigger_sensor_channel_id: profileForm.trigger_sensor_channel_id!,
      enabled: profileForm.enabled
    })
    ElMessage.success('气候配置已更新')
    profileDialogVisible.value = false
    await loadProfile()
  } catch { /* handled by interceptor */ }
  finally { profileSubmitLoading.value = false }
}

async function handleDeleteProfile() {
  if (!profileId.value) return
  await ElMessageBox.confirm('确认删除该气候配置？', '提示', { type: 'warning' })
  await climateApi.deleteClimateProfile(profileId.value)
  ElMessage.success('已删除')
  router.push('/strategy/climate')
}

// ── Stage CRUD ──
const stageFormVisible = ref(false)
const isEditStage = ref(false)
const stageFormRef = ref<FormInstance>()
const stageSubmitLoading = ref(false)
const editingStageId = ref<number | null>(null)

const emptyStageForm = () => ({
  stage_level: 1,
  name: '',
  trigger_operator: '>' as string,
  trigger_threshold: 0,
  hysteresis: undefined as number | undefined
})

const stageForm = reactive(emptyStageForm())

const stageFormRules: FormRules = {
  stage_level: [{ required: true, message: '请输入级别', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  trigger_operator: [{ required: true, message: '请选择操作符', trigger: 'change' }],
  trigger_threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }]
}

function openCreateStage() {
  isEditStage.value = false
  editingStageId.value = null
  Object.assign(stageForm, emptyStageForm())
  stageFormVisible.value = true
}

function openEditStage(stage: ClimateStage) {
  isEditStage.value = true
  editingStageId.value = stage.id
  Object.assign(stageForm, {
    stage_level: stage.stage_level,
    name: stage.name,
    trigger_operator: stage.trigger_operator,
    trigger_threshold: stage.trigger_threshold,
    hysteresis: stage.hysteresis
  })
  stageFormVisible.value = true
}

async function handleStageSubmit() {
  if (!stageFormRef.value) return
  try {
    await stageFormRef.value.validate()
  } catch { return }
  if (!profileId.value) return

  stageSubmitLoading.value = true
  try {
    const payload = {
      stage_level: stageForm.stage_level,
      name: stageForm.name,
      trigger_operator: stageForm.trigger_operator,
      trigger_threshold: stageForm.trigger_threshold,
      hysteresis: stageForm.hysteresis
    }
    if (isEditStage.value && editingStageId.value) {
      await climateApi.updateClimateProfileStage(profileId.value, editingStageId.value, payload)
      ElMessage.success('阶段已更新')
    } else {
      await climateApi.createClimateProfileStage(profileId.value, payload)
      ElMessage.success('阶段已创建')
    }
    stageFormVisible.value = false
    await refreshStages()
    await loadProfile()
  } catch { /* handled by interceptor */ }
  finally { stageSubmitLoading.value = false }
}

async function removeStage(stageId: number) {
  if (!profileId.value) return
  await ElMessageBox.confirm('确认删除该阶段？', '提示', { type: 'warning' })
  await climateApi.deleteClimateProfileStage(profileId.value, stageId)
  ElMessage.success('已删除')
  delete actionsByStage[stageId]
  await refreshStages()
  await loadProfile()
}

// ── Action CRUD ──
const actionFormVisible = ref(false)
const isEditAction = ref(false)
const actionFormRef = ref<FormInstance>()
const actionSubmitLoading = ref(false)
const editingActionId = ref<number | null>(null)
const currentActionStageId = ref<number | null>(null)

const emptyActionForm = () => ({
  actuator_channel_id: undefined as number | undefined,
  command_type: 'SWITCH' as string,
  execution_order: 1,
  command_payload_str: '{}',
  enabled: true
})

const actionForm = reactive(emptyActionForm())

const actionFormRules: FormRules = {
  actuator_channel_id: [{ required: true, message: '请选择执行器通道', trigger: 'change' }],
  command_type: [{ required: true, message: '请选择命令类型', trigger: 'change' }],
  command_payload_str: [
    { required: true, message: '请输入命令参数', trigger: 'blur' },
    { validator: (_rule, value, cb) => { try { JSON.parse(value as string); cb() } catch { cb(new Error('JSON格式无效')) } }, trigger: 'blur' }
  ]
}

function openCreateAction(stage: ClimateStage) {
  isEditAction.value = false
  editingActionId.value = null
  currentActionStageId.value = stage.id
  Object.assign(actionForm, emptyActionForm())
  actionFormVisible.value = true
}

function openEditAction(stage: ClimateStage, action: ClimateStageAction) {
  isEditAction.value = true
  editingActionId.value = action.id
  currentActionStageId.value = stage.id
  Object.assign(actionForm, {
    actuator_channel_id: action.actuator_channel_id,
    command_type: action.command_type,
    execution_order: action.execution_order,
    command_payload_str: action.command_payload,
    enabled: action.enabled
  })
  actionFormVisible.value = true
}

async function handleActionSubmit() {
  if (!actionFormRef.value) return
  try {
    await actionFormRef.value.validate()
  } catch { return }
  if (!profileId.value || !currentActionStageId.value) return

  actionSubmitLoading.value = true
  try {
    let command_payload: Record<string, unknown>
    try {
      command_payload = JSON.parse(actionForm.command_payload_str)
    } catch {
      ElMessage.error('命令参数JSON格式无效')
      actionSubmitLoading.value = false
      return
    }

    const payload = {
      actuator_channel_id: actionForm.actuator_channel_id!,
      command_type: actionForm.command_type,
      command_payload,
      execution_order: actionForm.execution_order || undefined,
      enabled: actionForm.enabled
    }
    if (isEditAction.value && editingActionId.value) {
      await climateApi.updateClimateStageAction(profileId.value, currentActionStageId.value, editingActionId.value, payload)
      ElMessage.success('动作已更新')
    } else {
      await climateApi.createClimateStageAction(profileId.value, currentActionStageId.value, payload)
      ElMessage.success('动作已创建')
    }
    actionFormVisible.value = false
    await refreshActionsForStage(currentActionStageId.value)
    await loadProfile()
  } catch { /* handled by interceptor */ }
  finally { actionSubmitLoading.value = false }
}

async function removeAction(stageId: number, actionId: number) {
  if (!profileId.value) return
  await ElMessageBox.confirm('确认删除该动作？', '提示', { type: 'warning' })
  await climateApi.deleteClimateStageAction(profileId.value, stageId, actionId)
  ElMessage.success('已删除')
  await refreshActionsForStage(stageId)
  await loadProfile()
}

// ── Manual Execute Dialog ──
const executeDialogVisible = ref(false)
const executeLoading = ref(false)
const executeFormRef = ref<FormInstance>()

const emptyExecuteForm = () => ({
  trigger_value: 0,
  to_stage_level: 1,
  from_stage_level: undefined as number | undefined
})

const executeForm = reactive(emptyExecuteForm())

const executeFormRules: FormRules = {
  trigger_value: [{ required: true, message: '请输入触发值', trigger: 'blur' }],
  to_stage_level: [{ required: true, message: '请输入目标阶段级别', trigger: 'blur' }]
}

function openExecuteDialog() {
  Object.assign(executeForm, emptyExecuteForm())
  executeDialogVisible.value = true
}

async function handleExecute() {
  if (!executeFormRef.value) return
  try {
    await executeFormRef.value.validate()
  } catch { return }
  if (!profileId.value) return

  executeLoading.value = true
  try {
    const payload: { trigger_value: number; to_stage_level: number; from_stage_level?: number } = {
      trigger_value: executeForm.trigger_value,
      to_stage_level: executeForm.to_stage_level
    }
    if (executeForm.from_stage_level !== undefined && executeForm.from_stage_level > 0) {
      payload.from_stage_level = executeForm.from_stage_level
    }
    const result = await climateApi.executeClimateProfile(profileId.value, payload)
    ElMessage.success(`执行完成，共执行 ${result.executed_actions_count} 个动作`)
    executeDialogVisible.value = false
    // Refresh execution logs if on that tab
    if (activeTab.value === 'logs') {
      fetchExecutionLogs()
    }
  } catch { /* handled by interceptor */ }
  finally { executeLoading.value = false }
}

// ── Init ──
onMounted(() => {
  loadAllData()
  // Warm action lookups
  loadActuatorChannels()
})
</script>

<style scoped lang="scss">
.climate-profile-detail-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
      .page-title {
        font-size: 22px;
        font-weight: 700;
        color: var(--color-text-primary);
        margin: 0;
      }
    }
    .header-right {
      display: flex;
      gap: 8px;
    }
  }

  .loading-container {
    padding: 40px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }

  .info-card {
    margin-bottom: 16px;
  }

  .card-title {
    font-size: 15px;
    font-weight: 600;
  }

  .stats-row {
    margin-bottom: 16px;
    .stat-card {
      text-align: center;
      padding: 20px 16px;
      background: var(--bg-card);
      border-radius: var(--radius-md);
      box-shadow: var(--shadow-card);
      .stat-value {
        font-size: 28px;
        font-weight: 700;
        color: var(--color-primary);
      }
      .stat-label {
        font-size: 13px;
        color: var(--color-text-secondary);
        margin-top: 4px;
      }
    }
  }

  .tab-toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    .filter-separator {
      color: var(--color-text-secondary);
    }
  }

  .expanded-section {
    padding: 12px 24px;
    .expanded-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
      .section-subtitle {
        font-size: 14px;
        font-weight: 600;
      }
    }
  }

  .hint {
    color: var(--color-text-secondary);
    font-size: 12px;
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--spacing-md);
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }
}
</style>
