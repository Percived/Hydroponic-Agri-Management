<template>
  <div class="rules-page">
    <div class="page-header">
      <h1 class="page-title">策略管理</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增策略
      </el-button>
    </div>

    <div class="table-container">
      <el-table :data="policies" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="policy_code" label="策略编码" width="150" />
        <el-table-column prop="name" label="策略名称" min-width="160" />
        <el-table-column prop="policy_type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.policy_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="90" />
        <el-table-column prop="retry_limit" label="重试" width="80" />
        <el-table-column prop="timeout_sec" label="超时(s)" width="90" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">{{ row.enabled === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="warning" link @click="handlePublish(row)">发布</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="760px">
      <el-form :model="formData" label-width="120px">
        <el-form-item label="策略编码">
          <el-input v-model="formData.policy_code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="策略名称">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="所属温室">
              <el-select v-model="formData.greenhouse_id" placeholder="请选择温室" filterable style="width: 100%" @change="onGreenhouseChange">
                <el-option v-for="gh in greenhouses" :key="gh.id" :label="gh.name" :value="gh.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="种植区">
              <el-select v-model="formData.growing_zone_id" placeholder="可选" clearable style="width: 100%" @change="onGrowingZoneChange">
                <el-option v-for="zone in growingZones" :key="zone.id" :label="zone.name" :value="zone.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-input-number v-model="formData.priority" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="策略类型">
              <el-select v-model="formData.policy_type" style="width: 100%">
                <el-option label="阈值 THRESHOLD" value="THRESHOLD" />
                <el-option label="定时 SCHEDULE" value="SCHEDULE" />
                <el-option label="时长 DURATION" value="DURATION" :disabled="true" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="重试次数">
              <el-input-number v-model="formData.retry_limit" :min="0" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="超时(秒)">
              <el-input-number v-model="formData.timeout_sec" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="生效起始">
              <el-date-picker
                v-model="formData.effective_from"
                type="datetime"
                placeholder="可选"
                clearable
                style="width: 100%"
                format="YYYY-MM-DD HH:mm"
                :disabled-date="disabledPastDate"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="失效时间">
              <el-date-picker
                v-model="formData.effective_to"
                type="datetime"
                placeholder="可选"
                clearable
                style="width: 100%"
                format="YYYY-MM-DD HH:mm"
                :disabled-date="disabledInvalidToDate"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- DURATION 未实现提示 -->
        <el-alert
          v-if="formData.policy_type === 'DURATION'"
          title="时长策略后端尚未实现，请选择阈值或定时策略"
          type="warning"
          :closable="false"
          style="margin-bottom: 16px"
        />

        <!-- THRESHOLD + SCHEDULE: 策略条件 -->
        <template v-if="formData.policy_type !== 'DURATION'">
          <el-divider>
            策略条件
            <el-switch
              v-if="formData.policy_type === 'SCHEDULE'"
              v-model="scheduleUseCondition"
              size="small"
              style="margin-left: 8px"
            />
            <span v-if="formData.policy_type === 'SCHEDULE'" style="font-size:12px;color:#909399;margin-left:4px">
              {{ scheduleUseCondition ? '启用条件检查' : '无条件（仅定时执行）' }}
            </span>
          </el-divider>

          <template v-if="formData.policy_type === 'THRESHOLD' || scheduleUseCondition">
            <el-form-item label="指标代码">
              <el-select v-model="conditionForm.metric_code" placeholder="选择指标" style="width: 100%">
                <el-option v-for="m in metrics" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
              </el-select>
            </el-form-item>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="运算符">
                  <el-select v-model="conditionForm.operator" style="width: 100%">
                    <el-option label="大于 >" value=">" />
                    <el-option label="大于等于 >=" value=">=" />
                    <el-option label="小于 <" value="<" />
                    <el-option label="小于等于 <=" value="<=" />
                    <el-option label="等于 =" value="=" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="阈值">
                  <el-input-number v-model="conditionForm.threshold_value" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :span="6">
                <el-form-item label="滞后值">
                  <el-input-number v-model="conditionForm.hysteresis" :min="0" :precision="2" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="窗口(秒)">
                  <el-input-number v-model="conditionForm.window_sec" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="持续(秒)">
                  <el-input-number v-model="conditionForm.required_duration_sec" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="聚合方式">
                  <el-select v-model="conditionForm.aggregation" placeholder="默认 last" clearable style="width: 100%">
                    <el-option label="最新 last" value="last" />
                    <el-option label="平均 avg" value="avg" />
                    <el-option label="最大 max" value="max" />
                    <el-option label="最小 min" value="min" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
          </template>
        </template>

        <el-divider>目标动作</el-divider>
        <el-form-item label="执行器通道">
          <el-select v-model="targetChannelId" placeholder="选择执行器通道" filterable style="width: 100%">
            <el-option v-for="ch in actuatorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.actuator_type})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="命令类型">
              <el-select v-model="targetCommandType" style="width: 100%">
                <el-option label="开关" value="SWITCH" />
                <el-option label="设置值" value="SET_VALUE" />
                <el-option label="校准" value="CALIBRATE" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="执行顺序">
              <el-input-number v-model="targetExecutionOrder" :min="0" :max="99" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="命令负载">
          <el-input v-model="targetPayloadRaw" type="textarea" :rows="3" placeholder='{"state":"ON"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { deviceApi, policyApi, greenhouseApi, metricApi } from '@/api'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { ControlPolicy, ActuatorChannel, Greenhouse, GrowingZone, MetricDefinition } from '@/types'
