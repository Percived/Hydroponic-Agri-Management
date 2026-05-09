<template>
  <div class="pest-observations-page">
    <div class="page-header">
      <h1 class="page-title">病虫害观察</h1>
      <el-button type="primary" @click="openCreateObservation">
        <el-icon><Plus /></el-icon>
        新增观察
      </el-button>
    </div>

    <div class="filter-section">
      <el-select v-model="filters.greenhouse_id" placeholder="选择温室" clearable filterable style="width: 180px">
        <el-option v-for="g in greenhouses" :key="g.id" :label="`${g.name} (ID:${g.id})`" :value="g.id" />
      </el-select>
      <el-select v-model="filters.batch_id" placeholder="选择批次" clearable filterable style="width: 180px">
        <el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (ID:${b.id})`" :value="b.id" />
      </el-select>
      <el-select v-model="filters.severity" placeholder="严重程度" clearable style="width: 140px">
        <el-option label="轻度" value="LIGHT" />
        <el-option label="中度" value="MODERATE" />
        <el-option label="重度" value="SEVERE" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="observations" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="温室" width="160">
          <template #default="{ row }">{{ greenhouseName(row.greenhouse_id) }}</template>
        </el-table-column>
        <el-table-column label="种植区" width="160">
          <template #default="{ row }">{{ growingZoneName(row.growing_zone_id) }}</template>
        </el-table-column>
        <el-table-column label="批次" width="160">
          <template #default="{ row }">{{ batchName(row.batch_id) }}</template>
        </el-table-column>
        <el-table-column prop="pest_or_disease" label="病虫害" width="140" />
        <el-table-column prop="severity" label="严重程度" width="110">
          <template #default="{ row }">
            <el-tag :type="severityTagType(row.severity)">{{ severityName(row.severity) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="受害面积%" width="120">
          <template #default="{ row }">{{ row.affected_area_pct ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="受害株数" width="120">
          <template #default="{ row }">{{ row.affected_plant_count ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="symptoms" label="症状" min-width="180">
          <template #default="{ row }">{{ row.symptoms || '-' }}</template>
        </el-table-column>
        <el-table-column prop="observed_at" label="观察时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.observed_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditObservation(row)">编辑</el-button>
            <el-button type="success" link @click="openTreatmentsDialog(row)">治疗</el-button>
            <el-button type="danger" link @click="removeObservation(row.id)">删除</el-button>
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

    <!-- 观察新增/编辑弹窗 -->
    <el-dialog v-model="obsDialogVisible" :title="isEditObs ? '编辑观察记录' : '新增观察记录'" width="600px">
      <el-form ref="obsFormRef" :model="obsForm" :rules="obsFormRules" label-width="120px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="温室" prop="greenhouse_id">
              <el-select v-model="obsForm.greenhouse_id" placeholder="选择温室" filterable style="width: 100%" @change="onObsGreenhouseChange">
                <el-option v-for="g in greenhouses" :key="g.id" :label="`${g.name} (ID:${g.id})`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="种植区">
              <el-select v-model="obsForm.growing_zone_id" placeholder="选择种植区" clearable filterable style="width: 100%">
                <el-option v-for="z in filteredGrowingZones" :key="z.id" :label="`${z.name} (ID:${z.id})`" :value="z.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="批次">
              <el-select v-model="obsForm.batch_id" placeholder="选择批次" clearable filterable style="width: 100%">
                <el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (ID:${b.id})`" :value="b.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="严重程度" prop="severity">
              <el-select v-model="obsForm.severity" style="width: 100%">
                <el-option label="轻度" value="LIGHT" />
                <el-option label="中度" value="MODERATE" />
                <el-option label="重度" value="SEVERE" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="病虫害" prop="pest_or_disease">
              <el-input v-model="obsForm.pest_or_disease" placeholder="描述病虫害种类" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="观察时间" prop="observed_at">
              <el-date-picker
                v-model="obsForm.observed_at"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="受害面积%">
              <el-input v-model="obsForm.affected_area_pct" placeholder="如 15.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="受害株数">
              <el-input v-model="obsForm.affected_plant_count" placeholder="如 50" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="症状描述">
              <el-input v-model="obsForm.symptoms" type="textarea" :rows="2" placeholder="症状描述" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="obsDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="obsSubmitLoading" @click="handleObsSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 治疗记录弹窗（嵌套） -->
    <el-dialog v-model="treatDialogVisible" title="治疗记录" width="700px">
      <div style="margin-bottom: 12px">
        <el-button type="primary" size="small" @click="openCreateTreatment">新增治疗</el-button>
      </div>
      <el-table :data="treatments" v-loading="treatLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="treatment_type" label="类型" width="100">
          <template #default="{ row }">{{ treatTypeName(row.treatment_type) }}</template>
        </el-table-column>
        <el-table-column prop="product_name" label="产品" width="140" />
        <el-table-column prop="dosage" label="用量" width="100" />
        <el-table-column prop="application_method" label="施用方法" width="100">
          <template #default="{ row }">{{ appMethodName(row.application_method) }}</template>
        </el-table-column>
        <el-table-column label="安全间隔(天)" width="120">
          <template #default="{ row }">{{ row.safety_interval_days ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="treated_at" label="处理时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.treated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link size="small" @click="removeTreatment(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="treatDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 治疗新增弹窗 -->
    <el-dialog v-model="treatFormVisible" title="新增治疗记录" width="550px">
      <el-form ref="treatFormRef" :model="treatForm" :rules="treatFormRules" label-width="120px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="温室" prop="greenhouse_id">
              <el-select v-model="treatForm.greenhouse_id" placeholder="选择温室" filterable style="width: 100%">
                <el-option v-for="g in greenhouses" :key="g.id" :label="`${g.name} (ID:${g.id})`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="治疗类型" prop="treatment_type">
              <el-select v-model="treatForm.treatment_type" style="width: 100%">
                <el-option label="化学" value="CHEMICAL" />
                <el-option label="生物" value="BIOLOGICAL" />
                <el-option label="物理" value="PHYSICAL" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="产品名称" prop="product_name">
              <el-input v-model="treatForm.product_name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="有效成分">
              <el-input v-model="treatForm.active_ingredient" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="用量" prop="dosage">
              <el-input v-model="treatForm.dosage" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="施用方法" prop="application_method">
              <el-select v-model="treatForm.application_method" style="width: 100%">
                <el-option label="喷雾" value="SPRAY" />
                <el-option label="灌根" value="DRENCH" />
                <el-option label="熏蒸" value="FOG" />
                <el-option label="释放" value="RELEASE" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="安全间隔(天)">
              <el-input-number v-model="treatForm.safety_interval_days" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="再进入间隔(h)">
              <el-input-number v-model="treatForm.reentry_interval_hours" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="处理时间" prop="treated_at">
              <el-date-picker
                v-model="treatForm.treated_at"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="treatForm.note" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="treatFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="treatSubmitLoading" @click="handleTreatSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { pestApi, greenhouseApi, cropApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, cropBatchLabel, fallbackIdLabel, greenhouseLabel, growingZoneLabel } from '@/utils/labels'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { PestDiseaseObservation, TreatmentRecord, Greenhouse, CropBatch, GrowingZone } from '@/types'

