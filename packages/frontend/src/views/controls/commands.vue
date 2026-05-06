<template>
  <div class="commands-page">
    <div class="page-header">
      <h1 class="page-title">指令回执</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        下发命令
      </el-button>
    </div>

    <div class="filter-bar">
      <el-select v-model="statusFilter" clearable placeholder="状态筛选" style="width: 200px" @change="fetchData">
        <el-option label="PENDING" value="PENDING" />
        <el-option label="QUEUED" value="QUEUED" />
        <el-option label="SENT" value="SENT" />
        <el-option label="ACKED" value="ACKED" />
        <el-option label="TIMEOUT" value="TIMEOUT" />
        <el-option label="FAILED" value="FAILED" />
      </el-select>
    </div>

    <div class="table-container">
      <el-table :data="commands" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="执行器通道" width="150">
          <template #default="{ row }">{{ getChannelName(row.actuator_channel_id) }}</template>
        </el-table-column>
        <el-table-column prop="command_type" label="类型" width="110">
          <template #default="{ row }">{{ getCommandTypeName(row.command_type) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="getCommandStatusType(row.status)">
              {{ getCommandStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showDetail(row)">回执</el-button>
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

    <el-dialog v-model="createDialogVisible" title="下发命令" width="500px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="120px">
        <el-form-item label="执行器通道" prop="actuator_channel_id">
          <el-select v-model="formData.actuator_channel_id" placeholder="请选择执行器通道" filterable style="width: 100%">
            <el-option v-for="ch in actuatorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.actuator_type})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令类型" prop="command_type">
          <el-select v-model="formData.command_type" placeholder="请选择命令类型" style="width: 100%">
            <el-option label="开关" value="SWITCH" />
            <el-option label="设置值" value="SET_VALUE" />
            <el-option label="校准" value="CALIBRATE" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令负载" prop="payload">
          <el-input v-model="payloadStr" type="textarea" :rows="4" placeholder='请输入 JSON 格式负载，如 {"state":"ON"}' />
          <div v-if="payloadError" class="payload-error">{{ payloadError }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="命令回执详情" width="700px">
      <el-descriptions v-if="currentCommand" :column="2" border>
        <el-descriptions-item label="命令ID">{{ currentCommand.id }}</el-descriptions-item>
        <el-descriptions-item label="执行器通道">{{ getChannelName(currentCommand.actuator_channel_id) }}</el-descriptions-item>
        <el-descriptions-item label="命令类型">{{ getCommandTypeName(currentCommand.command_type) }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">{{ getCommandStatusName(currentCommand.status) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(currentCommand.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="发送时间">{{ formatDateTime(currentCommand.sent_at) }}</el-descriptions-item>
        <el-descriptions-item label="确认时间">{{ formatDateTime(currentCommand.acked_at) }}</el-descriptions-item>
        <el-descriptions-item label="请求ID">{{ currentCommand.request_id || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-alert
        v-if="receiptSummary"
        :title="receiptSummary"
        type="info"
        show-icon
        class="receipt-summary"
      />

      <div v-if="receipts.length > 0" class="receipt-list">
        <h4>回执列表</h4>
        <el-table :data="receipts" size="small" stripe>
          <el-table-column prop="receipt_seq" label="序号" width="70" />
          <el-table-column prop="receipt_status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getCommandStatusType(row.receipt_status)">{{ row.receipt_status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ack_code" label="确认码" width="100" />
          <el-table-column prop="ack_message" label="消息" min-width="160" />
          <el-table-column prop="ack_at" label="确认时间" width="180">
            <template #default="{ row }">{{ formatDateTime(row.ack_at) }}</template>
          </el-table-column>
        </el-table>
      </div>
      <div v-else class="empty-receipts">
        <el-empty description="暂无回执数据" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { commandApi, deviceApi } from '@/api'
import { formatDateTime, getCommandTypeName, getCommandStatusType, getCommandStatusName } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { ControlCommand, ControlCommandReceipt, ActuatorChannel } from '@/types'

const loading = ref(false)
const commands = ref<ControlCommand[]>([])
const total = ref(0)
const statusFilter = ref<string>()

const actuatorChannels = ref<ActuatorChannel[]>([])

const channelNameMap = computed(() => {
  const map = new Map<number, string>()
  for (const ch of actuatorChannels.value) {
    map.set(ch.id, `${ch.channel_code} (${ch.actuator_type})`)
  }
  return map
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 创建弹窗
const createDialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const payloadStr = ref('{"state": "ON"}')
const payloadError = ref('')

const formData = reactive({
  actuator_channel_id: null as number | null,
  command_type: 'SWITCH',
  payload: {} as Record<string, unknown>
})

const formRules: FormRules = {
  actuator_channel_id: [{ required: true, message: '请选择执行器通道', trigger: 'change' }],
  command_type: [{ required: true, message: '请选择命令类型', trigger: 'change' }]
}

// 详情弹窗
const detailDialogVisible = ref(false)
const currentCommand = ref<ControlCommand | null>(null)
const receipts = ref<ControlCommandReceipt[]>([])
const receiptSummary = ref('')

// 监听命令类型变化，更新默认负载
watch(() => formData.command_type, (type) => {
  if (type === 'SWITCH') {
    payloadStr.value = '{"state": "ON"}'
  } else if (type === 'SET_VALUE') {
    payloadStr.value = '{"value": 50}'
  } else {
    payloadStr.value = '{}'
  }
})

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    const data = await commandApi.getCommands(params)
    commands.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 加载执行器通道
async function loadActuatorChannels() {
  try {
    const data = await deviceApi.getActuatorChannels({ page_size: LARGE_PAGE_SIZE })
    actuatorChannels.value = data.items
  } catch {
    // ignore
  }
}

// 打开创建弹窗
function openCreateDialog() {
  formData.actuator_channel_id = null
  formData.command_type = 'SWITCH'
  payloadStr.value = '{"state": "ON"}'
  payloadError.value = ''
  createDialogVisible.value = true
}

// 提交命令
async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  // 验证 JSON 格式
  try {
    formData.payload = JSON.parse(payloadStr.value)
    payloadError.value = ''
  } catch {
    payloadError.value = 'JSON 格式错误'
    return
  }

  if (!formData.actuator_channel_id) return

  submitLoading.value = true
  try {
    await commandApi.createCommand({
      actuator_channel_id: formData.actuator_channel_id,
      command_type: formData.command_type,
      payload: formData.payload
    })
    ElMessage.success('命令下发成功')
    createDialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

// 显示详情
function showDetail(command: ControlCommand) {
  currentCommand.value = command
  loadReceipts(command.id)
  detailDialogVisible.value = true
}

async function loadReceipts(commandId: number) {
  try {
    const result = await commandApi.getCommandReceipts(commandId)
    receipts.value = (result.items || []).sort((a, b) => a.receipt_seq - b.receipt_seq)
    receiptSummary.value = computeReceiptSummary(receipts.value)
  } catch {
    receipts.value = []
    receiptSummary.value = ''
  }
}

function computeReceiptSummary(items: ControlCommandReceipt[]): string {
  if (items.length === 0) return '暂无回执数据'
  const first = items[0]
  const last = items[items.length - 1]
  const start = first.ack_at ? new Date(first.ack_at).getTime() : new Date(first.created_at).getTime()
  const end = last.ack_at ? new Date(last.ack_at).getTime() : new Date(last.created_at).getTime()
  const latency = Number.isFinite(start) && Number.isFinite(end) ? `${Math.max(0, end - start)}ms` : '-'
  const fail = items.find((it) => it.receipt_status === 'FAILED' || it.receipt_status === 'TIMEOUT')
  const failMessage = fail?.ack_message || fail?.ack_code || '-'
  return `最终状态: ${last.receipt_status}, 链路延迟: ${latency}, 失败原因: ${failMessage}`
}

function getChannelName(channelId: number): string {
  return channelNameMap.value.get(channelId) || `通道#${channelId}`
}

onMounted(() => {
  fetchData()
  loadActuatorChannels()
})
</script>

<style scoped lang="scss">
.commands-page {
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
    text-wrap: balance;
  }

  .table-container {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
  }

  .filter-bar {
    margin-bottom: 12px;
  }

  .receipt-summary {
    margin: 12px 0;
  }

  .receipt-list {
    margin-top: 16px;
    h4 {
      margin: 0 0 8px;
      font-size: 14px;
      font-weight: 600;
    }
  }

  .empty-receipts {
    margin-top: 16px;
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--spacing-md);
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }

  .payload-code {
    background: var(--color-primary-bg-light);
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    border: 1px solid var(--border-color-light);
  }

  .payload-error {
    color: var(--color-danger);
    font-size: 12px;
    margin-top: 4px;
  }

  .error-text {
    color: var(--color-danger);
  }
}
</style>
