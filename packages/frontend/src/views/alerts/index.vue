<template>
  <div class="alerts-page">
    <div class="page-header">
      <h1 class="page-title">告警列表</h1>
    </div>

    <div class="stats-bar">
      <div class="stat-item open"><span class="stat-dot"></span>开放: {{ stats.open_count }}</div>
      <div class="stat-item ack"><span class="stat-dot"></span>已确认: {{ stats.acknowledged_count }}</div>
      <div class="stat-item resolved"><span class="stat-dot"></span>已解决: {{ stats.resolved_count }}</div>
    </div>

    <div class="filter-section">
      <el-select v-model="filters.type" placeholder="告警类型" clearable style="width: 150px">
        <el-option label="阈值告警" value="THRESHOLD" />
        <el-option label="设备离线" value="DEVICE_OFFLINE" />
        <el-option label="系统" value="SYSTEM" />
      </el-select>
      <el-select v-model="filters.level" placeholder="告警级别" clearable style="width: 130px">
        <el-option label="严重" value="CRITICAL" />
        <el-option label="警告" value="WARN" />
        <el-option label="信息" value="INFO" />
      </el-select>
      <el-select v-model="filters.status" placeholder="状态" clearable style="width: 150px">
        <el-option label="开放" value="OPEN" />
        <el-option label="已确认" value="ACKNOWLEDGED" />
        <el-option label="已解决" value="RESOLVED" />
        <el-option label="已忽略" value="IGNORED" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="alerts" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="110">
          <template #default="{ row }">{{ getAlertTypeName(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getAlertLevelType(row.level)">{{ getAlertLevelName(row.level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="消息" min-width="220" />
        <el-table-column label="通道" width="120">
          <template #default="{ row }">
            <span v-if="row.sensor_channel_id">传感器 #{{ row.sensor_channel_id }}</span>
            <span v-else-if="row.actuator_channel_id">执行器 #{{ row.actuator_channel_id }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="getAlertStatusType(row.status)">{{ getAlertStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggered_at" label="触发时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.triggered_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canHandle" type="primary" link @click="goWorkflow(row.id)">流程</el-button>
            <el-button type="info" link @click="goTimeline(row.id)">时间线</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { alertApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime, getAlertLevelName, getAlertLevelType, getAlertStatusName, getAlertStatusType, getAlertTypeName } from '@/utils/format'
import type { Alert, AlertLevel, AlertStats, AlertStatus, AlertType } from '@/types'

const router = useRouter()
const { canControlDevice } = usePermission()
const canHandle = computed(() => canControlDevice())

const loading = ref(false)
const alerts = ref<Alert[]>([])
const stats = ref<AlertStats>({ open_count: 0, acknowledged_count: 0, resolved_count: 0, critical_count: 0, warn_count: 0, info_count: 0 })
const total = ref(0)

const filters = reactive({
  type: '' as '' | AlertType,
  level: '' as '' | AlertLevel,
  status: '' as '' | AlertStatus
})
const pagination = reactive({
  page: 1,
  pageSize: 20
})

async function fetchData() {
  loading.value = true
  try {
    const [alertData, statsData] = await Promise.all([
      alertApi.getAlerts({
        page: pagination.page,
        page_size: pagination.pageSize,
        type: filters.type || undefined,
        level: filters.level || undefined,
        status: filters.status || undefined
      }),
      alertApi.getAlertStats()
    ])
    alerts.value = alertData.items
    total.value = alertData.total
    stats.value = statsData
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.type = ''
  filters.level = ''
  filters.status = ''
  pagination.page = 1
  fetchData()
}

function goWorkflow(alertId: number) {
  router.push({ path: '/alerts/workflow', query: { alertId: String(alertId) } })
}
function goTimeline(alertId: number) {
  router.push({ path: '/alerts/timeline', query: { alertId: String(alertId) } })
}

onMounted(fetchData)
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

    &.resolved .stat-dot {
      background: var(--color-success);
      box-shadow: 0 0 6px rgba(103, 194, 58, 0.5);
    }
  }

  .filter-section {
    margin-bottom: 16px;
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
