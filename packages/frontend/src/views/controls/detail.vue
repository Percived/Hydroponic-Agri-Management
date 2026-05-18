<template>
  <div class="policy-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/strategy/policies')" text>
          <el-icon><ArrowLeft /></el-icon>返回策略列表
        </el-button>
        <h1 class="page-title">{{ policy?.name || '策略详情' }}</h1>
        <el-tag v-if="policy" size="large">{{ typeLabel(policy.policy_type) }}</el-tag>
      </div>
      <div class="header-right" v-if="policy">
        <el-button type="primary" @click="openEditDialog">
          <el-icon><Edit /></el-icon>编辑
        </el-button>
        <el-button type="warning" :disabled="isPublished" :loading="publishLoading" @click="handlePublish">
          {{ isPublished ? '已发布' : '发布' }}
        </el-button>
      </div>
    </div>

    <template v-if="policy">
      <el-tabs v-model="activeTab">
        <!-- Tab 1: 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <el-descriptions :column="2" border class="info-descriptions">
            <el-descriptions-item label="策略编码">{{ policy.policy_code }}</el-descriptions-item>
            <el-descriptions-item label="策略名称">{{ policy.name }}</el-descriptions-item>
            <el-descriptions-item label="策略类型">
              <el-tag size="small">{{ typeLabel(policy.policy_type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="优先级">{{ policy.priority }}</el-descriptions-item>
            <el-descriptions-item label="重试次数">{{ policy.retry_limit }}</el-descriptions-item>
            <el-descriptions-item label="超时(秒)">{{ policy.timeout_sec }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="isEnabled ? 'success' : 'info'" size="small">{{ isEnabled ? '启用' : '停用' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="发布状态">
              <el-tag :type="isPublished ? 'success' : 'warning'" size="small">{{ isPublished ? '已发布' : '未发布' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="版本">{{ policy.version }}</el-descriptions-item>
            <el-descriptions-item label="生效时间">
              {{ policy.effective_from ? formatDateTime(policy.effective_from) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="失效时间">
              {{ policy.effective_to ? formatDateTime(policy.effective_to) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDateTime(policy.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatDateTime(policy.updated_at) }}</el-descriptions-item>
            <el-descriptions-item v-if="policy.published_at" label="发布时间">{{ formatDateTime(policy.published_at) }}</el-descriptions-item>
            <el-descriptions-item v-if="policy.policy_type === 'SCHEDULE'" label="计划描述" :span="2">
              {{ formatScheduleSummary(policy) }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- 触发条件 -->
          <div class="section-title">触发条件</div>
          <el-table :data="policy.conditions || []" stripe size="small" empty-text="无条件（仅定时执行）">
            <el-table-column prop="metric_code" label="指标" width="120">
              <template #default="{ row }">{{ getMetricName(row.metric_code) }}</template>
            </el-table-column>
            <el-table-column prop="operator" label="操作符" width="80" />
            <el-table-column prop="threshold_value" label="阈值" width="100" />
            <el-table-column prop="hysteresis" label="回差" width="80">
              <template #default="{ row }">{{ row.hysteresis ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="window_sec" label="窗口(秒)" width="100">
              <template #default="{ row }">{{ row.window_sec ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="aggregation" label="聚合方式" min-width="120">
              <template #default="{ row }">{{ row.aggregation || '-' }}</template>
            </el-table-column>
          </el-table>

          <!-- 执行目标 -->
          <div class="section-title">执行目标</div>
          <el-table :data="policy.targets || []" stripe size="small" empty-text="暂无执行目标">
            <el-table-column prop="actuator_channel_id" label="执行器通道" width="200">
              <template #default="{ row }">{{ actuatorChannelName(row.actuator_channel_id) }}</template>
            </el-table-column>
            <el-table-column prop="command_type" label="命令类型" width="120">
              <template #default="{ row }">{{ getCommandTypeName(row.command_type) }}</template>
            </el-table-column>
            <el-table-column prop="execution_order" label="执行顺序" width="100" />
            <el-table-column label="命令参数" min-width="200">
              <template #default="{ row }">{{ JSON.stringify(row.command_payload) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 执行历史 -->
        <el-tab-pane label="执行历史" name="executions">
          <div class="filter-section">
            <el-date-picker
              v-model="execFilters.from"
              type="datetime"
              placeholder="开始时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 190px"
            />
            <el-date-picker
              v-model="execFilters.to"
              type="datetime"
              placeholder="结束时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 190px"
            />
            <el-select v-model="execFilters.trigger_source" placeholder="触发来源" clearable style="width: 140px">
              <el-option label="遥测" value="TELEMETRY" />
              <el-option label="定时" value="SCHEDULE" />
              <el-option label="手动" value="MANUAL" />
            </el-select>
            <el-select v-model="execFilters.decision" placeholder="决策结果" clearable style="width: 140px">
              <el-option label="已执行" value="EXECUTED" />
              <el-option label="已跳过" value="SKIPPED" />
              <el-option label="失败" value="FAILED" />
              <el-option label="冲突" value="CONFLICT" />
            </el-select>
            <el-button type="primary" @click="fetchExecutions()">查询</el-button>
            <el-button @click="resetExecFilters()">重置</el-button>
          </div>

          <el-table :data="executions" v-loading="execLoading" stripe size="small">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="触发来源" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ triggerSourceLabel(row.trigger_source) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="触发指标" width="120">
              <template #default="{ row }">{{ row.trigger_metric_code || '-' }}</template>
            </el-table-column>
            <el-table-column prop="trigger_value" label="触发值" width="100">
              <template #default="{ row }">{{ row.trigger_value ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="决策" width="100">
              <template #default="{ row }">
                <el-tag :type="decisionTagType(row.decision)" size="small">{{ decisionLabel(row.decision) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="decision_reason" label="决策原因" min-width="180">
              <template #default="{ row }">{{ row.decision_reason || '-' }}</template>
            </el-table-column>
            <el-table-column label="执行时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.executed_at) }}</template>
            </el-table-column>
          </el-table>

          <div class="pagination-container">
            <el-pagination
              v-model:current-page="execPagination.page"
              v-model:page-size="execPagination.pageSize"
              :total="execTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchExecutions()"
              @current-change="fetchExecutions()"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else-if="!loading" description="策略不存在" />

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editDialogVisible" title="编辑策略" width="600px">
      <el-form ref="editFormRef" :model="editForm" label-width="120px">
        <el-form-item label="策略编码">
          <el-input :model-value="policy?.policy_code" disabled />
        </el-form-item>
        <el-form-item label="策略名称" required>
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-input-number v-model="editForm.priority" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="重试次数">
              <el-input-number v-model="editForm.retry_limit" :min="0" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="超时(秒)">
              <el-input-number v-model="editForm.timeout_sec" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用">
              <el-switch v-model="editForm.enabled" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="生效时间">
              <el-date-picker
                v-model="editForm.effective_from"
                type="datetime"
                placeholder="选择时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="失效时间">
              <el-date-picker
                v-model="editForm.effective_to"
                type="datetime"
                placeholder="选择时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <template v-if="policy?.policy_type === 'SCHEDULE'">
          <el-form-item label="计划模式">
            <el-radio-group v-model="editForm.schedule_mode">
              <el-radio value="ONCE">单次执行</el-radio>
              <el-radio value="DAILY">每日执行</el-radio>
              <el-radio value="WEEKLY">每周执行</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="editForm.schedule_mode === 'ONCE'" label="执行时间">
            <el-date-picker
              v-model="editForm.run_once_at"
              type="datetime"
              placeholder="选择执行时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item v-if="editForm.schedule_mode !== 'ONCE'" label="执行时刻">
            <el-time-picker
              v-model="editForm.time_of_day"
              placeholder="选择时刻"
              value-format="HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item v-if="editForm.schedule_mode === 'WEEKLY'" label="执行星期">
            <el-checkbox-group v-model="editForm.weekdays">
              <el-checkbox v-for="opt in weekdayOptions" :key="opt.value" :value="opt.value" :label="opt.label" />
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="时区">
            <el-select v-model="editForm.timezone" style="width: 100%">
              <el-option label="Asia/Shanghai" value="Asia/Shanghai" />
              <el-option label="Asia/Tokyo" value="Asia/Tokyo" />
            </el-select>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSubmitLoading" @click="handleEditSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Edit } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, FormInstance } from 'element-plus'
import { policyApi, deviceApi, metricApi } from '@/api'
import { formatDateTime, getMetricName, getCommandTypeName, populateMetricNames } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { actuatorChannelLabel, actuatorDeviceLabel, buildIdLabelMap, fallbackIdLabel } from '@/utils/labels'
import type { ControlPolicy, PolicyExecution, ActuatorChannel, ActuatorDevice } from '@/types'

const route = useRoute()
const router = useRouter()

const policyId = computed(() => Number(route.params.id))
const activeTab = ref('info')

// ── Policy detail ──
const loading = ref(false)
const policy = ref<ControlPolicy | null>(null)

const isEnabled = computed(() => {
  if (!policy.value) return false
  return policy.value.enabled === true || policy.value.enabled === 1
})

const isPublished = computed(() => {
  return Boolean(policy.value?.published_at)
})

function typeLabel(t: string) {
  const map: Record<string, string> = { THRESHOLD: '阈值', SCHEDULE: '定时', DURATION: '时长' }
  return map[t] || t
}

// ── Schedule summary (reused from rules.vue) ──
const weekdayOptions = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 7, label: '周日' }
]

function decodeWeekdaysMask(mask?: number): number[] {
  if (!mask) return []
  return weekdayOptions
    .map(item => item.value)
    .filter(day => (day === 7 ? (mask & (1 << 6)) !== 0 : (mask & (1 << (day - 1))) !== 0))
}

function formatScheduleSummary(p: ControlPolicy): string {
  if (p.policy_type !== 'SCHEDULE') return '-'
  if (!p.schedule_mode) return '计划未配置'
  if (p.schedule_mode === 'ONCE') {
    return p.run_once_at ? `${formatDateTime(p.run_once_at)} 单次` : '单次时间未配置'
  }
  if (p.schedule_mode === 'DAILY') {
    return p.time_of_day ? `每日 ${p.time_of_day}` : '每日时刻未配置'
  }
  if (p.schedule_mode === 'WEEKLY') {
    const weekdays = decodeWeekdaysMask(p.weekdays_mask)
      .map(day => weekdayOptions.find(item => item.value === day)?.label)
      .filter(Boolean)
      .join('/')
    return weekdays && p.time_of_day ? `${weekdays} ${p.time_of_day}` : '每周计划未配置'
  }
  return '计划未配置'
}

// ── Actuator channel labels ──
const actuatorChannels = ref<ActuatorChannel[]>([])
const actuatorDevices = ref<ActuatorDevice[]>([])

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

function actuatorChannelName(channelId: number) {
  return actuatorChannelLabelById.value[channelId] || fallbackIdLabel('通道', channelId)
}

async function loadActuatorDevices() {
  try {
    const data = await deviceApi.getActuatorDevices({ page_size: LARGE_PAGE_SIZE })
    actuatorDevices.value = data.items
  } catch { /* ignore */ }
}

async function loadActuatorChannels() {
  try {
    const data = await deviceApi.getActuatorChannels({ page_size: LARGE_PAGE_SIZE })
    actuatorChannels.value = data.items
  } catch { /* ignore */ }
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    populateMetricNames(data.items)
  } catch { /* ignore */ }
}

async function fetchPolicy() {
  loading.value = true
  try {
    policy.value = await policyApi.getPolicy(policyId.value)
  } catch {
    policy.value = null
  } finally {
    loading.value = false
  }
}

// ── Executions ──
const execLoading = ref(false)
const executions = ref<PolicyExecution[]>([])
const execTotal = ref(0)
const execPagination = reactive({ page: 1, pageSize: 20 })

const execFilters = reactive({
  from: '' as string,
  to: '' as string,
  trigger_source: '' as string,
  decision: '' as string
})

function triggerSourceLabel(s: string) {
  const map: Record<string, string> = { TELEMETRY: '遥测', SCHEDULE: '定时', MANUAL: '手动' }
  return map[s] || s
}

function decisionLabel(d: string) {
  const map: Record<string, string> = { EXECUTED: '已执行', SKIPPED: '已跳过', FAILED: '失败', CONFLICT: '冲突' }
  return map[d] || d
}

function decisionTagType(d: string) {
  const map: Record<string, string> = { EXECUTED: 'success', SKIPPED: 'info', FAILED: 'danger', CONFLICT: 'warning' }
  return map[d] || 'info'
}

async function fetchExecutions() {
  execLoading.value = true
  try {
    const params: Record<string, unknown> = {
      policy_id: policyId.value,
      page: execPagination.page,
      page_size: execPagination.pageSize
    }
    if (execFilters.from) params.from = execFilters.from
    if (execFilters.to) params.to = execFilters.to
    if (execFilters.trigger_source) params.trigger_source = execFilters.trigger_source
    if (execFilters.decision) params.decision = execFilters.decision
    const data = await policyApi.getPolicyExecutions(params)
    executions.value = data.items
    execTotal.value = data.total
  } catch {
    executions.value = []
    execTotal.value = 0
  } finally {
    execLoading.value = false
  }
}

function resetExecFilters() {
  execFilters.from = ''
  execFilters.to = ''
  execFilters.trigger_source = ''
  execFilters.decision = ''
  execPagination.page = 1
  fetchExecutions()
}

// ── Edit dialog ──
const editDialogVisible = ref(false)
const editSubmitLoading = ref(false)
const editFormRef = ref<FormInstance>()

const emptyEditForm = () => ({
  name: '',
  priority: 50,
  retry_limit: 3,
  timeout_sec: 30,
  enabled: true,
  effective_from: '' as string,
  effective_to: '' as string,
  schedule_mode: 'ONCE' as 'ONCE' | 'DAILY' | 'WEEKLY',
  run_once_at: '' as string,
  time_of_day: '' as string,
  weekdays: [] as number[],
  timezone: 'Asia/Shanghai'
})

const editForm = reactive(emptyEditForm())

function encodeWeekdaysMask(weekdays: number[]): number | undefined {
  if (!weekdays.length) return undefined
  return weekdays.reduce((mask, day) => {
    if (day === 7) return mask | (1 << 6)
    return mask | (1 << (day - 1))
  }, 0)
}

function openEditDialog() {
  if (!policy.value) return
  const p = policy.value
  Object.assign(editForm, {
    name: p.name,
    priority: p.priority,
    retry_limit: p.retry_limit,
    timeout_sec: p.timeout_sec,
    enabled: isEnabled.value,
    effective_from: p.effective_from || '',
    effective_to: p.effective_to || '',
    schedule_mode: p.schedule_mode || 'ONCE',
    run_once_at: p.run_once_at || '',
    time_of_day: p.time_of_day || '',
    weekdays: decodeWeekdaysMask(p.weekdays_mask),
    timezone: p.timezone || 'Asia/Shanghai'
  })
  editDialogVisible.value = true
}

async function handleEditSubmit() {
  editSubmitLoading.value = true
  try {
    const payload: Record<string, unknown> = {
      name: editForm.name,
      priority: editForm.priority,
      retry_limit: editForm.retry_limit,
      timeout_sec: editForm.timeout_sec,
      enabled: editForm.enabled,
      effective_from: editForm.effective_from || undefined,
      effective_to: editForm.effective_to || undefined
    }
    if (policy.value?.policy_type === 'SCHEDULE') {
      payload.schedule_mode = editForm.schedule_mode
      if (editForm.schedule_mode === 'ONCE') {
        payload.run_once_at = editForm.run_once_at || undefined
      } else {
        payload.time_of_day = editForm.time_of_day || undefined
      }
      if (editForm.schedule_mode === 'WEEKLY') {
        payload.weekdays_mask = encodeWeekdaysMask(editForm.weekdays)
      }
      payload.timezone = editForm.timezone
    }
    await policyApi.updatePolicy(policyId.value, payload)
    ElMessage.success('策略已更新，当前为未发布状态，请重新发布后生效')
    editDialogVisible.value = false
    await fetchPolicy()
  } catch {
    // handled by interceptor
  } finally {
    editSubmitLoading.value = false
  }
}

// ── Publish ──
const publishLoading = ref(false)

async function handlePublish() {
  if (!policy.value || isPublished.value) return
  try {
    await ElMessageBox.confirm('确认发布该策略？发布后将立即生效。', '提示', { type: 'warning' })
  } catch {
    return
  }
  publishLoading.value = true
  try {
    await policyApi.publishPolicy(policyId.value)
    ElMessage.success('策略已发布')
    await fetchPolicy()
  } catch {
    // handled by interceptor
  } finally {
    publishLoading.value = false
  }
}

// ── Tab switch → auto-load executions ──
let executionsLoaded = false
watch(activeTab, (tab) => {
  if (tab === 'executions' && !executionsLoaded) {
    executionsLoaded = true
    fetchExecutions()
  }
})

onMounted(() => {
  loadActuatorDevices()
  loadActuatorChannels()
  loadMetrics()
  fetchPolicy()
})
</script>

<style scoped lang="scss">
.policy-detail-page {
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
  }

  .info-descriptions {
    margin-bottom: 24px;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text-primary);
    margin: 20px 0 12px;
    padding-left: 8px;
    border-left: 3px solid var(--el-color-primary);
  }

  .filter-section {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    margin-bottom: 16px;
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    align-items: center;
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