// ── Observations ──
const loading = ref(false)
const observations = ref<PestDiseaseObservation[]>([])
const total = ref(0)
const greenhouses = ref<Greenhouse[]>([])
const batches = ref<CropBatch[]>([])
const growingZones = ref<GrowingZone[]>([])

const greenhouseLabelById = computed(() =>
  buildIdLabelMap(greenhouses.value, g => g.id, greenhouseLabel, '温室')
)
const batchLabelById = computed(() =>
  buildIdLabelMap(batches.value, b => b.id, cropBatchLabel, '批次')
)
const growingZoneLabelById = computed(() =>
  buildIdLabelMap(growingZones.value, z => z.id, growingZoneLabel, '种植区')
)

function greenhouseName(greenhouseId?: number) {
  if (!greenhouseId) return fallbackIdLabel('温室', greenhouseId)
  return greenhouseLabelById.value[greenhouseId] || fallbackIdLabel('温室', greenhouseId)
}

function batchName(batchId?: number) {
  if (!batchId) return fallbackIdLabel('批次', batchId)
  return batchLabelById.value[batchId] || fallbackIdLabel('批次', batchId)
}

function growingZoneName(zoneId?: number) {
  if (!zoneId) return fallbackIdLabel('种植区', zoneId)
  return growingZoneLabelById.value[zoneId] || fallbackIdLabel('种植区', zoneId)
}

const filteredGrowingZones = computed(() => {
  if (!obsForm.greenhouse_id) return growingZones.value
  return growingZones.value.filter(z => z.greenhouse_id === obsForm.greenhouse_id)
})

function onObsGreenhouseChange() {
  obsForm.growing_zone_id = undefined
}

