<template>
  <AppLayout>
    <div class="commands-page">
      <div class="page-header">
        <h1 class="page-title">控制命令</h1>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          下发命令
        </el-button>
      </div>

      <!-- 命令历史 -->
      <div class="table-container">
        <el-table :data="commands" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="device_name" label="目标设备" width="150">
            <template #default="{ row }">
              {{ getDeviceName(row) }}
            </template>
          </el-table-column>
          <el-table-column prop="command_type" label="类型" width="120">
            <template #default="{ row }">
              {{ getCommandTypeName(row.command_type) }}
            </template>
          </el-table-column>
          <el-table-column prop="payload" label="负载" min-width="180">
            <template #default="{ row }">
              <code class="payload-code">{{ JSON.stringify(row.payload) }}</code>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getCommandStatusType(row.status)">
                {{ getCommandStatusName(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="showDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
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

      <!-- 下发命令弹窗 -->
      <el-dialog v-model="createDialogVisible" title="下发命令" width="500px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
          <el-form-item label="目标设备" prop="device_id">
            <el-select v-model="formData.device_id" placeholder="请选择设备" filterable style="width: 100%">
              <el-option
                v-for="device in actuators"
                :key="device.id"
                :label="device.name"
                :value="device.id"
              />
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
            <el-input
              v-model="payloadStr"
              type="textarea"
              :rows="4"
              placeholder='请输入 JSON 格式负载，如 {"state": "ON"}'
            />
            <div v-if="payloadError" class="payload-error">{{ payloadError }}</div>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>

      <!-- 命令详情弹窗 -->
      <el-dialog v-model="detailDialogVisible" title="命令详情" width="500px">
        <el-descriptions v-if="currentCommand" :column="1" border>
          <el-descriptions-item label="命令ID">{{ currentCommand.id }}</el-descriptions-item>
          <el-descriptions-item label="目标设备">{{ getDeviceName(currentCommand) }}</el-descriptions-item>
          <el-descriptions-item label="命令类型">{{ getCommandTypeName(currentCommand.command_type) }}</el-descriptions-item>
          <el-descriptions-item label="命令负载">
            <code class="payload-code">{{ JSON.stringify(currentCommand.payload, null, 2) }}</code>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getCommandStatusType(currentCommand.status)">
              {{ getCommandStatusName(currentCommand.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentCommand.error_message" label="错误信息">
            <span class="error-text">{{ currentCommand.error_message }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDateTime(currentCommand.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentCommand.executed_at" label="执行时间">
            {{ formatDateTime(currentCommand.executed_at) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-dialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { AppLayout } from '@/components/layout'
import { controlApi } from '@/api'
import { useDeviceStore } from '@/stores/device'
import { formatDateTime, getCommandTypeName, getCommandStatusType, getCommandStatusName } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { CommandType, DeviceType } from '@/types'
import type { ControlCommand, CreateCommandParams, Device } from '@/types'

const deviceStore = useDeviceStore()

// 数据
const loading = ref(false)
const commands = ref<ControlCommand[]>([])
const total = ref(0)

// 执行器列表
const actuators = computed(() =>
  (deviceStore.devices as Device[]).filter(d => d.type === 'ACTUATOR')
)

const deviceNameMap = computed(() => {
  const map = new Map<number, string>()
  for (const device of actuators.value) {
    map.set(device.id, device.name)
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
  device_id: null as number | null,
  command_type: CommandType.SWITCH,
  payload: {}
})

const formRules: FormRules = {
  device_id: [{ required: true, message: '请选择设备', trigger: 'change' }],
  command_type: [{ required: true, message: '请选择命令类型', trigger: 'change' }]
}

// 详情弹窗
const detailDialogVisible = ref(false)
const currentCommand = ref<ControlCommand | null>(null)

// 监听命令类型变化，更新默认负载
watch(() => formData.command_type, (type) => {
  if (type === CommandType.SWITCH) {
    payloadStr.value = '{"state": "ON"}'
  } else if (type === CommandType.SET_VALUE) {
    payloadStr.value = '{"value": 50}'
  } else {
    payloadStr.value = '{}'
  }
})

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await controlApi.getCommands({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    commands.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 打开创建弹窗
function openCreateDialog() {
  formData.device_id = null
  formData.command_type = CommandType.SWITCH
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

  if (!formData.device_id) return

  submitLoading.value = true
  try {
    const params: CreateCommandParams = {
      device_id: formData.device_id,
      command_type: formData.command_type,
      payload: formData.payload
    }
    await controlApi.createCommand(params)
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
  detailDialogVisible.value = true
}

function getDeviceName(command: ControlCommand | null): string {
  if (!command) return '-'
  const commandWithFallback = command as ControlCommand & { target_device_name?: string }
  return commandWithFallback.device_name
    || commandWithFallback.target_device_name
    || deviceNameMap.value.get(commandWithFallback.device_id)
    || `设备${commandWithFallback.device_id}`
}

onMounted(() => {
  fetchData()
  // 加载设备列表
  deviceStore.fetchDevices({ type: DeviceType.ACTUATOR, page: 1, page_size: LARGE_PAGE_SIZE })
})
</script>

<style scoped lang="scss">
.commands-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    text-wrap: balance;
  }

  .table-container {
    background: #fff;
    border-radius: 4px;
    padding: 16px;
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .payload-code {
    background: #f5f7fa;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 12px;
  }

  .payload-error {
    color: #f56c6c;
    font-size: 12px;
    margin-top: 4px;
  }

  .error-text {
    color: #f56c6c;
  }
}
</style>
