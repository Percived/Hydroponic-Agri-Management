<template>
  <div class="climate-page">
    <div class="page-header">
      <h1 class="page-title">气候联动</h1>
      <el-button type="primary" @click="openCreateProfile">
        <el-icon><Plus /></el-icon>
        新增配置
      </el-button>
      <el-button type="success" @click="openCreateFull">
        <el-icon><Plus /></el-icon>
        高级创建
      </el-button>
    </div>

    <div class="filter-section">
      <el-select v-model="filters.greenhouse_id" placeholder="选择温室" clearable style="width: 200px">
        <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="profiles" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="greenhouse_id" label="温室ID" width="100" />
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="trigger_metric_code" label="触发指标" width="120" />
        <el-table-column label="阶段数" width="80">
          <template #default="{ row }">{{ row.stages_count ?? row.stages?.length ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditProfile(row)">编辑</el-button>
            <el-button type="success" link @click="openStagesDialog(row)">阶段</el-button>
            <el-button type="warning" link @click="openExecLogs(row)">日志</el-button>
            <el-button type="success" link @click="openExecuteDialog(row)">执行</el-button>
            <el-button type="danger" link @click="removeProfile(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </div>

    <!-- ── Profile 新增/编辑弹窗 ── -->
    <el-dialog v-model="profileDialogVisible" :title="isEditProfile ? '编辑气候配置' : '新增气候配置'" width="500px">
      <el-form ref="profileFormRef" :model="profileForm" :rules="profileFormRules" label-width="120px">
        <el-form-item label="温室" prop="greenhouse_id">
          <el-select v-model="profileForm.greenhouse_id" placeholder="选择温室" filterable style="width: 100%">
            <el-option v-for="gh in greenhouses" :key="gh.id" :label="`${gh.name} (ID:${gh.id})`" :value="gh.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="profileForm.code" maxlength="64" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="profileForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="profileForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="触发指标" prop="trigger_metric_code">
          <el-select v-model="profileForm.trigger_metric_code" placeholder="选择指标" filterable style="width: 100%">
            <el-option v-for="m in metrics" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
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

    <!-- ── Stages 弹窗 (Level 2) ── -->
    <el-dialog v-model="stagesDialogVisible" title="气候阶段" width="800px">
      <div style="margin-bottom: 12px">
        <el-button type="primary" size="small" @click="openCreateStage">新增阶段</el-button>
      </div>
      <el-table :data="stages" v-loading="stagesLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="stage_level" label="级别" width="70" />
        <el-table-column prop="name" label="名称" width="140" />
        <el-table-column label="触发条件" min-width="200">
          <template #default="{ row }">
            {{ row.trigger_operator }} {{ row.trigger_threshold }}
            <span v-if="row.hysteresis" class="hint">(回差: {{ row.hysteresis }})</span>
          </template>
        </el-table-column>
        <el-table-column label="动作数" width="80">
          <template #default="{ row }">{{ row.action_count ?? row.actions?.length ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditStage(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="openActionsDialog(row)">动作</el-button>
            <el-button type="danger" link size="small" @click="removeStage(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="stagesDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- ── Stage 新增/编辑弹窗 ── -->
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

    <!-- ── Actions 弹窗 (Level 3) ── -->
    <el-dialog v-model="actionsDialogVisible" title="阶段动作" width="800px">
      <div style="margin-bottom: 12px">
        <el-button type="primary" size="small" @click="openCreateAction">新增动作</el-button>
      </div>
      <el-table :data="actions" v-loading="actionsLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="actuator_channel_id" label="通道ID" width="100" />
        <el-table-column prop="command_type" label="命令类型" width="120" />
        <el-table-column label="命令参数" min-width="200">
          <template #default="{ row }">{{ JSON.stringify(row.command_payload) }}</template>
        </el-table-column>
        <el-table-column prop="execution_order" label="顺序" width="70" />
        <el-table-column label="启用" width="70">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditAction(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="removeAction(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="actionsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- ── Action 新增/编辑弹窗 ── -->
    <el-dialog v-model="actionFormVisible" :title="isEditAction ? '编辑动作' : '新增动作'" width="500px">
      <el-form ref="actionFormRef" :model="actionForm" :rules="actionFormRules" label-width="120px">
        <el-form-item label="执行器通道" prop="actuator_channel_id">
          <el-select v-model="actionForm.actuator_channel_id" placeholder="选择执行器通道" filterable style="width: 100%">
            <el-option v-for="ch in actuatorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.actuator_type})`" :value="ch.id" />
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
          <el-input-number v-model="actionForm.execution_order" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="actionForm.enabled" />
        </el-form-item>
        <el-form-item label="命令参数" prop="command_payload">
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

    <!-- ── 执行日志弹窗 ── -->
    <el-dialog v-model="logsDialogVisible" title="执行日志" width="800px">
      <el-table :data="execLogs" v-loading="logsLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="profile_id" label="配置ID" width="90" />
        <el-table-column prop="from_stage_level" label="从级别" width="90" />
        <el-table-column prop="to_stage_level" label="到级别" width="90" />
        <el-table-column prop="trigger_value" label="触发值" width="100" />
        <el-table-column prop="executed_actions_count" label="执行动作数" width="110" />
        <el-table-column prop="executed_at" label="执行时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.executed_at) }}</template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="logsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- ── 手动执行弹窗 ── -->
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

    <!-- ── 高级创建弹窗 (JSON) ── -->
    <el-dialog v-model="fullCreateVisible" title="高级创建 — 一键创建完整气候配置" width="700px">
      <el-alert type="info" :closable="false" style="margin-bottom: 12px">
        在一个请求中创建气候配置及其所有阶段和动作。请输入完整的 JSON：
      </el-alert>
      <el-input
        v-model="fullCreateJson"
        type="textarea"
        :rows="16"
        placeholder='{
  "greenhouse_id": 1,
  "code": "TEMP_CTRL",
  "name": "温度控制",
  "trigger_metric_code": "TEMP",
  "stages": [
    {
      "stage_level": 1,
      "name": "高温预警",
      "trigger_operator": ">",
      "trigger_threshold": 30,
      "hysteresis": 1.0,
      "actions": [
        {
          "actuator_channel_id": 1,
          "command_type": "SWITCH",
          "command_payload": {"value": 1},
          "execution_order": 1
        }
      ]
    }
  ]
}'
      />
      <template #footer>
        <el-button @click="fullCreateVisible = false">取消</el-button>
        <el-button type="primary" :loading="fullCreateLoading" @click="handleFullCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { climateApi, greenhouseApi, deviceApi, metricApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { ClimateProfile, ClimateStage, ClimateStageAction, ClimateExecutionLog, Greenhouse, ActuatorChannel, MetricDefinition } from '@/types'