const filters = reactive({
  greenhouse_id: undefined as number | undefined,
  batch_id: undefined as number | undefined,
  severity: '' as string
})

const pagination = reactive({ page: 1, pageSize: 20 })

const obsDialogVisible = ref(false)
const isEditObs = ref(false)
const obsFormRef = ref<FormInstance>()
const obsSubmitLoading = ref(false)
const editingObsId = ref<number | null>(null)

const emptyObsForm = () => ({
  greenhouse_id: undefined as number | undefined,
  growing_zone_id: undefined as number | undefined,
  batch_id: undefined as number | undefined,
  observed_at: '',
  pest_or_disease: '',
  severity: 'LIGHT' as string,
  affected_area_pct: undefined as number | undefined,
  affected_plant_count: undefined as number | undefined,
  symptoms: '' as string
})

const obsForm = reactive(emptyObsForm())

const obsFormRules: FormRules = {
  greenhouse_id: [{ required: true, message: '请输入温室ID', trigger: 'blur' }],
  pest_or_disease: [{ required: true, message: '请输入病虫害种类', trigger: 'blur' }],
  severity: [{ required: true, message: '请选择严重程度', trigger: 'change' }],
  observed_at: [{ required: true, message: '请选择观察时间', trigger: 'change' }]
}

function severityName(s: string) {
  const map: Record<string, string> = { LIGHT: '轻度', MODERATE: '中度', SEVERE: '重度' }
  return map[s] || s
}

function severityTagType(s: string) {
  const map: Record<string, string> = { LIGHT: 'warning', MODERATE: '', SEVERE: 'danger' }
  return map[s] || 'info'
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.greenhouse_id) params.greenhouse_id = filters.greenhouse_id
    if (filters.batch_id) params.batch_id = filters.batch_id
    if (filters.severity) params.severity = filters.severity
    const data = await pestApi.getPestObservations(params)
    observations.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.greenhouse_id = undefined
  filters.batch_id = undefined
  filters.severity = ''
  pagination.page = 1
  fetchData()
}

function openCreateObservation() {
  isEditObs.value = false
  editingObsId.value = null
  Object.assign(obsForm, emptyObsForm())
  obsDialogVisible.value = true
}

function openEditObservation(obs: PestDiseaseObservation) {
  isEditObs.value = true
  editingObsId.value = obs.id
  Object.assign(obsForm, {
    greenhouse_id: obs.greenhouse_id,
    growing_zone_id: obs.growing_zone_id,
    batch_id: obs.batch_id,
    observed_at: obs.observed_at,
    pest_or_disease: obs.pest_or_disease,
    severity: obs.severity,
    affected_area_pct: obs.affected_area_pct,
    affected_plant_count: obs.affected_plant_count,
    symptoms: obs.symptoms || ''
  })
  obsDialogVisible.value = true
}

async function handleObsSubmit() {
  if (!obsFormRef.value) return
  try {
    await obsFormRef.value.validate()
  } catch {
    return
  }

  obsSubmitLoading.value = true
  try {
    const payload = {
      greenhouse_id: obsForm.greenhouse_id!,
      growing_zone_id: obsForm.growing_zone_id || undefined,
      batch_id: obsForm.batch_id || undefined,
      observed_at: obsForm.observed_at,
      pest_or_disease: obsForm.pest_or_disease,
      severity: obsForm.severity,
      affected_area_pct: obsForm.affected_area_pct ? Number(obsForm.affected_area_pct) : undefined,
      affected_plant_count: obsForm.affected_plant_count ? Number(obsForm.affected_plant_count) : undefined,
      symptoms: obsForm.symptoms || undefined
    }
    if (isEditObs.value && editingObsId.value) {
      await pestApi.updatePestObservation(editingObsId.value, payload)
      ElMessage.success('观察记录已更新')
    } else {
      await pestApi.createPestObservation(payload)
      ElMessage.success('观察记录已创建')
    }
    obsDialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    obsSubmitLoading.value = false
  }
}

async function removeObservation(id: number) {
  await ElMessageBox.confirm('确认删除该观察记录？', '提示', { type: 'warning' })
  await pestApi.deletePestObservation(id)
  ElMessage.success('已删除')
  fetchData()
}

// ── Treatments (nested) ──
const treatDialogVisible = ref(false)
const treatLoading = ref(false)
const treatments = ref<TreatmentRecord[]>([])
const currentObsId = ref<number | null>(null)

const treatFormVisible = ref(false)
const treatFormRef = ref<FormInstance>()
const treatSubmitLoading = ref(false)

