<template>
  <div class="batch-review-page">
    <div class="page-header">
      <h1 class="page-title">批次复盘</h1>
      <el-space>
        <el-button @click="exportSummary('txt')" :disabled="!reviewLoaded">导出TXT</el-button>
        <el-button type="primary" @click="exportSummary('json')" :disabled="!reviewLoaded">导出JSON</el-button>
      </el-space>
    </div>

    <div class="filter-section">
      <el-form :inline="true">
        <el-form-item label="批次">
          <el-select v-model="selectedBatchId" filterable placeholder="选择批次" style="width: 320px">
            <el-option v-for="batch in batches" :key="batch.id" :label="`${batch.batch_no} (#${batch.id})`" :value="batch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="snapshotType" style="width: 140px">
            <el-option label="日报" value="DAILY" />
            <el-option label="周报" value="WEEKLY" />
            <el-option label="阶段汇总" value="STAGE_SUMMARY" />
            <el-option label="终期复盘" value="FINAL" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="range"
            type="datetimerange"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 360px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="fetchReview">查询</el-button>
          <el-button type="success" :loading="generating" @click="generateReview" :disabled="!selectedBatchId">生成复盘</el-button>
        </el-form-item>
      </el-form>
    </div>

    <batch-review-board :series="chartSeries" :events="eventPoints" :timeline-events="timelineEvents" />

    <!-- Snapshot list -->
    <el-card class="summary-card">
      <template #header>
        <span>历史快照 ({{ snapshotTotal }})</span>
      </template>
      <el-table :data="snapshotList" v-loading="snapshotLoading" stripe size="small">
        <el-table-column prop="snapshot_type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="snapshotTagType(row.snapshot_type)">{{ row.snapshot_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="window_start" label="窗口开始" width="180">
          <template #default="{ row }">{{ formatDateTime(row.window_start) }}</template>
        </el-table-column>
        <el-table-column prop="window_end" label="窗口结束" width="180">
          <template #default="{ row }">{{ formatDateTime(row.window_end) }}</template>
        </el-table-column>
        <el-table-column prop="generated_at" label="生成时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.generated_at) }}</template>
        </el-table-column>
        <el-table-column label="摘要" min-width="200">
          <template #default="{ row }">
            <template v-if="row.summary">
              <span v-if="row.summary.alert_count != null">告警: {{ row.summary.alert_count }} | </span>
              <span v-if="row.summary.metrics?.length">{{ row.summary.metrics.length }} 项指标 | </span>
              <span v-if="row.summary.energy_consumption?.total != null">能耗: {{ row.summary.energy_consumption.total }} kWh</span>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <div class="snapshot-pagination" v-if="snapshotTotal > snapshotPageSize">
        <el-pagination
          v-model:current-page="snapshotPage"
          :page-size="snapshotPageSize"
          :total="snapshotTotal"
          layout="prev, pager, next"
          size="small"
          @current-change="loadSnapshots"
        />
      </div>
    </el-card>

    <el-card class="summary-card">
      <template #header>复盘摘要</template>
      <el-descriptions :column="4" border v-if="snapshotSummary">
        <el-descriptions-item label="快照类型">{{ snapshotSummary.snapshot_type }}</el-descriptions-item>
        <el-descriptions-item label="告警数">{{ snapshotSummary.alert_count }}</el-descriptions-item>
        <el-descriptions-item label="控制动作">{{ snapshotSummary.control_count }}</el-descriptions-item>
        <el-descriptions-item label="失败数">{{ snapshotSummary.failure_count }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else description="暂无摘要快照" />
      <el-divider />
      <pre class="summary-json">{{ formattedSummary }}</pre>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cropApi } from '@/api'
import { get, post } from '@/api/request'
import BatchReviewBoard from '@/components/batch/BatchReviewBoard.vue'
import { formatDateTime } from '@/utils/format'
import type { CropBatch } from '@/types'

// 本地类型定义
interface BatchReviewSnapshot {
  snapshot_type: string
  alert_count: number
  control_count: number
  failure_count: number
}
interface ReviewTrendItem { metric_code: string; time: string; value?: number; avg?: number }
interface ReviewAlertItem { time?: string; triggered_at?: string; level?: string; message?: string }
interface ReviewControlItem { time?: string; created_at?: string; command_type?: string; status?: string }
interface ReviewData {
  environment_trends: ReviewTrendItem[]
  alerts: ReviewAlertItem[]
  controls: ReviewControlItem[]
  snapshots: BatchReviewSnapshot[]
  summary: Record<string, unknown>
}

// 指标名称映射
const MetricNames: Record<string, string> = {
  TEMP: '温度', HUMIDITY: '湿度', PH: 'pH值', EC: '电导率',
  CO2: 'CO2', LIGHT: '光照', WATER_TEMP: '水温', DISSOLVED_O2: '溶氧',
  DO: '溶解氧', LEVEL: '液位', ORP: '氧化还原电位', TDS: '总溶解固体',
  O3: '臭氧浓度', TURBIDITY: '浊度', FLOW_RATE: '流量'
}

const loading = ref(false)
const generating = ref(false)
const batches = ref<CropBatch[]>([])
const selectedBatchId = ref<number>()
const snapshotType = ref('DAILY')
const range = ref<[string, string] | null>(null)
const reviewData = ref<ReviewData | null>(null)

// Snapshot list
const snapshotLoading = ref(false)
const snapshotList = ref<any[]>([])
const snapshotTotal = ref(0)
const snapshotPage = ref(1)
const snapshotPageSize = ref(10)

