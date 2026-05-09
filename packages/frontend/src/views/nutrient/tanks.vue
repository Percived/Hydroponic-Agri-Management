<template>
  <div class="nutrient-tanks-page">
    <div class="page-header">
      <h1 class="page-title">营养液槽</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增液槽
      </el-button>
    </div>

    <div class="filter-section">
      <el-select v-model="filters.growing_zone_id" placeholder="种植区" clearable filterable style="width: 180px">
        <el-option v-for="z in zones" :key="z.id" :label="`${z.name} (ID:${z.id})`" :value="z.id" />
      </el-select>
      <el-select v-model="filters.status" placeholder="状态" clearable style="width: 140px">
        <el-option label="使用中" value="IN_USE" />
        <el-option label="维护中" value="MAINTENANCE" />
        <el-option label="停用" value="DISABLED" />
      </el-select>
      <el-input v-model="filters.keyword" placeholder="搜索液槽编号" clearable style="width: 200px" />
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="tanks" v-loading="loading" stripe @expand-change="onExpandChange">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-monitoring">
              <SensorCard
                label="EC"
                :value="sensorData[String(row.id)]?.ec"
                unit="mS/cm"
              />
              <SensorCard
                label="pH"
                :value="sensorData[String(row.id)]?.ph"
                unit=""
              />
              <SensorCard
                label="液位"
                :value="sensorData[String(row.id)]?.level"
                unit="m"
              />
              <SensorCard
                label="温度"
                :value="sensorData[String(row.id)]?.temp"
                unit="°C"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="code" label="液槽编号" width="160" />
        <el-table-column prop="growing_zone_id" label="种植区ID" width="100" />
        <el-table-column label="总容积(L)" width="100">
          <template #default="{ row }">{{ row.total_volume_liter }}</template>
        </el-table-column>
        <el-table-column label="当前容积(L)" width="100">
          <template #default="{ row }">{{ row.current_volume_liter ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="传感器" width="90" align="center">
          <template #default="{ row }">
            <span class="sensor-binding-count">{{ boundSensorCount(row) }}/4</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="removeTank(row.id)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑液槽' : '新增液槽'" width="560px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="130px">
        <el-form-item label="种植区" prop="growing_zone_id">
          <el-select v-model="formData.growing_zone_id" placeholder="选择种植区" filterable style="width: 100%">
            <el-option v-for="z in zones" :key="z.id" :label="`${z.name} (ID:${z.id})`" :value="z.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="液槽编号" prop="code">
          <el-input v-model="formData.code" placeholder="请输入编号" maxlength="64" />
        </el-form-item>
        <el-form-item label="总容积(L)" prop="total_volume_liter">
          <el-input-number v-model="formData.total_volume_liter" :min="1" :precision="1" style="width: 100%" />
        </el-form-item>
        <el-divider content-position="left">传感器通道绑定</el-divider>
        <el-form-item label="EC 传感器通道">
          <el-select v-model="formData.ec_sensor_channel_id" placeholder="不绑定" clearable filterable style="width: 100%">
            <el-option v-for="ch in ecChannels" :key="ch.id" :label="channelLabel(ch)" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="pH 传感器通道">
          <el-select v-model="formData.ph_sensor_channel_id" placeholder="不绑定" clearable filterable style="width: 100%">
            <el-option v-for="ch in phChannels" :key="ch.id" :label="channelLabel(ch)" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="液位传感器通道">
          <el-select v-model="formData.level_sensor_channel_id" placeholder="不绑定" clearable filterable style="width: 100%">
            <el-option v-for="ch in levelChannels" :key="ch.id" :label="channelLabel(ch)" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="温度传感器通道">
          <el-select v-model="formData.temp_sensor_channel_id" placeholder="不绑定" clearable filterable style="width: 100%">
            <el-option v-for="ch in tempChannels" :key="ch.id" :label="channelLabel(ch)" :value="ch.id" />
          </el-select>
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
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { nutrientApi, greenhouseApi, deviceApi, telemetryApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { EXTRA_LARGE_PAGE_SIZE } from '@/utils/constants'
import type { NutrientTank, GrowingZone, SensorChannel } from '@/types'

// ── Sensor Card Component ──
import { h, defineComponent } from 'vue'

const SensorCard = defineComponent({
  name: 'SensorCard',
  props: {
    label: { type: String, required: true },
    value: { type: Object as () => { value: number; quality_flag: string; collected_at: string } | null, default: null },
    unit: { type: String, default: '' }
  },
  setup(props) {
    return () => {
      const d = props.value
      const cls = d
        ? d.quality_flag === 'normal' ? 'card-normal' : d.quality_flag === 'out_of_range' ? 'card-danger' : 'card-warn'
        : 'card-empty'
      return h('div', { class: `monitor-card ${cls}` }, [
        h('div', { class: 'card-label' }, props.label),
        d
          ? [
              h('div', { class: 'card-value' }, `${d.value} ${props.unit}`),
              h('div', { class: 'card-time' }, formatTime(d.collected_at)),
              h('div', { class: 'card-quality' }, d.quality_flag === 'normal' ? '正常' : d.quality_flag === 'out_of_range' ? '异常' : d.quality_flag)
            ]
          : h('div', { class: 'card-placeholder' }, '未绑定传感器')
      ])
    }
  }
})

function formatTime(iso: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// ── State ──
const loading = ref(false)
const tanks = ref<NutrientTank[]>([])
const total = ref(0)
const zones = ref<GrowingZone[]>([])
const allChannels = ref<SensorChannel[]>([])

// Filter channels by metric_code
const ecChannels = computed(() => allChannels.value.filter(ch => ch.metric_code === 'EC'))
const phChannels = computed(() => allChannels.value.filter(ch => ch.metric_code === 'PH'))
const levelChannels = computed(() => allChannels.value.filter(ch => ch.metric_code === 'LEVEL'))
const tempChannels = computed(() => allChannels.value.filter(ch => ch.metric_code === 'TEMP' || ch.metric_code === 'TEMPERATURE'))

function channelLabel(ch: SensorChannel): string {
  return `${ch.channel_code} (ID:${ch.id})`
}

const filters = reactive({
  growing_zone_id: undefined as number | undefined,
  status: '' as string,
  keyword: ''
})

const pagination = reactive({ page: 1, pageSize: 20 })

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive({
  growing_zone_id: undefined as number | undefined,
  code: '',
  total_volume_liter: 100,
  ec_sensor_channel_id: undefined as number | undefined,
  ph_sensor_channel_id: undefined as number | undefined,
  level_sensor_channel_id: undefined as number | undefined,
  temp_sensor_channel_id: undefined as number | undefined
})

const formRules: FormRules = {
  growing_zone_id: [{ required: true, message: '请输入种植区ID', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入液槽编号', trigger: 'blur' },
    { min: 1, max: 64, message: '编号长度为 1-64 个字符', trigger: 'blur' }
  ],
  total_volume_liter: [{ required: true, message: '请输入总容积', trigger: 'blur' }]
}

// ── Real-time sensor data ──
interface TankSensorData {
  ec: { value: number; quality_flag: string; collected_at: string } | null
  ph: { value: number; quality_flag: string; collected_at: string } | null
  level: { value: number; quality_flag: string; collected_at: string } | null
  temp: { value: number; quality_flag: string; collected_at: string } | null
}

const sensorData = ref<Record<string, TankSensorData>>({})
const expandedRows = ref<Set<string>>(new Set())
let pollTimer: ReturnType<typeof setInterval> | null = null

function getBoundChannelIds(row: NutrientTank): number[] {
  const ids: number[] = []
  if (row.ec_sensor_channel_id) ids.push(row.ec_sensor_channel_id)
  if (row.ph_sensor_channel_id) ids.push(row.ph_sensor_channel_id)
  if (row.level_sensor_channel_id) ids.push(row.level_sensor_channel_id)
  if (row.temp_sensor_channel_id) ids.push(row.temp_sensor_channel_id)
  return ids
}

function boundSensorCount(row: NutrientTank): number {
  return getBoundChannelIds(row).length
}

const channelIdToTankMap = ref<Record<string, { tankId: string; field: string }>>({})

function buildChannelMap() {
  const map: Record<string, { tankId: string; field: string }> = {}
  for (const tank of tanks.value) {
    const id = String(tank.id)
    if (tank.ec_sensor_channel_id) map[String(tank.ec_sensor_channel_id)] = { tankId: id, field: 'ec' }
    if (tank.ph_sensor_channel_id) map[String(tank.ph_sensor_channel_id)] = { tankId: id, field: 'ph' }
    if (tank.level_sensor_channel_id) map[String(tank.level_sensor_channel_id)] = { tankId: id, field: 'level' }
    if (tank.temp_sensor_channel_id) map[String(tank.temp_sensor_channel_id)] = { tankId: id, field: 'temp' }
  }
  channelIdToTankMap.value = map
}

async function fetchLatestTelemetry() {
  const allIds: number[] = []
  for (const tank of tanks.value) {
    allIds.push(...getBoundChannelIds(tank))
  }
  if (allIds.length === 0) return

  try {
    const resp = await telemetryApi.getChannelsLatest(allIds)
    if (!resp?.items) return
    // Group by tank
    const data: Record<string, TankSensorData> = {}
    for (const item of resp.items) {
      const mapping = channelIdToTankMap.value[String(item.sensor_channel_id)]
      if (!mapping) continue
      const tid = mapping.tankId
      if (!data[tid]) {
        data[tid] = { ec: null, ph: null, level: null, temp: null }
      }
      data[tid][mapping.field as keyof TankSensorData] = {
        value: item.value,
        quality_flag: item.quality_flag,
        collected_at: item.collected_at
      }
    }
    sensorData.value = data
  } catch {
    // silent - telemetry API handles errors silently
  }
}

function onExpandChange(_row: NutrientTank, expandedRows_: NutrientTank[]) {
  expandedRows.value.clear()
  for (const r of expandedRows_) {
    expandedRows.value.add(String(r.id))
  }
  // Start/stop polling based on whether any row is expanded
  if (expandedRows_.length > 0 && !pollTimer) {
    fetchLatestTelemetry()
    pollTimer = setInterval(fetchLatestTelemetry, 15000)
  } else if (expandedRows_.length === 0 && pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function statusName(status: string) {
  const map: Record<string, string> = { ACTIVE: '使用中', INACTIVE: '维护中', EMPTY: '空置' }
  return map[status] || status
}

function statusTagType(status: string) {
  const map: Record<string, string> = { ACTIVE: 'success', INACTIVE: 'warning', EMPTY: 'info' }
  return map[status] || 'info'
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.growing_zone_id) params.growing_zone_id = filters.growing_zone_id
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword
    const data = await nutrientApi.getNutrientTanks(params)
    tanks.value = data.items
    total.value = data.total
    buildChannelMap()
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.growing_zone_id = undefined
  filters.status = ''
  filters.keyword = ''
  pagination.page = 1
  fetchData()
}

function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    growing_zone_id: undefined,
    code: '',
    total_volume_liter: 100,
    ec_sensor_channel_id: undefined,
    ph_sensor_channel_id: undefined,
    level_sensor_channel_id: undefined,
    temp_sensor_channel_id: undefined
  })
  dialogVisible.value = true
}

function openEditDialog(tank: NutrientTank) {
  isEdit.value = true
  editingId.value = tank.id
  Object.assign(formData, {
    growing_zone_id: tank.growing_zone_id,
    code: tank.code,
    total_volume_liter: tank.total_volume_liter,
    ec_sensor_channel_id: tank.ec_sensor_channel_id ?? undefined,
    ph_sensor_channel_id: tank.ph_sensor_channel_id ?? undefined,
    level_sensor_channel_id: tank.level_sensor_channel_id ?? undefined,
    temp_sensor_channel_id: tank.temp_sensor_channel_id ?? undefined
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    const payload: Record<string, unknown> = {
      growing_zone_id: formData.growing_zone_id!,
      code: formData.code,
      total_volume_liter: formData.total_volume_liter
    }
    if (formData.ec_sensor_channel_id) payload.ec_sensor_channel_id = formData.ec_sensor_channel_id
    if (formData.ph_sensor_channel_id) payload.ph_sensor_channel_id = formData.ph_sensor_channel_id
    if (formData.level_sensor_channel_id) payload.level_sensor_channel_id = formData.level_sensor_channel_id
    if (formData.temp_sensor_channel_id) payload.temp_sensor_channel_id = formData.temp_sensor_channel_id

    if (isEdit.value && editingId.value) {
      await nutrientApi.updateNutrientTank(editingId.value, payload)
      ElMessage.success('液槽已更新')
    } else {
      await nutrientApi.createNutrientTank(payload as any)
      ElMessage.success('液槽已创建')
    }
    dialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    submitLoading.value = false
  }
}

async function removeTank(id: number) {
  await ElMessageBox.confirm('确认删除该液槽？', '提示', { type: 'warning' })
  await nutrientApi.deleteNutrientTank(id)
  ElMessage.success('已删除')
  fetchData()
}

async function loadZones() {
  try {
    const data = await greenhouseApi.getGrowingZones({ page_size: EXTRA_LARGE_PAGE_SIZE })
    zones.value = data.items
  } catch { /* ignore */ }
}

async function loadChannels() {
  try {
    const data = await deviceApi.getSensorChannels({ page_size: EXTRA_LARGE_PAGE_SIZE })
    allChannels.value = data.items
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchData()
  loadZones()
  loadChannels()
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>

<style scoped lang="scss">
.nutrient-tanks-page {
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
  }
  .filter-section {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    box-shadow: var(--shadow-card);
    margin-bottom: 16px;
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    align-items: center;
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
  .sensor-binding-count {
    color: var(--color-text-secondary);
    font-size: 13px;
  }
}

// ── Expand monitoring cards ──
.expand-monitoring {
  display: flex;
  gap: 16px;
  padding: 16px 20px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
  flex-wrap: wrap;

  .monitor-card {
    flex: 1;
    min-width: 160px;
    max-width: 220px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    padding: 14px 16px;
    text-align: center;
    transition: box-shadow 0.2s;

    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }

    .card-label {
      font-size: 12px;
      color: var(--color-text-secondary);
      margin-bottom: 6px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .card-value {
      font-size: 24px;
      font-weight: 700;
      margin-bottom: 4px;
    }

    .card-time {
      font-size: 11px;
      color: var(--color-text-tertiary);
      margin-bottom: 2px;
    }

    .card-quality {
      font-size: 11px;
    }

    .card-placeholder {
      color: var(--color-text-tertiary);
      font-size: 13px;
      padding: 12px 0;
      border: 1px dashed var(--border-color);
      border-radius: var(--radius-sm);
    }

    &.card-normal {
      border-left: 3px solid #67c23a;
      .card-value { color: #67c23a; }
      .card-quality { color: #67c23a; }
    }
    &.card-danger {
      border-left: 3px solid #f56c6c;
      .card-value { color: #f56c6c; }
      .card-quality { color: #f56c6c; }
    }
    &.card-warn {
      border-left: 3px solid #e6a23c;
      .card-value { color: #e6a23c; }
      .card-quality { color: #e6a23c; }
    }
    &.card-empty {
      border-left: 3px solid var(--border-color);
    }
  }
}
</style>