const emptyTreatForm = () => ({
  greenhouse_id: undefined as number | undefined,
  treatment_type: 'CHEMICAL' as string,
  product_name: '',
  active_ingredient: '' as string,
  dosage: '',
  application_method: 'SPRAY' as string,
  safety_interval_days: undefined as number | undefined,
  reentry_interval_hours: undefined as number | undefined,
  treated_at: '',
  note: '' as string
})

const treatForm = reactive(emptyTreatForm())

const treatFormRules: FormRules = {
  greenhouse_id: [{ required: true, message: '请输入温室ID', trigger: 'blur' }],
  treatment_type: [{ required: true, message: '请选择治疗类型', trigger: 'change' }],
  product_name: [{ required: true, message: '请输入产品名称', trigger: 'blur' }],
  dosage: [{ required: true, message: '请输入用量', trigger: 'blur' }],
  application_method: [{ required: true, message: '请选择施用方法', trigger: 'change' }],
  treated_at: [{ required: true, message: '请选择处理时间', trigger: 'change' }]
}

function treatTypeName(t: string) {
  const map: Record<string, string> = { CHEMICAL: '化学', BIOLOGICAL: '生物', PHYSICAL: '物理' }
  return map[t] || t
}

function appMethodName(m: string) {
  const map: Record<string, string> = { SPRAY: '喷雾', DRENCH: '灌根', FOG: '熏蒸', RELEASE: '释放' }
  return map[m] || m
}

async function openTreatmentsDialog(obs: PestDiseaseObservation) {
  currentObsId.value = obs.id
  treatDialogVisible.value = true
  treatLoading.value = true
  try {
    const data = await pestApi.getObservationTreatments(obs.id)
    treatments.value = data.items
  } catch {
    treatments.value = []
  } finally {
    treatLoading.value = false
  }
}

function openCreateTreatment() {
  Object.assign(treatForm, {
    ...emptyTreatForm(),
    greenhouse_id: filters.greenhouse_id
  })
  treatFormVisible.value = true
}

async function handleTreatSubmit() {
  if (!treatFormRef.value) return
  try {
    await treatFormRef.value.validate()
  } catch {
    return
  }

  treatSubmitLoading.value = true
  try {
    await pestApi.createTreatmentRecord({
      observation_id: currentObsId.value || undefined,
      greenhouse_id: treatForm.greenhouse_id!,
      treatment_type: treatForm.treatment_type,
      product_name: treatForm.product_name,
      active_ingredient: treatForm.active_ingredient || undefined,
      dosage: treatForm.dosage,
      application_method: treatForm.application_method,
      safety_interval_days: treatForm.safety_interval_days,
      reentry_interval_hours: treatForm.reentry_interval_hours,
      treated_at: treatForm.treated_at,
      note: treatForm.note || undefined
    })
    ElMessage.success('治疗记录已创建')
    treatFormVisible.value = false
    // Refresh treatments list
    if (currentObsId.value) {
      treatLoading.value = true
      try {
        const data = await pestApi.getObservationTreatments(currentObsId.value)
        treatments.value = data.items
      } catch {
        treatments.value = []
      } finally {
        treatLoading.value = false
      }
    }
  } catch {
    // handled by interceptor
  } finally {
    treatSubmitLoading.value = false
  }
}

async function removeTreatment(id: number) {
  await ElMessageBox.confirm('确认删除该治疗记录？', '提示', { type: 'warning' })
  await pestApi.deleteTreatmentRecord(id)
  ElMessage.success('已删除')
  // Refresh
  if (currentObsId.value) {
    treatLoading.value = true
    try {
      const data = await pestApi.getObservationTreatments(currentObsId.value)
      treatments.value = data.items
    } catch {
      treatments.value = []
    } finally {
      treatLoading.value = false
    }
  }
}

async function loadGreenhouses() {
  try {
    const data = await greenhouseApi.getGreenhouses({ page_size: LARGE_PAGE_SIZE })
    greenhouses.value = data.items
  } catch { /* ignore */ }
}

async function loadBatches() {
  try {
    const data = await cropApi.getBatches({ page_size: LARGE_PAGE_SIZE })
    batches.value = data.items
  } catch { /* ignore */ }
}

async function loadGrowingZones() {
  try {
    const data = await greenhouseApi.getGrowingZones({ page_size: LARGE_PAGE_SIZE })
    growingZones.value = data.items
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchData()
  loadGreenhouses()
  loadBatches()
  loadGrowingZones()
})
</script>

<style scoped lang="scss">
.pest-observations-page {
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
}
</style>
