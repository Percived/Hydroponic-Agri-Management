<template>
  <div class="alert-workflow-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/alerts/list')" text>
          <el-icon><ArrowLeft /></el-icon>返回告警列表
        </el-button>
        <h1 class="page-title">告警处置 #{{ alertId }}</h1>
      </div>
      <el-button @click="goTimeline">
        <el-icon><Clock /></el-icon>完整时间线
      </el-button>
    </div>

    <template v-if="alert">
      <!-- Alert Detail Card -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-title">告警详情</span>
            <el-tag :type="alertLevelTag(alert.level)" size="large">{{ alertLevelName(alert.level) }}</el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="ID">{{ alert.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(alert.status)">{{ statusName(alert.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="类型">{{ typeName(alert.type) }}</el-descriptions-item>
          <el-descriptions-item label="指标代码">{{ alert.metric_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="触发值">{{ alert.trigger_value ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="触发时间">{{ formatDateTime(alert.triggered_at) }}</el-descriptions-item>
          <el-descriptions-item label="解决时间">{{ formatDateTime(alert.resolved_at) }}</el-descriptions-item>
          <el-descriptions-item label="解决人">{{ alert.resolved_by ? `用户 #${alert.resolved_by}` : '-' }}</el-descriptions-item>
          <el-descriptions-item label="设备编号">{{ alert.device_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="消息" :span="2">{{ alert.message }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Actions Card -->
      <el-card class="section-card" v-if="canHandle">
        <template #header>
          <span class="card-title">状态操作</span>
        </template>

        <div class="action-buttons">
          <template v-if="alert.status === 'OPEN'">
            <el-button type="primary" @click="openTransitionDialog('ACKNOWLEDGED')">
              <el-icon><Check /></el-icon>确认告警
            </el-button>
            <el-button type="success" @click="openTransitionDialog('RESOLVED')">
              <el-icon><CircleCheck /></el-icon>标记解决
            </el-button>
            <el-button type="info" @click="openTransitionDialog('IGNORED')">
              <el-icon><Remove /></el-icon>忽略
            </el-button>
          </template>

          <template v-else-if="alert.status === 'ACKNOWLEDGED'">
            <el-button type="success" @click="openTransitionDialog('RESOLVED')">
              <el-icon><CircleCheck /></el-icon>标记解决
            </el-button>
            <el-button type="info" @click="openTransitionDialog('IGNORED')">
              <el-icon><Remove /></el-icon>忽略
            </el-button>
            <el-button @click="openTransitionDialog('OPEN')">
              <el-icon><RefreshLeft /></el-icon>重新打开
            </el-button>
          </template>

          <template v-else-if="alert.status === 'RESOLVED'">
            <el-button @click="openTransitionDialog('OPEN')">
              <el-icon><RefreshLeft /></el-icon>重新打开
            </el-button>
          </template>

          <template v-else-if="alert.status === 'IGNORED'">
            <el-button @click="openTransitionDialog('OPEN')">
              <el-icon><RefreshLeft /></el-icon>重新打开
            </el-button>
          </template>
        </div>

        <!-- Comment form -->
        <div class="comment-section">
          <el-input
            v-model="commentText"
            type="textarea"
            :rows="2"
            placeholder="添加备注（可选）"
            maxlength="255"
            show-word-limit
          />
          <el-button type="primary" :loading="commentLoading" @click="addComment" style="margin-top: 8px">
            添加备注
          </el-button>
        </div>
      </el-card>

      <!-- Status Transition Dialog -->
      <el-dialog v-model="transitionDialogVisible" :title="`确认${statusName(transitionTarget)}`" width="440px">
        <p style="margin-bottom: 12px">
          将告警状态从 <strong>{{ statusName(alert.status) }}</strong> 变更为 <strong>{{ statusName(transitionTarget) }}</strong>
        </p>
        <el-input
          v-model="transitionComment"
          type="textarea"
          :rows="2"
          placeholder="备注（可选）"
          maxlength="255"
          show-word-limit
        />
        <template #footer>
          <el-button @click="transitionDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="transitionLoading" @click="doTransition">确认</el-button>
        </template>
      </el-dialog>

      <!-- Timeline Card -->
      <el-card class="section-card">
        <template #header>
          <span class="card-title">事件时间线</span>
        </template>
        <el-timeline v-if="timeline.length">
          <el-timeline-item
            v-for="event in sortedTimeline"
            :key="event.id"
            :timestamp="formatDateTime(event.event_time)"
            :type="event.event_source === 'MANUAL' ? 'primary' : 'info'"
          >
            <div class="tl-title">{{ event.event_type }}</div>
            <div class="tl-meta">来源：{{ event.event_source }} / 操作人：{{ event.operator_id || '-' }}</div>
            <div v-if="event.comment" class="tl-comment">{{ event.comment }}</div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无时间线事件" :image-size="60" />
      </el-card>
    </template>

    <el-empty v-else-if="!loading" description="告警不存在" :image-size="80" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Check, CircleCheck, Clock, RefreshLeft, Remove } from '@element-plus/icons-vue'
import { alertApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime } from '@/utils/format'
import type { Alert, AlertStatus, AlertTimelineEvent } from '@/types'

const route = useRoute()
const router = useRouter()
const { canControlDevice } = usePermission()

const canHandle = computed(() => canControlDevice())

const alertId = computed(() => Number(route.query.alertId || 0))

const loading = ref(false)
const alert = ref<Alert | null>(null)
const timeline = ref<AlertTimelineEvent[]>([])

const sortedTimeline = computed(() =>
  [...timeline.value].sort((a, b) => new Date(a.event_time).getTime() - new Date(b.event_time).getTime())
)

// Comment
const commentText = ref('')
const commentLoading = ref(false)

// Status transition dialog
const transitionDialogVisible = ref(false)
const transitionTarget = ref<AlertStatus>('OPEN')
const transitionComment = ref('')
const transitionLoading = ref(false)

function alertLevelTag(level: string) {
  const map: Record<string, string> = { INFO: 'info', WARN: 'warning', CRITICAL: 'danger' }
  return map[level] || 'info'
}

function alertLevelName(level: string) {
  const map: Record<string, string> = { INFO: '信息', WARN: '警告', CRITICAL: '严重' }
  return map[level] || level
}

function statusTagType(status: string) {
  const map: Record<string, string> = { OPEN: 'danger', ACKNOWLEDGED: 'warning', RESOLVED: 'success', IGNORED: 'info' }
  return map[status] || 'info'
}

function statusName(status: string) {
  const map: Record<string, string> = { OPEN: '开放', ACKNOWLEDGED: '已确认', RESOLVED: '已解决', IGNORED: '已忽略' }
  return map[status] || status
}

function typeName(type: string) {
  const map: Record<string, string> = { THRESHOLD: '阈值告警', DEVICE_OFFLINE: '设备离线', SYSTEM: '系统' }
  return map[type] || type
}

function goTimeline() {
  router.push({ path: '/alerts/timeline', query: { alertId: String(alertId.value) } })
}

async function fetchAlert() {
  if (!alertId.value) {
    router.replace('/alerts/list')
    return
  }
  loading.value = true
  try {
    alert.value = await alertApi.getAlert(alertId.value)
  } catch {
    ElMessage.error('加载告警详情失败')
    alert.value = null
  } finally {
    loading.value = false
  }
}

async function fetchTimeline() {
  if (!alertId.value) return
  try {
    const res = await alertApi.getAlertTimeline(alertId.value)
    timeline.value = res.items || []
  } catch {
    timeline.value = []
  }
}

function openTransitionDialog(target: AlertStatus) {
  transitionTarget.value = target
  transitionComment.value = ''
  transitionDialogVisible.value = true
}

async function doTransition() {
  if (!alert.value) return
  transitionLoading.value = true
  try {
    const payload: { status: AlertStatus; resolved_at?: string; comment?: string } = {
      status: transitionTarget.value
    }
    if (transitionTarget.value === 'RESOLVED') {
      payload.resolved_at = new Date().toISOString()
    }
    if (transitionComment.value.trim()) {
      payload.comment = transitionComment.value.trim()
    }
    await alertApi.updateAlertStatus(alert.value.id, payload)
    ElMessage.success(`告警状态已更新为 ${statusName(transitionTarget.value)}`)
    transitionDialogVisible.value = false
    await fetchAlert()
    await fetchTimeline()
  } catch {
    ElMessage.error('状态更新失败')
  } finally {
    transitionLoading.value = false
  }
}

async function addComment() {
  if (!alert.value || !commentText.value.trim()) return
  commentLoading.value = true
  try {
    await alertApi.createAlertTimelineEvent(alert.value.id, {
      event_type: 'COMMENT',
      event_source: 'MANUAL',
      comment: commentText.value.trim(),
      event_time: new Date().toISOString()
    })
    ElMessage.success('备注已添加')
    commentText.value = ''
    await fetchTimeline()
  } catch {
    ElMessage.error('添加备注失败')
  } finally {
    commentLoading.value = false
  }
}

onMounted(async () => {
  await fetchAlert()
  await fetchTimeline()
})
</script>

<style scoped lang="scss">
.alert-workflow-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .page-title {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
    }
  }

  .section-card {
    margin-bottom: 16px;
    .card-title {
      font-weight: 600;
    }
    .card-header-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .action-buttons {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 16px;
  }

  .comment-section {
    border-top: 1px solid var(--border-light);
    padding-top: 16px;
  }

  .tl-title {
    font-weight: 600;
  }

  .tl-meta {
    color: var(--text-secondary);
    font-size: 12px;
    margin-top: 2px;
  }

  .tl-comment {
    margin-top: 4px;
    font-size: 13px;
    color: var(--text-primary);
  }
}
</style>