// ── Profiles ──
const loading = ref(false)
const profiles = ref<ClimateProfile[]>([])
const total = ref(0)

const filters = reactive({
  greenhouse_id: undefined as number | undefined
})

const greenhouses = ref<Greenhouse[]>([])
const actuatorChannels = ref<ActuatorChannel[]>([])
const metrics = ref<MetricDefinition[]>([])

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch {
    greenhouses.value = []
  }
}

async function loadActuatorChannels() {
  try {
    const data = await deviceApi.getActuatorChannels({ page_size: LARGE_PAGE_SIZE })
    actuatorChannels.value = data.items
  } catch {
    actuatorChannels.value = []
  }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    metrics.value = data.items
  } catch {
    metrics.value = []
  }
}

const pagination = reactive({ page: 1, pageSize: 20 })

const profileDialogVisible = ref(false)
const isEditProfile = ref(false)
const profileFormRef = ref<FormInstance>()
const profileSubmitLoading = ref(false)
const editingProfileId = ref<number | null>(null)

const emptyProfileForm = () => ({
  greenhouse_id: undefined as number | undefined,
  code: '',
  name: '',
  description: '' as string,
  trigger_metric_code: '',
  enabled: true
})

const profileForm = reactive(emptyProfileForm())

const profileFormRules: FormRules = {
  greenhouse_id: [{ required: true, message: '请输入温室ID', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  trigger_metric_code: [{ required: true, message: '请输入触发指标', trigger: 'blur' }]
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.greenhouse_id) params.greenhouse_id = filters.greenhouse_id
    const data = await climateApi.getClimateProfiles(params)
    profiles.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.greenhouse_id = undefined
  pagination.page = 1
  fetchData()
}

function openCreateProfile() {
  isEditProfile.value = false
  editingProfileId.value = null
  Object.assign(profileForm, emptyProfileForm())
  profileDialogVisible.value = true
}

function openEditProfile(profile: ClimateProfile) {
  isEditProfile.value = true
  editingProfileId.value = profile.id
  Object.assign(profileForm, {
    greenhouse_id: profile.greenhouse_id,
    code: profile.code,
    name: profile.name,
    description: profile.description || '',
    trigger_metric_code: profile.trigger_metric_code,
    enabled: profile.enabled
  })
  profileDialogVisible.value = true
}

async function handleProfileSubmit() {
  if (!profileFormRef.value) return
  try {
    await profileFormRef.value.validate()
  } catch {
    return
  }

  profileSubmitLoading.value = true
  try {
    if (isEditProfile.value && editingProfileId.value) {
      await climateApi.updateClimateProfile(editingProfileId.value, {
        name: profileForm.name,
        description: profileForm.description || undefined,
        trigger_metric_code: profileForm.trigger_metric_code,
        enabled: profileForm.enabled
      })
      ElMessage.success('气候配置已更新')
    } else {
      await climateApi.createClimateProfile({
        greenhouse_id: profileForm.greenhouse_id!,
        code: profileForm.code,
        name: profileForm.name,
        description: profileForm.description || undefined,
        trigger_metric_code: profileForm.trigger_metric_code,
        enabled: profileForm.enabled
      })
      ElMessage.success('气候配置已创建')
    }
    profileDialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    profileSubmitLoading.value = false
  }
}

async function removeProfile(id: number) {
  await ElMessageBox.confirm('确认删除该气候配置？', '提示', { type: 'warning' })
  await climateApi.deleteClimateProfile(id)
  ElMessage.success('已删除')
  fetchData()
}

// ── Stages (Level 2) ──
const stagesDialogVisible = ref(false)
const stagesLoading = ref(false)
const stages = ref<ClimateStage[]>([])
const currentProfileId = ref<number | null>(null)

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

async function openStagesDialog(profile: ClimateProfile) {
  currentProfileId.value = profile.id
  stagesDialogVisible.value = true
  stagesLoading.value = true
  try {
    const data = await climateApi.getClimateProfileStages(profile.id)
    stages.value = data.items
  } catch {
    stages.value = []
  } finally {
    stagesLoading.value = false
  }
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
  } catch {
    return
  }
  if (!currentProfileId.value) return

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
      await climateApi.updateClimateProfileStage(currentProfileId.value, editingStageId.value, payload)
      ElMessage.success('阶段已更新')
    } else {
      await climateApi.createClimateProfileStage(currentProfileId.value, payload)
      ElMessage.success('阶段已创建')
    }
    stageFormVisible.value = false
    // Refresh
    stagesLoading.value = true
    try {
      const data = await climateApi.getClimateProfileStages(currentProfileId.value)
      stages.value = data.items
    } catch {
      stages.value = []
    } finally {
      stagesLoading.value = false
    }
  } catch {
    // handled by interceptor
  } finally {
    stageSubmitLoading.value = false
  }
}