const reviewLoaded = computed(() => !!reviewData.value)
const snapshotSummary = computed<BatchReviewSnapshot | null>(
  () => (reviewData.value?.snapshots?.[0] as BatchReviewSnapshot) || null
)
const formattedSummary = computed(() => JSON.stringify(reviewData.value?.summary || {}, null, 2))

const chartSeries = computed(() => {
  const trends = reviewData.value?.environment_trends || []
  const group: Record<string, Array<{ time: string; value: number }>> = {}
  for (const item of trends) {
    const code = item.metric_code || 'UNKNOWN'
    if (!group[code]) group[code] = []
    const value = Number(item.value ?? item.avg ?? 0)
    group[code].push({ time: item.time, value })
  }
  return Object.entries(group).map(([metric, data]) => ({
    name: MetricNames[metric] || metric,
    data
  }))
})

const eventPoints = computed(() => {
  const alerts = (reviewData.value?.alerts || []).map((a) => ({
    time: a.time || a.triggered_at || '',
    value: baselineValue.value,
    label: `告警:${a.message || '-'}`,
    eventType: 'alert' as const
  }))
  const controls = (reviewData.value?.controls || []).map((c) => ({
    time: c.time || c.created_at || '',
    value: baselineValue.value,
    label: `控制:${c.command_type || '-'}(${c.status || '-'})`,
    eventType: 'control' as const
  }))
  return [...alerts, ...controls].filter((e) => !!e.time)
})

const timelineEvents = computed(() => {
  const alerts = (reviewData.value?.alerts || []).map((a) => ({
    type: 'alert' as const,
    time: a.time || a.triggered_at || '',
    label: `${a.level || '-'} ${a.message || '-'}`
  }))
  const controls = (reviewData.value?.controls || []).map((c) => ({
    type: 'control' as const,
    time: c.time || c.created_at || '',
    label: `${c.command_type || '-'} ${c.status || '-'}`
  }))
  return [...alerts, ...controls].filter((e) => !!e.time).sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime())
})

const baselineValue = computed(() => {
  const values = chartSeries.value.flatMap((s) => s.data.map((d) => d.value))
  return values.length > 0 ? Math.max(...values) : 0
})

async function initBatches() {
  const res = await cropApi.getBatches({ page: 1, page_size: 200 })
  batches.value = res.items
  if (batches.value.length > 0) {
    selectedBatchId.value = batches.value[0].id
    loadSnapshots()
  }
}

async function loadSnapshots() {
  if (!selectedBatchId.value) return
  snapshotLoading.value = true
  try {
    const data = await get<{ items: any[]; total: number }>(`/reviews/batches/${selectedBatchId.value}`, {
      snapshot_type: snapshotType.value,
      page: snapshotPage.value,
      page_size: snapshotPageSize.value
    })
    snapshotList.value = data.items || []
    snapshotTotal.value = data.total || 0
  } catch {
    // ignore
  } finally {
    snapshotLoading.value = false
  }
}

async function generateReview() {
  if (!selectedBatchId.value || !range.value) {
    ElMessage.warning('请选择批次和时间范围')
    return
  }
  generating.value = true
  try {
    await post('/reviews/generate', {
      batch_id: selectedBatchId.value,
      snapshot_type: snapshotType.value,
      window_start: range.value[0],
      window_end: range.value[1]
    })
    ElMessage.success('复盘快照已生成')
    await loadSnapshots()
  } catch {
    ElMessage.error('生成失败')
  } finally {
    generating.value = false
  }
}

function snapshotTagType(type: string) {
  const map: Record<string, string> = { DAILY: 'info', WEEKLY: '', STAGE_SUMMARY: 'warning', FINAL: 'success' }
  return map[type] || 'info'
}

async function fetchReview() {
  if (!selectedBatchId.value) {
    ElMessage.warning('请先选择批次')
    return
  }
  loading.value = true
  try {
    const data = await get<ReviewData>(`/reviews/batches/${selectedBatchId.value}/review`, {
      from: range.value?.[0],
      to: range.value?.[1],
      snapshot_type: snapshotType.value
    })
    reviewData.value = data
  } finally {
    loading.value = false
  }
}

function exportSummary(type: 'txt' | 'json') {
  if (!reviewData.value || !selectedBatchId.value) return
  const fileName = `batch-review-${selectedBatchId.value}-${Date.now()}.${type}`
  const content =
    type === 'json'
      ? JSON.stringify(reviewData.value, null, 2)
      : [
          `Batch ID: ${selectedBatchId.value}`,
          `Snapshot Type: ${snapshotType.value}`,
          `Alert Count: ${snapshotSummary.value?.alert_count ?? '-'}`,
          `Control Count: ${snapshotSummary.value?.control_count ?? '-'}`,
          `Failure Count: ${snapshotSummary.value?.failure_count ?? '-'}`,
          `Summary: ${JSON.stringify(reviewData.value.summary || {})}`
        ].join('\n')
  const blob = new Blob([content], { type: type === 'json' ? 'application/json' : 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  await initBatches()
  const end = new Date()
  const start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000)
  range.value = [start.toISOString(), end.toISOString()]
})
</script>

<style scoped lang="scss">
.batch-review-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  .page-title {
    margin: 0;
    font-size: 22px;
    font-weight: 700;
  }
  .filter-section {
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
  }
  .summary-card {
    margin-top: 16px;
  }
  .snapshot-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border-light);
  }
  .summary-json {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    font-size: 12px;
  }
}
</style>
