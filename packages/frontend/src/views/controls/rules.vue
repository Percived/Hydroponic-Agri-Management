<template>
    <div class="rules-page">
      <div class="page-header">
        <h1 class="page-title">控制规则</h1>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增规则
        </el-button>
      </div>

      <!-- 规则列表 -->
      <div class="table-container">
        <el-table :data="rules" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="规则名称" width="150" />
          <el-table-column prop="metric_code" label="监控指标" width="100">
            <template #default="{ row }">
              {{ getMetricName(row.metric_code) }}
            </template>
          </el-table-column>
          <el-table-column label="条件" width="150">
            <template #default="{ row }">
              {{ row.metric_code }} {{ row.operator }} {{ row.threshold }}
            </template>
          </el-table-column>
          <el-table-column prop="target_device_name" label="目标设备" width="120" />
          <el-table-column label="执行动作" min-width="180">
            <template #default="{ row }">
              {{ getCommandTypeName(row.command_type) }}:
              <code class="payload-code">{{ JSON.stringify(row.command_payload) }}</code>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="100">
            <template #default="{ row }">
              <el-switch
                :model-value="row.enabled"
                @change="(val: boolean) => toggleRuleStatus(row, val)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
              <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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

      <!-- 新增/编辑规则弹窗 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规则' : '新增规则'" width="600px">
        <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
          <el-form-item label="规则名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入规则名称" maxlength="64" />
          </el-form-item>
          <el-form-item label="监控指标" prop="metric_code">
            <el-select v-model="formData.metric_code" placeholder="请选择监控指标" style="width: 100%">
              <el-option label="温度" value="TEMP" />
              <el-option label="湿度" value="HUMIDITY" />
              <el-option label="pH值" value="PH" />
              <el-option label="电导率" value="EC" />
              <el-option label="CO2" value="CO2" />
              <el-option label="光照" value="LIGHT" />
            </el-select>
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="比较运算" prop="operator">
                <el-select v-model="formData.operator" placeholder="请选择" style="width: 100%">
                  <el-option label="大于 (>)" value=">" />
                  <el-option label="大于等于 (>=)" value=">=" />
                  <el-option label="小于 (<)" value="<" />
                  <el-option label="小于等于 (<=)" value="<=" />
                  <el-option label="等于 (==)" value="==" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="阈值" prop="threshold">
                <el-input-number v-model="formData.threshold" :precision="2" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="目标设备" prop="target_device_id">
            <el-select
              v-model="formData.target_device_id"
              :placeholder="actuatorsLoading ? '正在加载执行器...' : '请选择执行器'"
              :disabled="actuatorsLoading"
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="device in actuators"
                :key="device.id"
                :label="device.name"
                :value="device.id"
              />
            </el-select>
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="命令类型" prop="command_type">
                <el-select v-model="formData.command_type" placeholder="请选择" style="width: 100%">
                  <el-option label="开关" value="SWITCH" />
                  <el-option label="设置值" value="SET_VALUE" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="是否启用" prop="enabled">
                <el-switch v-model="formData.enabled" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="执行动作" prop="command_payload">
            <el-input
              v-model="payloadStr"
              type="textarea"
              :rows="3"
              placeholder='请输入 JSON 格式负载，如 {"state": "ON"}'
            />
            <div v-if="payloadError" class="payload-error">{{ payloadError }}</div>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { controlApi, deviceApi } from '@/api'
import { getCommandTypeName, getMetricName } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import { CommandType, DeviceType } from '@/types'
import type { ControlRule, ControlRuleFormData, Device } from '@/types'

// 数据
const loading = ref(false)
const rules = ref<ControlRule[]>([])
const total = ref(0)

// 执行器列表（用于规则目标设备下拉）
const actuators = ref<Device[]>([])
const actuatorsLoading = ref(false)

async function loadActuators() {
  actuatorsLoading.value = true
  try {
    const result = await deviceApi.getDevices({
      type: DeviceType.ACTUATOR,
      page: 1,
      page_size: LARGE_PAGE_SIZE
    })
    actuators.value = result.items
  } catch {
    actuators.value = []
  } finally {
    actuatorsLoading.value = false
  }
}

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 弹窗
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)
const payloadStr = ref('{"state": "ON"}')
const payloadError = ref('')

const formData = reactive<ControlRuleFormData>({
  name: '',
  metric_code: 'TEMP',
  operator: '>',
  threshold: 30,
  target_device_id: null as unknown as number,
  command_type: CommandType.SWITCH,
  command_payload: {},
  enabled: true
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' },
    { min: 1, max: 64, message: '规则名称长度为 1-64 个字符', trigger: 'blur' }
  ],
  metric_code: [{ required: true, message: '请选择监控指标', trigger: 'change' }],
  operator: [{ required: true, message: '请选择比较运算符', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
  target_device_id: [{ required: true, message: '请选择目标设备', trigger: 'change' }],
  command_type: [{ required: true, message: '请选择命令类型', trigger: 'change' }]
}

// 监听命令类型变化
watch(() => formData.command_type, (type) => {
  if (type === CommandType.SWITCH) {
    payloadStr.value = '{"state": "ON"}'
  } else if (type === CommandType.SET_VALUE) {
    payloadStr.value = '{"value": 50}'
  }
})

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await controlApi.getRules({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    rules.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 切换规则状态
async function toggleRuleStatus(rule: ControlRule, enabled: boolean) {
  try {
    await controlApi.toggleRule(rule.id, enabled)
    ElMessage.success(enabled ? '规则已启用' : '规则已禁用')
    fetchData()
  } catch {
    // 错误已处理
  }
}

// 打开新增弹窗
async function openCreateDialog() {
  if (!actuatorsLoading.value && actuators.value.length === 0) {
    await loadActuators()
  }
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    name: '',
    metric_code: 'TEMP',
    operator: '>',
    threshold: 30,
    target_device_id: null,
    command_type: CommandType.SWITCH,
    command_payload: {},
    enabled: true
  })
  payloadStr.value = '{"state": "ON"}'
  payloadError.value = ''
  dialogVisible.value = true
}

// 打开编辑弹窗
async function openEditDialog(rule: ControlRule) {
  if (!actuatorsLoading.value && actuators.value.length === 0) {
    await loadActuators()
  }
  isEdit.value = true
  editingId.value = rule.id
  Object.assign(formData, {
    name: rule.name,
    metric_code: rule.metric_code,
    operator: rule.operator,
    threshold: rule.threshold,
    target_device_id: rule.target_device_id,
    command_type: rule.command_type,
    command_payload: rule.command_payload,
    enabled: rule.enabled
  })
  payloadStr.value = JSON.stringify(rule.command_payload)
  payloadError.value = ''
  dialogVisible.value = true
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  // 验证 JSON 格式
  try {
    formData.command_payload = JSON.parse(payloadStr.value)
    payloadError.value = ''
  } catch {
    payloadError.value = 'JSON 格式错误'
    return
  }

  submitLoading.value = true
  try {
    if (isEdit.value && editingId.value) {
      await controlApi.updateRule(editingId.value, formData)
      ElMessage.success('规则更新成功')
    } else {
      await controlApi.createRule(formData)
      ElMessage.success('规则创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

// 删除规则
async function handleDelete(rule: ControlRule) {
  try {
    await ElMessageBox.confirm(`确定要删除规则"${rule.name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await controlApi.deleteRule(rule.id)
    ElMessage.success('规则删除成功')
    fetchData()
  } catch {
    // 用户取消或错误已处理
  }
}

onMounted(() => {
  fetchData()
  loadActuators()
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
}
</style>