async function removeStage(stageId: number) {
  if (!currentProfileId.value) return
  await ElMessageBox.confirm('确认删除该阶段？', '提示', { type: 'warning' })
  await climateApi.deleteClimateProfileStage(currentProfileId.value, stageId)
  ElMessage.success('已删除')
  stagesLoading.value = true
  try {
    const data = await climateApi.getClimateProfileStages(currentProfileId.value)
    stages.value = data.items
  } catch {
    stages.value = []
  } finally {
    stagesLoading.value = false
  }
}

// ── Actions (Level 3) ──
const actionsDialogVisible = ref(false)
const actionsLoading = ref(false)
const actions = ref<ClimateStageAction[]>([])
const currentStageId = ref<number | null>(null)

const actionFormVisible = ref(false)
const isEditAction = ref(false)
const actionFormRef = ref<FormInstance>()
const actionSubmitLoading = ref(false)
const editingActionId = ref<number | null>(null)

const emptyActionForm = () => ({
  actuator_channel_id: undefined as number | undefined,
  command_type: 'SWITCH' as string,
  execution_order: 1,
  command_payload_str: '{}',
  enabled: true
})

const actionForm = reactive(emptyActionForm())

const actionFormRules: FormRules = {
  actuator_channel_id: [{ required: true, message: '请输入通道ID', trigger: 'blur' }],
  command_type: [{ required: true, message: '请选择命令类型', trigger: 'change' }],
  command_payload_str: [
    { required: true, message: '请输入命令参数', trigger: 'blur' },
    { validator: (_rule, value, cb) => { try { JSON.parse(value as string); cb() } catch { cb(new Error('JSON格式无效')) } }, trigger: 'blur' }
  ]
}

async function openActionsDialog(stage: ClimateStage) {
  currentStageId.value = stage.id
  actionsDialogVisible.value = true
  actionsLoading.value = true
  try {
    const data = await climateApi.getClimateStageActions(currentProfileId.value!, stage.id)
    actions.value = data.items
  } catch {
    actions.value = []
  } finally {
    actionsLoading.value = false
  }
}

function openCreateAction() {
  isEditAction.value = false
  editingActionId.value = null
  Object.assign(actionForm, emptyActionForm())
  actionFormVisible.value = true
}