import { populateMetricNames } from '@/utils/format'

const loading = ref(false)
const submitLoading = ref(false)
const policies = ref<ControlPolicy[]>([])
const total = ref(0)
const pagination = reactive({ page: 1, pageSize: 20 })
const actuatorChannels = ref<ActuatorChannel[]>([])
const greenhouses = ref<Greenhouse[]>([])
const growingZones = ref<GrowingZone[]>([])
const metrics = ref<MetricDefinition[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingPolicyId = ref<number | null>(null)

const formData = reactive({
  policy_code: '',
  name: '',
  policy_type: 'THRESHOLD' as string,
  greenhouse_id: null as number | null,
  growing_zone_id: undefined as number | undefined,
  priority: 50,
  retry_limit: 3,
  timeout_sec: 30,
  effective_from: null as Date | null,
  effective_to: null as Date | null
})

const conditionForm = reactive({
  metric_code: 'TEMP',
  operator: '>',
  threshold_value: 30,
  hysteresis: undefined as number | undefined,
  window_sec: undefined as number | undefined,
  required_duration_sec: undefined as number | undefined,
  aggregation: undefined as string | undefined
})

const targetChannelId = ref<number | undefined>()
const targetCommandType = ref('SWITCH')
const targetPayloadRaw = ref('{"state":"ON"}')
const targetExecutionOrder = ref(0)

// SCHEDULE 类型: 是否启用条件检查
const scheduleUseCondition = ref(false)

// 动态标题
const dialogTitle = computed(() => {
  const tLabel = typeLabel(formData.policy_type)
  return isEdit.value ? `编辑策略 - ${tLabel}` : `新增策略 - ${tLabel}`
})

function typeLabel(t: string) {
  const map: Record<string, string> = { THRESHOLD: '阈值', SCHEDULE: '定时', DURATION: '时长' }
  return map[t] || t
}


async function fetchData() {
  loading.value = true
  try {
    const data = await policyApi.getPolicies({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    policies.value = data.items
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function loadActuatorChannels(greenhouseId?: number, growingZoneId?: number) {
  try {
    const params: Record<string, unknown> = { page_size: LARGE_PAGE_SIZE }
    if (greenhouseId) params.greenhouse_id = greenhouseId
    if (growingZoneId) params.growing_zone_id = growingZoneId
    const data = await deviceApi.getActuatorChannels(params)
    actuatorChannels.value = data.items
  } catch {
    actuatorChannels.value = []
  }
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { /* ignore */ }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    metrics.value = data.items
    populateMetricNames(data.items)
  } catch { /* ignore */ }
}

async function loadGrowingZones(greenhouseId?: number) {
  if (!greenhouseId) return
  try {
    const data = await greenhouseApi.getGrowingZones({ greenhouse_id: greenhouseId, page_size: LARGE_PAGE_SIZE })
    growingZones.value = data.items
  } catch {
    growingZones.value = []
  }
}

function disabledPastDate(date: Date): boolean {
  return date.getTime() < Date.now() - 60 * 1000
}

function disabledInvalidToDate(date: Date): boolean {
  if (date.getTime() < Date.now() - 60 * 1000) return true
  if (formData.effective_from) {
    return date.getTime() <= formData.effective_from.getTime()
  }
  return false
}

function onGreenhouseChange(greenhouseId: number | null) {
  formData.growing_zone_id = undefined
  growingZones.value = []
  if (greenhouseId) {
    loadGrowingZones(greenhouseId)
    loadActuatorChannels(greenhouseId)
  } else {
    actuatorChannels.value = []
  }
}

function onGrowingZoneChange(growingZoneId: number | undefined) {
  if (growingZoneId && formData.greenhouse_id) {
    loadActuatorChannels(formData.greenhouse_id, growingZoneId)
  } else if (formData.greenhouse_id) {
    loadActuatorChannels(formData.greenhouse_id)
  } else {
    actuatorChannels.value = []
  }
}

function openCreateDialog() {
  isEdit.value = false
  editingPolicyId.value = null
  formData.policy_code = `POL-${Date.now().toString().slice(-6)}`
  formData.name = ''
  formData.policy_type = 'THRESHOLD'
  formData.greenhouse_id = null
  formData.growing_zone_id = undefined
  formData.priority = 50
  formData.retry_limit = 3
  formData.timeout_sec = 30
  formData.effective_from = null
  formData.effective_to = null
  conditionForm.metric_code = 'TEMP'
  conditionForm.operator = '>'
  conditionForm.threshold_value = 30
  conditionForm.hysteresis = undefined
  conditionForm.window_sec = undefined
  conditionForm.required_duration_sec = undefined
  conditionForm.aggregation = undefined
  targetChannelId.value = undefined
  targetCommandType.value = 'SWITCH'
  targetPayloadRaw.value = '{"state":"ON"}'
  targetExecutionOrder.value = 0
  scheduleUseCondition.value = false
  growingZones.value = []
  actuatorChannels.value = []
  dialogVisible.value = true
}

// THRESHOLD 必须使用条件，SCHEDULE 切换时重置
watch(() => formData.policy_type, (val) => {
  if (val === 'THRESHOLD') scheduleUseCondition.value = false
  // THRESHOLD 的 scheduleUseCondition 不生效，条件始终显示
})

async function openEditDialog(policy: ControlPolicy) {
  isEdit.value = true
  editingPolicyId.value = policy.id
  formData.policy_code = policy.policy_code
  formData.name = policy.name
  formData.policy_type = policy.policy_type
  formData.greenhouse_id = policy.greenhouse_id
  formData.growing_zone_id = policy.growing_zone_id
  formData.priority = policy.priority || 50
  formData.retry_limit = policy.retry_limit || 3
  formData.timeout_sec = policy.timeout_sec || 30
  formData.effective_from = policy.effective_from ? new Date(policy.effective_from) : null
  formData.effective_to = policy.effective_to ? new Date(policy.effective_to) : null

  // Load conditions
  scheduleUseCondition.value = false
  try {
    const condResult = await policyApi.getPolicyConditions(policy.id)
    if (condResult.items && condResult.items.length > 0) {
      const c = condResult.items[0]
      conditionForm.metric_code = c.metric_code
      conditionForm.operator = c.operator
      conditionForm.threshold_value = c.threshold_value
      conditionForm.hysteresis = c.hysteresis
      conditionForm.window_sec = c.window_sec
      conditionForm.required_duration_sec = c.required_duration_sec
      conditionForm.aggregation = c.aggregation
      if (policy.policy_type === 'SCHEDULE') scheduleUseCondition.value = true
    }
  } catch { /* ignore */ }

  // Load targets
  try {
    const tgtResult = await policyApi.getPolicyTargets(policy.id)
    if (tgtResult.items && tgtResult.items.length > 0) {
      const t = tgtResult.items[0]
      targetChannelId.value = t.actuator_channel_id
      targetCommandType.value = t.command_type
      targetPayloadRaw.value = JSON.stringify(t.command_payload)
      targetExecutionOrder.value = t.execution_order
    }
  } catch { /* ignore */ }

  if (policy.greenhouse_id) {
    loadGrowingZones(policy.greenhouse_id)
    loadActuatorChannels(policy.greenhouse_id, policy.growing_zone_id)
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const now = Date.now()

  if (formData.effective_from && formData.effective_from.getTime() < now) {
    ElMessage.warning('生效起始时间不能早于当前时间')
    return
  }
  if (formData.effective_to) {
    if (formData.effective_to.getTime() < now) {
      ElMessage.warning('失效时间不能早于当前时间')
      return
    }
    if (formData.effective_from && formData.effective_to.getTime() <= formData.effective_from.getTime()) {
      ElMessage.warning('失效时间必须晚于生效起始时间')
      return
    }
  }

  submitLoading.value = true
  try {
    if (isEdit.value && editingPolicyId.value) {
      const payload: any = {
        name: formData.name,
        policy_type: formData.policy_type,
        growing_zone_id: formData.growing_zone_id,
        priority: formData.priority,
        retry_limit: formData.retry_limit,
        timeout_sec: formData.timeout_sec,
        effective_from: formData.effective_from?.toISOString() || null,
        effective_to: formData.effective_to?.toISOString() || null
      }
      // Remove null fields to avoid overwriting with null
      if (payload.effective_from === null) delete payload.effective_from
      if (payload.effective_to === null) delete payload.effective_to
      await policyApi.updatePolicy(editingPolicyId.value, payload as any)
      ElMessage.success('策略更新成功')
    } else {
      // Create new policy
      const useCond = formData.policy_type === 'THRESHOLD' || scheduleUseCondition.value
      const createPayload = {
        policy_code: formData.policy_code,
        name: formData.name,
        policy_type: formData.policy_type,
        greenhouse_id: formData.greenhouse_id!,
        growing_zone_id: formData.growing_zone_id,
        priority: formData.priority,
        retry_limit: formData.retry_limit,
        timeout_sec: formData.timeout_sec,
        effective_from: formData.effective_from?.toISOString() || undefined,
        effective_to: formData.effective_to?.toISOString() || undefined,
        conditions: useCond ? [{
          metric_code: conditionForm.metric_code,
          operator: conditionForm.operator,
          threshold_value: conditionForm.threshold_value,
          hysteresis: conditionForm.hysteresis,
          window_sec: conditionForm.window_sec,
          required_duration_sec: conditionForm.required_duration_sec,
          aggregation: conditionForm.aggregation
        }] : [],
        targets: targetChannelId.value ? [{
          actuator_channel_id: targetChannelId.value,
          command_type: targetCommandType.value,
          command_payload: (() => {
            try { return JSON.parse(targetPayloadRaw.value || '{}') }
            catch { return {} }
          })(),
          execution_order: targetExecutionOrder.value
        }] : []
      }
      await policyApi.createPolicy(createPayload as any)
      ElMessage.success('策略创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitLoading.value = false
  }
}

async function handlePublish(policy: ControlPolicy) {
  try {
    await policyApi.publishPolicy(policy.id)
    ElMessage.success('策略已发布')
    fetchData()
  } catch {
    // error handled
  }
}

async function handleDelete(policy: ControlPolicy) {
  await ElMessageBox.confirm(`确认删除策略「${policy.name}」？`, '提示', { type: 'warning' })
  await policyApi.deletePolicy(policy.id)
  ElMessage.success('策略已删除')
  fetchData()
}

onMounted(() => {
  fetchData()
  loadActuatorChannels()
  loadGreenhouses()
  loadMetrics()
})
</script>

<style scoped lang="scss">
.rules-page {
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
}
</style>
