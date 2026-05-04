<template>
  <AppLayout>
    <div class="audit-logs-page">
      <div class="page-header">
        <h1 class="page-title">审计日志</h1>
      </div>

      <!-- 筛选区 -->
      <div class="filter-section">
        <el-select v-model="filters.action" placeholder="操作类型" clearable style="width: 150px">
          <el-option label="登录" value="LOGIN" />
          <el-option label="登出" value="LOGOUT" />
          <el-option label="创建设备" value="CREATE_DEVICE" />
          <el-option label="更新设备" value="UPDATE_DEVICE" />
          <el-option label="删除设备" value="DELETE_DEVICE" />
          <el-option label="控制命令" value="CONTROL_CMD" />
          <el-option label="创建规则" value="CREATE_RULE" />
          <el-option label="更新规则" value="UPDATE_RULE" />
          <el-option label="删除规则" value="DELETE_RULE" />
          <el-option label="创建用户" value="CREATE_USER" />
          <el-option label="更新用户" value="UPDATE_USER" />
          <el-option label="处理告警" value="UPDATE_ALERT" />
        </el-select>
        <el-date-picker
          v-model="timeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DDTHH:mm:ss"
          style="width: 360px"
        />
        <el-button type="primary" @click="fetchData">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <!-- 日志列表 -->
      <div class="table-container">
        <el-table :data="logs" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户" width="120" />
          <el-table-column prop="action" label="操作类型" width="120">
            <template #default="{ row }">
              {{ getAuditActionName(row.action) }}
            </template>
          </el-table-column>
          <el-table-column prop="target_type" label="目标类型" width="100">
            <template #default="{ row }">
              {{ getTargetTypeName(row.target_type) }}
            </template>
          </el-table-column>
          <el-table-column prop="target_id" label="目标ID" width="80">
            <template #default="{ row }">
              {{ row.target_id || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="detail" label="详情" min-width="200">
            <template #default="{ row }">
              {{ row.detail || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="ip_address" label="IP地址" width="130">
            <template #default="{ row }">
              {{ row.ip_address || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
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
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { AppLayout } from '@/components/layout'
import { auditApi } from '@/api'
import { formatDateTime, getAuditActionName, getTargetTypeName } from '@/utils/format'
import type { AuditLog, AuditAction } from '@/types'

// 数据
const loading = ref(false)
const logs = ref<AuditLog[]>([])
const total = ref(0)

// 筛选条件
const filters = reactive({
  action: ''
})
const timeRange = ref<[string, string] | null>(null)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
})

// 获取数据
async function fetchData() {
  loading.value = true
  try {
    const data = await auditApi.getAuditLogs({
      page: pagination.page,
      page_size: pagination.pageSize,
      action: (filters.action as AuditAction) || undefined,
      start_time: timeRange.value?.[0] || undefined,
      end_time: timeRange.value?.[1] || undefined
    })
    logs.value = data.items
    total.value = data.total
  } catch {
    // 错误已处理
  } finally {
    loading.value = false
  }
}

// 重置筛选
function resetFilters() {
  filters.action = ''
  timeRange.value = null
  pagination.page = 1
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.audit-logs-page {
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
}
</style>