function openEditAction(action: ClimateStageAction) {
  isEditAction.value = true
  editingActionId.value = action.id
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
  } catch {
    return
  }
  if (!currentProfileId.value || !currentStageId.value) return

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
      await climateApi.updateClimateStageAction(
        currentProfileId.value, currentStageId.value, editingActionId.value, payload
      )
      ElMessage.success('动作已更新')
    } else {
      await climateApi.createClimateStageAction(
        currentProfileId.value, currentStageId.value, payload
      )
      ElMessage.success('动作已创建')
    }
    actionFormVisible.value = false
    // Refresh
    actionsLoading.value = true
    try {
      const data = await climateApi.getClimateStageActions(currentProfileId.value, currentStageId.value)
      actions.value = data.items
    } catch {
      actions.value = []
    } finally {
      actionsLoading.value = false
    }
  } catch {
    // handled by interceptor
  } finally {
    actionSubmitLoading.value = false
  }
}

async function removeAction(actionId: number) {
  if (!currentProfileId.value || !currentStageId.value) return
  await ElMessageBox.confirm('确认删除该动作？', '提示', { type: 'warning' })
  await climateApi.deleteClimateStageAction(currentProfileId.value, currentStageId.value, actionId)
  ElMessage.success('已删除')
  actionsLoading.value = true
  try {
    const data = await climateApi.getClimateStageActions(currentProfileId.value, currentStageId.value)
    actions.value = data.items
  } catch {
    actions.value = []
  } finally {
    actionsLoading.value = false
  }
}

// ── Execution Logs ──
const logsDialogVisible = ref(false)
const logsLoading = ref(false)
const execLogs = ref<ClimateExecutionLog[]>([])

async function openExecLogs(profile: ClimateProfile) {
  logsDialogVisible.value = true
  logsLoading.value = true
  try {
    const data = await climateApi.getClimateExecutionLogs({ profile_id: profile.id })
    execLogs.value = data.items
  } catch {
    execLogs.value = []
  } finally {
    logsLoading.value = false
  }
}

// ── Manual Execute ──
const executeDialogVisible = ref(false)
const executeLoading = ref(false)
const executeFormRef = ref<FormInstance>()
const executingProfileId = ref<number | null>(null)

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

function openExecuteDialog(profile: ClimateProfile) {
  executingProfileId.value = profile.id
  Object.assign(executeForm, emptyExecuteForm())
  executeDialogVisible.value = true
}

async function handleExecute() {
  if (!executeFormRef.value) return
  try {
    await executeFormRef.value.validate()
  } catch {
    return
  }
  if (!executingProfileId.value) return

  executeLoading.value = true
  try {
    const payload: { trigger_value: number; to_stage_level: number; from_stage_level?: number } = {
      trigger_value: executeForm.trigger_value,
      to_stage_level: executeForm.to_stage_level
    }
    if (executeForm.from_stage_level !== undefined && executeForm.from_stage_level > 0) {
      payload.from_stage_level = executeForm.from_stage_level
    }
    const result = await climateApi.executeClimateProfile(executingProfileId.value, payload)
    ElMessage.success(`执行完成，共执行 ${result.executed_actions_count} 个动作`)
    executeDialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    executeLoading.value = false
  }
}

// ── Full Create (JSON) ──
const fullCreateVisible = ref(false)
const fullCreateLoading = ref(false)
const fullCreateJson = ref('')

function openCreateFull() {
  fullCreateJson.value = ''
  fullCreateVisible.value = true
}

async function handleFullCreate() {
  if (!fullCreateJson.value.trim()) {
    ElMessage.error('请输入 JSON 配置')
    return
  }

  let parsed: Record<string, unknown>
  try {
    parsed = JSON.parse(fullCreateJson.value)
  } catch {
    ElMessage.error('JSON 格式无效')
    return
  }

  fullCreateLoading.value = true
  try {
    const result = await climateApi.createClimateProfileFull(parsed as never)
    ElMessage.success(`气候配置已创建 (ID: ${result.id})`)
    fullCreateVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    fullCreateLoading.value = false
  }
}

onMounted(() => {
  loadGreenhouses()
  loadActuatorChannels()
  loadMetrics()
  fetchData()
})
</script>

<style scoped lang="scss">
.climate-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  .page-title {
    font-size: 22px;
    font-weight: 700;
    color: var(--color-text-primary);
    margin: 0;
  }
  .filter-section {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
    margin-bottom: 16px;
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    align-items: center;
  }
  .table-container {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
  }
  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--spacing-md);
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }
  .hint {
    color: var(--color-text-secondary);
    font-size: 12px;
  }
}
</style>
