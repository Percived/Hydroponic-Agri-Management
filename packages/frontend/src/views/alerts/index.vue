<template>
    <div class="alerts-page">
      <div class="page-header">
        <h1 class="page-title">告警中心</h1>
      </div>

      <!-- 告警统计 -->
      <div class="stats-bar">
        <div class="stat-item open">
          <span class="stat-dot"></span>
          开放: {{ stats.open_count }}
        </div>
        <div class="stat-item ack">
          <span class="stat-dot"></span>
          已确认: {{ stats.ack_count }}
        </div>
        <div class="stat-item closed">
          <span class="stat-dot"></span>
          已关闭: {{ stats.closed_count }}
        </div>
      </div>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-select v-model="filters.type" placeholder="告警类型" clearable style="width: 120px">
          <el-option label="阈值告警" value="THRESHOLD" />
          <el-option label="离线告警" value="OFFLINE" />
          <el-option label="设备故障" value="DEVICE_ERROR" />
        </el-select>
        <el-select v-model="filters.level" placeholder="告警级别" clearable style="width: 120px">
          <el-option label="严重" value="CRITICAL" />
          <el-option label="警告" value="WARN" />
          <el-option label="信息" value="INFO" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="开放" value="OPEN" />
          <el-option label="已确认" value="ACK" />
          <el-option label="已关闭" value="CLOSED" />
        </el-select>
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 告警列表 -->
      <div class="table-container">
        <el-table :data="alerts" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              {{ getAlertTypeName(row.type) }}
            </template>
          </el-table-column>
          <el-table-column prop="level" label="级别" width="100">
            <template #default="{ row }">
              <el-tag :type="getAlertLevelType(row.level)">
                {{ getAlertLevelName(row.level) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="消息" min-width="200" />
          <el-table-column prop="device_name" label="设备" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getAlertStatusType(row.status)">
                {{ getAlertStatusName(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="triggered_at" label="触发时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.triggered_at || row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="canHandle && row.status !== 'CLOSED'"
                type="primary"
                link
                @click="openHandleDialog(row)"
              >
                处理
              </el-button>
              <span v-else class="text-muted">-</span>
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

      <!-- 处理告警弹窗 -->
      <el-dialog v-model="handleDialogVisible" title="处理告警" width="400px">
        <el-form ref="formRef" :model="handleForm" :rules="handleRules" label-width="80px">
          <el-form-item label="状态" prop="status">
            <el-select v-model="handleForm.status" placeholder="请选择状态">
              <el-option label="已确认" value="ACK" />
              <el-option label="已关闭" value="CLOSED" />
            </el-select>
          </el-form-item>
          <el-form-item label="备注" prop="comment">
            <el-input
              v-model="handleForm.comment"
              type="textarea"
              :rows="3"
              placeholder="请输入备注（可选）"
              maxlength="255"
              show-word-limit
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="handleDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleAlert">确定</el-button>
        </template>
      </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { alertApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime, getAlertTypeName, getAlertLevelType, getAlertLevelName, getAlertStatusType, getAlertStatusName } from '@/utils/format'
import { AlertStatus } from '@/types'
import type { Alert, AlertStats, AlertType, AlertLevel, UpdateAlertStatusParams } from '@/types'

const { canControlDevice } = usePermission()
const canHandle = computed(() => canControlDevice())

// 数据
const loading = ref(false)
const alerts = ref<Alert[]>([])
const stats = ref<AlertStats>({
  open_count: 0,
  ack_count: 0,
  closed_count: 0
})
const total = ref(0)

// 筛选条件
const filters = reactive({
  type: '',
  level: '',
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 处理弹窗
const handleDialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const currentAlert = ref<Alert | null>(null)

const handleForm = reactive({
  status: 'ACK' as AlertStatus,
  comment: ''
})

const handleRules: FormRules = {
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const [alertData, statsData] = await Promise.all([
      alertApi.getAlerts({
        page: pagination.page,
        page_size: pagination.pageSize,
        type: (filters.type as AlertType) || undefined,
        level: (filters.level as AlertLevel) || undefined,
        status: (filters.status as AlertStatus) || undefined
      }),
      alertApi.getAlertStats()
    ])
    alerts.value = alertData.items
    total.value = alertData.total
    stats.value = statsData
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 重置筛选
function resetFilters() {
  filters.type = ''
  filters.level = ''
  filters.status = ''
  pagination.page = 1
  fetchData()
}

// 打开处理弹窗
function openHandleDialog(alert: Alert) {
  currentAlert.value = alert
  handleForm.status = alert.status === AlertStatus.OPEN ? AlertStatus.ACK : AlertStatus.CLOSED
  handleForm.comment = ''
  handleDialogVisible.value = true
}

// 处理告警
async function handleAlert() {
  if (!formRef.value || !currentAlert.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    const params: UpdateAlertStatusParams = {
      status: handleForm.status
    }
    if (handleForm.comment) {
      params.comment = handleForm.comment
    }
    await alertApi.updateAlertStatus(currentAlert.value.id, params)
    ElMessage.success('告警处理成功')
    handleDialogVisible.value = false
    fetchData()
  } catch {
    // 错误已处理
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.alerts-page {
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

  .stats-bar {
    display: flex;
    gap: 24px;
    background: var(--bg-card);
    padding: 16px 20px;
    border-radius: var(--radius-md);
    margin-bottom: 16px;
    box-shadow: var(--shadow-card);
  }

  .stat-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;

    .stat-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    &.open .stat-dot {
      background: var(--color-danger);
      box-shadow: 0 0 6px rgba(245, 108, 108, 0.5);
    }

    &.ack .stat-dot {
      background: var(--color-warning);
      box-shadow: 0 0 6px rgba(230, 162, 60, 0.5);
    }

    &.closed .stat-dot {
      background: var(--color-success);
      box-shadow: 0 0 6px rgba(103, 194, 58, 0.5);
    }
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

  .text-muted {
    color: var(--color-text-placeholder);
  }
}
</style>
