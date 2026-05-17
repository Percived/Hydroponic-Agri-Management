<template>
  <div class="nutrient-tank-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/nutrient/tanks')" text>
          <el-icon><ArrowLeft /></el-icon>返回营养液槽
        </el-button>
        <h1 class="page-title">{{ tank?.code || '液槽详情' }}</h1>
        <el-tag v-if="tank" size="large" :type="tankStatusTagType">{{ tankStatusLabel }}</el-tag>
      </div>
      <div class="header-right" v-if="tank">
        <el-button v-if="canManage" type="primary" @click="openEditDialog">
          <el-icon><Edit /></el-icon>编辑
        </el-button>
        <el-button v-if="canDelete" type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>删除
        </el-button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading && !tank" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- Error -->
    <el-result
      v-else-if="errorMsg" icon="error" :title="errorMsg"
      sub-title="请检查网络连接或返回列表重试"
    >
      <template #extra>
        <el-button type="primary" @click="loadAllData">重新加载</el-button>
        <el-button @click="router.push('/nutrient/tanks')">返回列表</el-button>
      </template>
    </el-result>

    <!-- Content -->
    <template v-else-if="tank">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 基本信息 + 传感器 -->
        <el-tab-pane label="基本信息" name="info">
          <el-card shadow="never" class="info-card">
            <template #header><span class="card-title">液槽信息</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="液槽编号">{{ tank.code }}</el-descriptions-item>
              <el-descriptions-item label="所属种植区">{{ growingZoneName(tank.growing_zone_id) }}</el-descriptions-item>
              <el-descriptions-item label="总容积 (L)">{{ tank.total_volume_liter }}</el-descriptions-item>
              <el-descriptions-item label="当前容积 (L)">{{ tank.current_volume_liter ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="tankStatusTagType">{{ tankStatusLabel }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="传感器绑定数">{{ sensorBindingCount }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDateTime(tank.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatDateTime(tank.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- Sensor Cards -->
          <el-card shadow="never">
            <template #header><span class="card-title">实时传感器数据</span></template>
            <div class="sensor-grid">
              <div class="monitor-card" :class="sensorCardClass('ec')">
                <div class="sensor-label">EC</div>
                <div class="sensor-value" v-if="sensorData.ec">
                  {{ sensorData.ec.value }} <span class="sensor-unit">mS/cm</span>
                </div>
                <div class="sensor-time" v-if="sensorData.ec">{{ formatDateTime(sensorData.ec.collected_at) }}</div>
                <div class="sensor-empty" v-else>未绑定传感器</div>
              </div>
              <div class="monitor-card" :class="sensorCardClass('ph')">
                <div class="sensor-label">pH</div>
                <div class="sensor-value" v-if="sensorData.ph">
                  {{ sensorData.ph.value }}
                </div>
                <div class="sensor-time" v-if="sensorData.ph">{{ formatDateTime(sensorData.ph.collected_at) }}</div>
                <div class="sensor-empty" v-else>未绑定传感器</div>
              </div>
              <div class="monitor-card" :class="sensorCardClass('level')">
                <div class="sensor-label">液位</div>
                <div class="sensor-value" v-if="sensorData.level">
                  {{ sensorData.level.value }} <span class="sensor-unit">m</span>
                </div>
                <div class="sensor-time" v-if="sensorData.level">{{ formatDateTime(sensorData.level.collected_at) }}</div>
                <div class="sensor-empty" v-else>未绑定传感器</div>
              </div>
              <div class="monitor-card" :class="sensorCardClass('temp')">
                <div class="sensor-label">水温</div>
                <div class="sensor-value" v-if="sensorData.temp">
                  {{ sensorData.temp.value }} <span class="sensor-unit">°C</span>
                </div>
                <div class="sensor-time" v-if="sensorData.temp">{{ formatDateTime(sensorData.temp.collected_at) }}</div>
                <div class="sensor-empty" v-else>未绑定传感器</div>
              </div>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Tab 2: 溶液更换 -->
        <el-tab-pane label="溶液更换" name="solutions">
          <el-table :data="solutionChanges" v-loading="solutionsLoading" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column label="更换类型" width="110">
              <template #default="{ row }">{{ changeTypeLabel(row.change_type) }}</template>
            </el-table-column>
            <el-table-column prop="volume_replaced_liter" label="更换量 (L)" width="110" />
            <el-table-column label="换前 EC/pH" width="140">
              <template #default="{ row }">
                <span v-if="row.before_ec != null || row.before_ph != null">
                  {{ row.before_ec ?? '-' }} / {{ row.before_ph ?? '-' }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="换后 EC/pH" width="140">
              <template #default="{ row }">
                <span v-if="row.after_ec != null || row.after_ph != null">
                  {{ row.after_ec ?? '-' }} / {{ row.after_ph ?? '-' }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="添加液 A/B" width="120">
              <template #default="{ row }">
                <template v-if="row.nutrient_a_added_ml != null || row.nutrient_b_added_ml != null">
                  A:{{ row.nutrient_a_added_ml ?? 0 }} B:{{ row.nutrient_b_added_ml ?? 0 }} ml
                </template>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="酸/碱 (ml)" width="120">
              <template #default="{ row }">
                <template v-if="row.acid_added_ml != null || row.alkali_added_ml != null">
                  {{ row.acid_added_ml ?? '-' }} / {{ row.alkali_added_ml ?? '-' }}
                </template>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="note" label="备注" min-width="150">
              <template #default="{ row }">{{ row.note || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.operated_at) }}</template>
            </el-table-column>
          </el-table>
          <div class="pagination-container" v-if="solutionsTotal > 0">
            <el-pagination
              v-model:current-page="solutionsPagination.page"
              v-model:page-size="solutionsPagination.pageSize"
              :total="solutionsTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchSolutionChanges"
              @current-change="fetchSolutionChanges"
            />
          </div>
          <el-empty v-if="!solutionsLoading && solutionChanges.length === 0" description="暂无溶液更换记录" />
        </el-tab-pane>

        <!-- Tab 3: 离子检测 -->
        <el-tab-pane label="离子检测" name="iontests">
          <el-table :data="ionTests" v-loading="ionTestsLoading" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="sample_code" label="样品编号" width="140" />
            <el-table-column label="检测方法" width="90">
              <template #default="{ row }">{{ testMethodLabel(row.test_method) }}</template>
            </el-table-column>
            <el-table-column label="大量元素" min-width="240">
              <template #default="{ row }">
                <div class="ion-compact" v-if="hasMacroIons(row)">
                  <span v-if="row.no3_n != null">NO₃-N:{{ row.no3_n }}</span>
                  <span v-if="row.nh4_n != null">NH₄-N:{{ row.nh4_n }}</span>
                  <span v-if="row.p != null">P:{{ row.p }}</span>
                  <span v-if="row.k != null">K:{{ row.k }}</span>
                  <span v-if="row.ca != null">Ca:{{ row.ca }}</span>
                  <span v-if="row.mg != null">Mg:{{ row.mg }}</span>
                  <span v-if="row.s != null">S:{{ row.s }}</span>
                </div>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="EC/pH" width="110">
              <template #default="{ row }">
                <template v-if="row.ec_at_sample != null || row.ph_at_sample != null">
                  {{ row.ec_at_sample ?? '-' }} / {{ row.ph_at_sample ?? '-' }}
                </template>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="采样时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.sampled_at) }}</template>
            </el-table-column>
          </el-table>
          <div class="pagination-container" v-if="ionTestsTotal > 0">
            <el-pagination
              v-model:current-page="ionTestsPagination.page"
              v-model:page-size="ionTestsPagination.pageSize"
              :total="ionTestsTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchIonTests"
              @current-change="fetchIonTests"
            />
          </div>
          <el-empty v-if="!ionTestsLoading && ionTests.length === 0" description="暂无离子检测记录" />
        </el-tab-pane>

        <!-- Tab 4: 浓缩液消耗 -->
        <el-tab-pane label="浓缩液消耗" name="concentrates">
          <el-table :data="concentrateLogs" v-loading="concentratesLoading" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column label="浓缩液" min-width="160">
              <template #default="{ row }">{{ concentrateLabel(row.inventory_id) }}</template>
            </el-table-column>
            <el-table-column prop="volume_used_ml" label="用量 (ml)" width="110" />
            <el-table-column label="使用时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.used_at) }}</template>
            </el-table-column>
            <el-table-column label="关联换液" width="90">
              <template #default="{ row }">{{ row.solution_change_id ?? '-' }}</template>
            </el-table-column>
          </el-table>
          <div class="pagination-container" v-if="concentratesTotal > 0">
            <el-pagination
              v-model:current-page="concentratesPagination.page"
              v-model:page-size="concentratesPagination.pageSize"
              :total="concentratesTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchConcentrateLogs"
              @current-change="fetchConcentrateLogs"
            />
          </div>
          <el-empty v-if="!concentratesLoading && concentrateLogs.length === 0" description="暂无浓缩液消耗记录" />
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else description="液槽不存在" />

    <!-- Edit Dialog -->
    <el-dialog v-model="editDialogVisible" title="编辑液槽" width="500px">
      <el-form ref="editFormRef" :model="editForm" :rules="editFormRules" label-width="120px">
        <el-form-item label="种植区" prop="growing_zone_id">
          <el-select v-model="editForm.growing_zone_id" placeholder="选择种植区" filterable style="width: 100%" disabled>
            <el-option v-for="z in zones" :key="z.id" :label="`${z.name} (${z.code})`" :value="z.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="总容积 (L)" prop="total_volume_liter">
          <el-input-number v-model="editForm.total_volume_liter" :min="0" :precision="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 100%">
            <el-option label="使用中" value="ACTIVE" />
            <el-option label="维护中" value="INACTIVE" />
            <el-option label="空置" value="EMPTY" />
          </el-select>
        </el-form-item>
        <el-divider content-position="left">传感器绑定</el-divider>
        <el-form-item label="EC 传感器">
          <el-select v-model="editForm.ec_sensor_channel_id" placeholder="选择 EC 传感器通道" clearable filterable style="width: 100%">
            <el-option v-for="ch in sensorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.metric_code})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="pH 传感器">
          <el-select v-model="editForm.ph_sensor_channel_id" placeholder="选择 pH 传感器通道" clearable filterable style="width: 100%">
            <el-option v-for="ch in sensorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.metric_code})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="液位传感器">
          <el-select v-model="editForm.level_sensor_channel_id" placeholder="选择液位传感器通道" clearable filterable style="width: 100%">
            <el-option v-for="ch in sensorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.metric_code})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="水温传感器">
          <el-select v-model="editForm.temp_sensor_channel_id" placeholder="选择水温传感器通道" clearable filterable style="width: 100%">
            <el-option v-for="ch in sensorChannels" :key="ch.id" :label="`${ch.channel_code} (${ch.metric_code})`" :value="ch.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSubmitting" @click="handleEditSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft, Edit, Delete } from '@element-plus/icons-vue'
import { nutrientApi, greenhouseApi, deviceApi, telemetryApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, fallbackIdLabel, growingZoneLabel } from '@/utils/labels'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type {
  NutrientTank, SolutionChangeEvent, IonTestRecord, ConcentrateUsageLog,
  GrowingZone, SensorChannel
} from '@/types'
import { Role } from '@/types'

const route = useRoute()
const router = useRouter()

const tankId = computed(() => Number(route.params.id))
const { hasRole, canControlDevice } = usePermission()
const canManage = computed(() => canControlDevice())
const canDelete = computed(() => hasRole(Role.ADMIN))

const loading = ref(false)
const errorMsg = ref('')
const tank = ref<NutrientTank | null>(null)
const activeTab = ref('info')

// Sensor polling
const sensorData = reactive({
  ec: null as { value: number; quality_flag: string; collected_at: string } | null,
  ph: null as { value: number; quality_flag: string; collected_at: string } | null,
  level: null as { value: number; quality_flag: string; collected_at: string } | null,
  temp: null as { value: number; quality_flag: string; collected_at: string } | null
})
let sensorPollTimer: ReturnType<typeof setInterval> | null = null

// Solution changes
const solutionChanges = ref<SolutionChangeEvent[]>([])
const solutionsLoading = ref(false)
const solutionsTotal = ref(0)
const solutionsPagination = reactive({ page: 1, pageSize: 20 })

// Ion tests
const ionTests = ref<IonTestRecord[]>([])
const ionTestsLoading = ref(false)
const ionTestsTotal = ref(0)
const ionTestsPagination = reactive({ page: 1, pageSize: 20 })

// Concentrate logs
const concentrateLogs = ref<ConcentrateUsageLog[]>([])
const concentratesLoading = ref(false)
const concentratesTotal = ref(0)
const concentratesPagination = reactive({ page: 1, pageSize: 20 })

// Lookup data
const zones = ref<GrowingZone[]>([])
const sensorChannels = ref<SensorChannel[]>([])

const growingZoneLabelById = computed(() =>
  buildIdLabelMap(zones.value, z => z.id, growingZoneLabel, '种植区')
)

function growingZoneName(zoneId?: number) {
  if (!zoneId) return '-'
  return growingZoneLabelById.value[zoneId] || fallbackIdLabel('种植区', zoneId)
}

const sensorBindingCount = computed(() => {
  if (!tank.value) return 0
  let count = 0
  if (tank.value.ec_sensor_channel_id) count++
  if (tank.value.ph_sensor_channel_id) count++
  if (tank.value.level_sensor_channel_id) count++
  if (tank.value.temp_sensor_channel_id) count++
  return count
})

const tankStatusLabel = computed(() => {
  const map: Record<string, string> = { ACTIVE: '使用中', INACTIVE: '维护中', EMPTY: '空置' }
  return map[tank.value?.status ?? ''] || tank.value?.status || '-'
})

const tankStatusTagType = computed(() => {
  const map: Record<string, string> = { ACTIVE: 'success', INACTIVE: 'warning', EMPTY: 'info' }
  return map[tank.value?.status ?? ''] || 'info'
})

function changeTypeLabel(type: string) {
  const map: Record<string, string> = { FULL_REPLACE: '完全更换', PARTIAL_REFRESH: '部分更换', TOP_UP: '补水' }
  return map[type] || type
}

function testMethodLabel(method: string) {
  const map: Record<string, string> = { LAB: '实验室', STRIP: '试纸', METER: '仪器' }
  return map[method] || method
}

function hasMacroIons(row: IonTestRecord) {
  return row.no3_n != null || row.nh4_n != null || row.p != null || row.k != null || row.ca != null || row.mg != null || row.s != null
}

function concentrateLabel(inventoryId: number) {
  return `浓缩液#${inventoryId}`
}

// Sensor Card CSS
function sensorCardClass(field: string) {
  const key = field as keyof typeof sensorData
  const val = sensorData[key]
  if (!val) return 'card-empty'
  if (val.quality_flag === 'normal') return 'card-normal'
  if (val.quality_flag === 'out_of_range') return 'card-danger'
  return 'card-warn'
}

// Data loading
async function loadReferenceData() {
  try {
    const [zonesResult, channelsResult] = await Promise.allSettled([
      greenhouseApi.getGrowingZones({ page_size: LARGE_PAGE_SIZE }),
      deviceApi.getSensorChannels({ page_size: LARGE_PAGE_SIZE })
    ])
    if (zonesResult.status === 'fulfilled') zones.value = zonesResult.value.items
    if (channelsResult.status === 'fulfilled') sensorChannels.value = channelsResult.value.items
  } catch { /* ignore */ }
}

async function loadTank() {
  if (!tankId.value) {
    errorMsg.value = '无效的液槽 ID'
    return
  }
  try {
    tank.value = await nutrientApi.getNutrientTank(tankId.value)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载液槽信息失败'
    errorMsg.value = msg
  }
}

// Sensor polling
function getBoundChannelIds(): number[] {
  if (!tank.value) return []
  const ids: number[] = []
  if (tank.value.ec_sensor_channel_id) ids.push(tank.value.ec_sensor_channel_id)
  if (tank.value.ph_sensor_channel_id) ids.push(tank.value.ph_sensor_channel_id)
  if (tank.value.level_sensor_channel_id) ids.push(tank.value.level_sensor_channel_id)
  if (tank.value.temp_sensor_channel_id) ids.push(tank.value.temp_sensor_channel_id)
  return ids
}

async function fetchSensorData() {
  const ids = getBoundChannelIds()
  if (ids.length === 0) return
  try {
    const data = await telemetryApi.getChannelsLatest(ids)
    const fieldMap: Record<number, keyof typeof sensorData> = {}
    if (tank.value) {
      if (tank.value.ec_sensor_channel_id) fieldMap[tank.value.ec_sensor_channel_id] = 'ec'
      if (tank.value.ph_sensor_channel_id) fieldMap[tank.value.ph_sensor_channel_id] = 'ph'
      if (tank.value.level_sensor_channel_id) fieldMap[tank.value.level_sensor_channel_id] = 'level'
      if (tank.value.temp_sensor_channel_id) fieldMap[tank.value.temp_sensor_channel_id] = 'temp'
    }
    // Reset all
    sensorData.ec = null; sensorData.ph = null; sensorData.level = null; sensorData.temp = null
    for (const [channelIdStr, value] of Object.entries(data)) {
      const cid = Number(channelIdStr)
      const field = fieldMap[cid]
      if (field && value) {
        sensorData[field] = value as typeof sensorData[typeof field]
      }
    }
  } catch { /* ignore */ }
}

function startSensorPolling() {
  fetchSensorData()
  sensorPollTimer = setInterval(fetchSensorData, 15000)
}

function stopSensorPolling() {
  if (sensorPollTimer) {
    clearInterval(sensorPollTimer)
    sensorPollTimer = null
  }
}

// Sub-entity loading
async function fetchSolutionChanges() {
  if (!tankId.value) return
  solutionsLoading.value = true
  try {
    const data = await nutrientApi.getSolutionChanges({
      tank_id: tankId.value,
      page: solutionsPagination.page,
      page_size: solutionsPagination.pageSize
    })
    solutionChanges.value = data.items
    solutionsTotal.value = data.total
  } catch {
    solutionChanges.value = []
    solutionsTotal.value = 0
  } finally {
    solutionsLoading.value = false
  }
}

async function fetchIonTests() {
  if (!tankId.value) return
  ionTestsLoading.value = true
  try {
    const data = await nutrientApi.getIonTests({
      tank_id: tankId.value,
      page: ionTestsPagination.page,
      page_size: ionTestsPagination.pageSize
    })
    ionTests.value = data.items
    ionTestsTotal.value = data.total
  } catch {
    ionTests.value = []
    ionTestsTotal.value = 0
  } finally {
    ionTestsLoading.value = false
  }
}

async function fetchConcentrateLogs() {
  if (!tankId.value) return
  concentratesLoading.value = true
  try {
    const data = await nutrientApi.getConcentrateUsageLogs({
      tank_id: tankId.value,
      page: concentratesPagination.page,
      page_size: concentratesPagination.pageSize
    })
    concentrateLogs.value = data.items
    concentratesTotal.value = data.total
  } catch {
    concentrateLogs.value = []
    concentratesTotal.value = 0
  } finally {
    concentratesLoading.value = false
  }
}

async function loadAllData() {
  errorMsg.value = ''
  loading.value = true
  await Promise.allSettled([loadTank(), loadReferenceData()])
  loading.value = false
  startSensorPolling()
}

// Edit dialog
const editDialogVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editSubmitting = ref(false)

const editForm = reactive({
  growing_zone_id: undefined as number | undefined,
  total_volume_liter: 0,
  status: 'ACTIVE' as string,
  ec_sensor_channel_id: undefined as number | undefined,
  ph_sensor_channel_id: undefined as number | undefined,
  level_sensor_channel_id: undefined as number | undefined,
  temp_sensor_channel_id: undefined as number | undefined
})

const editFormRules: FormRules = {
  total_volume_liter: [{ required: true, message: '请输入总容积', trigger: 'blur' }]
}

function openEditDialog() {
  if (!tank.value) return
  editForm.growing_zone_id = tank.value.growing_zone_id
  editForm.total_volume_liter = tank.value.total_volume_liter
  editForm.status = tank.value.status
  editForm.ec_sensor_channel_id = tank.value.ec_sensor_channel_id ?? undefined
  editForm.ph_sensor_channel_id = tank.value.ph_sensor_channel_id ?? undefined
  editForm.level_sensor_channel_id = tank.value.level_sensor_channel_id ?? undefined
  editForm.temp_sensor_channel_id = tank.value.temp_sensor_channel_id ?? undefined
  editDialogVisible.value = true
}

async function handleEditSubmit() {
  if (!editFormRef.value) return
  try {
    await editFormRef.value.validate()
  } catch { return }
  if (!tankId.value) return

  editSubmitting.value = true
  try {
    await nutrientApi.updateNutrientTank(tankId.value, {
      total_volume_liter: editForm.total_volume_liter,
      ec_sensor_channel_id: editForm.ec_sensor_channel_id ?? undefined,
      ph_sensor_channel_id: editForm.ph_sensor_channel_id ?? undefined,
      level_sensor_channel_id: editForm.level_sensor_channel_id ?? undefined,
      temp_sensor_channel_id: editForm.temp_sensor_channel_id ?? undefined
    } as never)
    ElMessage.success('液槽已更新')
    editDialogVisible.value = false
    await loadTank()
    fetchSensorData()
  } catch { /* handled by interceptor */ }
  finally { editSubmitting.value = false }
}

async function handleDelete() {
  await ElMessageBox.confirm('确认删除该液槽？', '提示', { type: 'warning' })
  try {
    await nutrientApi.deleteNutrientTank(tankId.value)
    ElMessage.success('已删除')
    router.push('/nutrient/tanks')
  } catch { /* handled by interceptor */ }
}

onMounted(() => {
  loadAllData()
})

onBeforeUnmount(() => {
  stopSensorPolling()
})
</script>

<style scoped lang="scss">
.nutrient-tank-detail-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
      .page-title {
        font-size: 22px;
        font-weight: 700;
        color: var(--color-text-primary);
        margin: 0;
      }
    }
    .header-right { display: flex; gap: 8px; }
  }
  .loading-container {
    padding: 40px;
    background: var(--bg-card);
    border-radius: var(--radius-md);
  }
  .info-card { margin-bottom: 16px; }
  .card-title { font-size: 15px; font-weight: 600; }

  // Sensor grid
  .sensor-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
  }
  .monitor-card {
    padding: 16px;
    border-radius: var(--radius-md);
    border-left: 4px solid #ccc;
    background: var(--bg-page);
    .sensor-label { font-size: 13px; color: var(--color-text-secondary); margin-bottom: 8px; }
    .sensor-value { font-size: 28px; font-weight: 700; .sensor-unit { font-size: 14px; font-weight: 400; color: var(--color-text-secondary); } }
    .sensor-time { font-size: 12px; color: var(--color-text-secondary); margin-top: 4px; }
    .sensor-empty { font-size: 14px; color: var(--color-text-secondary); }
  }
  .card-normal { border-left-color: var(--color-success); .sensor-value { color: var(--color-success); } }
  .card-danger { border-left-color: var(--color-danger); .sensor-value { color: var(--color-danger); } }
  .card-warn { border-left-color: var(--color-warning); .sensor-value { color: var(--color-warning); } }
  .card-empty { border-left-style: dashed; }

  .ion-compact {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    span { font-size: 12px; }
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--spacing-md);
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }
}
</style>
