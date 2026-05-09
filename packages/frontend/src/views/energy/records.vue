<template>
  <div class="energy-records-page">
    <div class="page-header">
      <h1 class="page-title">能耗记录</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增记录
      </el-button>
    </div>

    <!-- 汇总卡片 -->
    <div class="summary-cards" v-if="summaryItems.length">
      <div class="summary-card" v-for="item in summaryItems" :key="item.record_type">
        <div class="summary-label">{{ recordTypeName(item.record_type) }}</div>
        <div class="summary-value">{{ item.total_consumption }} {{ item.unit }}</div>
      </div>
    </div>

    <!-- 筛选区 -->
    <div class="filter-section">
      <el-select v-model="filters.record_type" placeholder="能耗类型" clearable style="width: 160px">
        <el-option label="电力" value="ELECTRICITY" />
        <el-option label="水" value="WATER" />
        <el-option label="CO2" value="CO2_GAS" />
      </el-select>
      <el-select v-model="filters.greenhouse_id" placeholder="选择温室" clearable filterable style="width: 180px">
        <el-option v-for="g in greenhouses" :key="g.id" :label="`${g.name} (ID:${g.id})`" :value="g.id" />
      </el-select>
      <el-date-picker
        v-model="filters.start_date"
        type="date"
        value-format="YYYY-MM-DD"
        placeholder="开始日期"
        style="width: 160px"
      />
      <el-date-picker
        v-model="filters.end_date"
        type="date"
        value-format="YYYY-MM-DD"
        placeholder="结束日期"
        style="width: 160px"
      />
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <!-- 表格 -->
    <div class="table-container">
      <el-table :data="records" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="温室" width="160">
          <template #default="{ row }">{{ greenhouseName(row.greenhouse_id) }}</template>
        </el-table-column>
        <el-table-column prop="record_type" label="能耗类型" width="120">
          <template #default="{ row }">{{ recordTypeName(row.record_type) }}</template>
        </el-table-column>
        <el-table-column prop="consumption_value" label="消耗量" width="120">
          <template #default="{ row }">{{ row.consumption_value }} {{ row.unit }}</template>
        </el-table-column>
        <el-table-column prop="record_period_start" label="周期开始" width="180">
          <template #default="{ row }">{{ formatDateTime(row.record_period_start) }}</template>
        </el-table-column>
        <el-table-column prop="record_period_end" label="周期结束" width="180">
          <template #default="{ row }">{{ formatDateTime(row.record_period_end) }}</template>
        </el-table-column>
        <el-table-column label="批次" width="160">
          <template #default="{ row }">{{ batchName(row.batch_id) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="removeRecord(row.id)">删除</el-button>
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

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑能耗记录' : '新增能耗记录'" width="550px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="110px">
        <el-form-item label="温室" prop="greenhouse_id">
          <el-select v-model="formData.greenhouse_id" placeholder="选择温室" filterable style="width: 100%">
            <el-option v-for="g in greenhouses" :key="g.id" :label="`${g.name} (ID:${g.id})`" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="能耗类型" prop="record_type">
          <el-select v-model="formData.record_type" style="width: 100%">
            <el-option label="电力" value="ELECTRICITY" />
            <el-option label="水" value="WATER" />
            <el-option label="CO2" value="CO2_GAS" />
          </el-select>
        </el-form-item>
        <el-form-item label="消耗量" prop="consumption_value">
          <el-input-number v-model="formData.consumption_value" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="单位" prop="unit">
          <el-select v-model="formData.unit" style="width: 100%">
            <el-option label="kWh" value="kWh" />
            <el-option label="m³" value="m³" />
            <el-option label="kg" value="kg" />
            <el-option label="L" value="L" />
          </el-select>
        </el-form-item>
        <el-form-item label="周期开始" prop="record_period_start">
          <el-date-picker
            v-model="formData.record_period_start"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="周期结束" prop="record_period_end">
          <el-date-picker
            v-model="formData.record_period_end"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="批次">
          <el-select v-model="formData.batch_id" placeholder="选择批次" clearable filterable style="width: 100%">
            <el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (ID:${b.id})`" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="抄表起始">
          <el-input-number v-model="formData.meter_reading_start" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="抄表结束">
          <el-input-number v-model="formData.meter_reading_end" :min="0" :precision="2" style="width: 100%" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { energyApi, greenhouseApi, cropApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, cropBatchLabel, fallbackIdLabel, greenhouseLabel } from '@/utils/labels'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { EnergyConsumptionRecord, EnergySummary, Greenhouse, CropBatch } from '@/types'

const loading = ref(false)
const records = ref<EnergyConsumptionRecord[]>([])
const total = ref(0)
const summaryItems = ref<EnergySummary[]>([])
const greenhouses = ref<Greenhouse[]>([])
const batches = ref<CropBatch[]>([])

const greenhouseLabelById = computed(() =>
  buildIdLabelMap(greenhouses.value, g => g.id, greenhouseLabel, '温室')
)
const batchLabelById = computed(() =>
  buildIdLabelMap(batches.value, b => b.id, cropBatchLabel, '批次')
)

function greenhouseName(greenhouseId?: number) {
  if (!greenhouseId) return fallbackIdLabel('温室', greenhouseId)
  return greenhouseLabelById.value[greenhouseId] || fallbackIdLabel('温室', greenhouseId)
}

function batchName(batchId?: number) {
  if (!batchId) return '-'
  return batchLabelById.value[batchId] || fallbackIdLabel('批次', batchId)
}

const filters = reactive({
  record_type: '' as string,
  greenhouse_id: undefined as number | undefined,
  start_date: '' as string,
  end_date: '' as string
})

const pagination = reactive({ page: 1, pageSize: 20 })

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive({
  greenhouse_id: undefined as number | undefined,
  record_type: 'ELECTRICITY',
  consumption_value: 0,
  unit: 'kWh',
  record_period_start: '',
  record_period_end: '',
  batch_id: undefined as number | undefined,
  meter_reading_start: undefined as number | undefined,
  meter_reading_end: undefined as number | undefined
})

const formRules: FormRules = {
  greenhouse_id: [{ required: true, message: '请输入温室ID', trigger: 'blur' }],
  record_type: [{ required: true, message: '请选择能耗类型', trigger: 'change' }],
  consumption_value: [{ required: true, message: '请输入消耗量', trigger: 'blur' }],
  unit: [{ required: true, message: '请选择单位', trigger: 'change' }],
  record_period_start: [{ required: true, message: '请选择周期开始时间', trigger: 'change' }],
  record_period_end: [{ required: true, message: '请选择周期结束时间', trigger: 'change' }]
}

function recordTypeName(type: string) {
  const map: Record<string, string> = { ELECTRICITY: '电力', WATER: '水', CO2_GAS: 'CO2' }
  return map[type] || type
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.record_type) params.record_type = filters.record_type
    if (filters.greenhouse_id) params.greenhouse_id = filters.greenhouse_id
    if (filters.start_date) params.start_date = filters.start_date
    if (filters.end_date) params.end_date = filters.end_date
    const data = await energyApi.getEnergyRecords(params)
    records.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function fetchSummary() {
  try {
    const data = await energyApi.getEnergySummary({})
    summaryItems.value = data.items
  } catch {
    // handled by interceptor
  }
}

function resetFilters() {
  filters.record_type = ''
  filters.greenhouse_id = undefined
  filters.start_date = ''
  filters.end_date = ''
  pagination.page = 1
  fetchData()
}

function openCreateDialog() {
  isEdit.value = false
  editingId.value = null
  Object.assign(formData, {
    greenhouse_id: undefined,
    record_type: 'ELECTRICITY',
    consumption_value: 0,
    unit: 'kWh',
    record_period_start: '',
    record_period_end: '',
    batch_id: undefined,
    meter_reading_start: undefined,
    meter_reading_end: undefined
  })
  dialogVisible.value = true
}

function openEditDialog(record: EnergyConsumptionRecord) {
  isEdit.value = true
  editingId.value = record.id
  Object.assign(formData, {
    greenhouse_id: record.greenhouse_id,
    record_type: record.record_type,
    consumption_value: record.consumption_value,
    unit: record.unit,
    record_period_start: record.record_period_start,
    record_period_end: record.record_period_end,
    batch_id: record.batch_id,
    meter_reading_start: record.meter_reading_start,
    meter_reading_end: record.meter_reading_end
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
    const payload = {
      greenhouse_id: formData.greenhouse_id!,
      record_type: formData.record_type,
      consumption_value: formData.consumption_value,
      unit: formData.unit,
      record_period_start: formData.record_period_start,
      record_period_end: formData.record_period_end,
      batch_id: formData.batch_id || undefined,
      meter_reading_start: formData.meter_reading_start || undefined,
      meter_reading_end: formData.meter_reading_end || undefined
    }
    if (isEdit.value && editingId.value) {
      await energyApi.updateEnergyRecord(editingId.value, payload)
      ElMessage.success('能耗记录已更新')
    } else {
      await energyApi.createEnergyRecord(payload)
      ElMessage.success('能耗记录已创建')
    }
    dialogVisible.value = false
    fetchData()
    fetchSummary()
  } catch {
    // handled by interceptor
  } finally {
    submitLoading.value = false
  }
}

async function removeRecord(id: number) {
  await ElMessageBox.confirm('确认删除该能耗记录？', '提示', { type: 'warning' })
  await energyApi.deleteEnergyRecord(id)
  ElMessage.success('已删除')
  fetchData()
  fetchSummary()
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

onMounted(() => {
  fetchData()
  fetchSummary()
  loadGreenhouses()
  loadBatches()
})
</script>

<style scoped lang="scss">
.energy-records-page {
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
  .summary-cards {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
  }
  .summary-card {
    flex: 1;
    background: var(--bg-card);
    border-radius: var(--radius-md);
    padding: 16px 20px;
    box-shadow: var(--shadow-card);
    text-align: center;
    .summary-label {
      font-size: 14px;
      color: var(--color-text-secondary);
      margin-bottom: 8px;
    }
    .summary-value {
      font-size: 24px;
      font-weight: 700;
      color: var(--color-primary);
    }
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
