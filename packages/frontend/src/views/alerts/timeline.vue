<template>
  <div class="timeline-page">
    <div class="page-header">
      <h1 class="page-title">告警时间线</h1>
    </div>

    <div class="panel filter-panel">
      <el-form :inline="true">
        <el-form-item label="告警">
          <el-select v-model="selectedAlertId" filterable placeholder="选择告警" style="width: 340px" @change="loadTimeline">
            <el-option
              v-for="item in alerts"
              :key="item.id"
              :label="`#${item.id} ${item.message}`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :disabled="!selectedAlertId" @click="loadTimeline">刷新</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="panel">
      <el-timeline v-loading="loading">
        <el-timeline-item
          v-for="event in timelineDisplay"
          :key="event.key"
          :timestamp="event.timeLabel"
          :type="event.type"
        >
          <div class="line-title">{{ event.title }}</div>
          <div class="line-meta">{{ event.meta }}</div>
          <div class="line-comment">{{ event.comment }}</div>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-if="!loading && timelineDisplay.length === 0" description="暂无时间线数据" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { alertApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { Alert, AlertTimelineEvent } from '@/types'

const route = useRoute()

const loading = ref(false)
const alerts = ref<Alert[]>([])
const selectedAlertId = ref<number>()
const timeline = ref<AlertTimelineEvent[]>([])

const timelineDisplay = computed(() => {
  return timeline.value.map((event) => ({
    key: `t-${event.id}`,
    time: event.event_time,
    timeLabel: formatDateTime(event.event_time),
    type: event.event_source === 'MANUAL' ? 'primary' : 'info',
    title: `事件: ${event.event_type}`,
    meta: `来源: ${event.event_source} / 操作人: ${event.operator_id || '-'}`,
    comment: event.comment || '-'
  })).sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime())
})

async function initOptions() {
  const alertRes = await alertApi.getAlerts({ page: 1, page_size: 200 })
  alerts.value = alertRes.items
  const routeAlertId = Number(route.query.alertId || 0)
  if (routeAlertId) {
    selectedAlertId.value = routeAlertId
  } else if (alerts.value.length > 0) {
    selectedAlertId.value = alerts.value[0].id
  }
}

async function loadTimeline() {
  if (!selectedAlertId.value) return
  loading.value = true
  try {
    const timelineRes = await alertApi.getAlertTimeline(selectedAlertId.value)
    timeline.value = timelineRes.items || []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await initOptions()
  await loadTimeline()
})
</script>

<style scoped lang="scss">
.timeline-page {
  .page-header {
    margin-bottom: 16px;
  }
  .page-title {
    margin: 0;
    font-size: 22px;
    font-weight: 700;
  }
  .panel {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
    padding: 16px;
  }
  .filter-panel {
    margin-bottom: 16px;
  }
  .line-title {
    font-weight: 600;
  }
  .line-meta {
    color: var(--color-text-secondary);
    font-size: 12px;
    margin-top: 2px;
  }
  .line-comment {
    margin-top: 4px;
    font-size: 12px;
  }
}
</style>
