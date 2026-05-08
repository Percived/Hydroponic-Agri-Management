<template>
  <div class="harvest-page">
    <div class="page-header">
      <h1 class="page-title">采收记录</h1>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        新增采收
      </el-button>
    </div>

    <!-- 批次汇总卡片 -->
    <div class="summary-cards" v-if="summary.total_weight_kg != null">
      <div class="summary-card">
        <div class="summary-label">总采收重量</div>
        <div class="summary-value">{{ summary.total_weight_kg }} kg</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">采收记录数</div>
        <div class="summary-value">{{ harvestCount }}</div>
      </div>
    </div>

    <!-- 筛选区 -->
    <div class="filter-section">
      <el-select v-model="filters.batch_id" clearable filterable style="width: 260px" placeholder="全部批次">
        <el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (#${b.id})`" :value="b.id" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <!-- 表格 -->
    <div class="table-container">
      <el-table :data="harvests" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="batch_id" label="批次ID" width="100" />
        <el-table-column prop="harvested_at" label="采收时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.harvested_at) }}</template>
        </el-table-column>
        <el-table-column prop="harvest_weight_kg" label="采收总重(kg)" width="140" />
        <el-table-column label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="gradeTagType(row.grade)">{{ gradeName(row.grade) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="grade_weight_kg" label="该等级重量(kg)" width="150" />
        <el-table-column prop="note" label="备注" min-width="200">
          <template #default="{ row }">{{ row.note || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
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

    <!-- 新增弹窗 -->
    <el-dialog v-model="dialogVisible" title="新增采收记录" width="500px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="130px">
        <el-form-item label="批次" prop="batch_id">
          <el-select v-model="formData.batch_id" filterable style="width: 100%">
            <el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (#${b.id})`" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="采收时间" prop="harvested_at">
          <el-date-picker
            v-model="formData.harvested_at"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="采收总重(kg)" prop="harvest_weight_kg">
          <el-input-number v-model="formData.harvest_weight_kg" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="等级" prop="grade">
          <el-select v-model="formData.grade" style="width: 100%">
            <el-option label="A级" value="A" />
            <el-option label="B级" value="B" />
            <el-option label="C级" value="C" />
            <el-option label="Waste(废弃物)" value="Waste" />
          </el-select>
        </el-form-item>
        <el-form-item label="该等级重量(kg)" prop="grade_weight_kg">
          <el-input-number v-model="formData.grade_weight_kg" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.note" type="textarea" :rows="2" placeholder="可选备注" />
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
import { computed, ref, reactive, onMounted } from 'vue'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { cropApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { CropBatch, HarvestRecord } from '@/types'

const loading = ref(false)
const harvests = ref<HarvestRecord[]>([])
const batches = ref<CropBatch[]>([])
const total = ref(0)
const summary = ref<{
  batch_id?: number
  total_weight_kg?: number
  grades?: { grade: string; weight_kg: number; count: number }[]
}>({})

const harvestCount = computed(() => {
  if (!summary.value.grades) return 0
  return summary.value.grades.reduce((sum, g) => sum + (g.count || 0), 0)
})

const filters = reactive({
  batch_id: undefined as number | undefined
})

const pagination = reactive({ page: 1, pageSize: 20 })

const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)

const formData = reactive({
  batch_id: undefined as number | undefined,
  harvested_at: '',
  harvest_weight_kg: 0,
  grade: 'A',
  grade_weight_kg: 0,
  note: '' as string
})

const formRules: FormRules = {
  batch_id: [{ required: true, message: '请输入批次ID', trigger: 'blur' }],
  harvested_at: [{ required: true, message: '请选择采收时间', trigger: 'change' }],
  harvest_weight_kg: [{ required: true, message: '请输入采收总重', trigger: 'blur' }],
  grade: [{ required: true, message: '请选择等级', trigger: 'change' }],
  grade_weight_kg: [{ required: true, message: '请输入该等级重量', trigger: 'blur' }]
}

function gradeName(grade: string) {
  const map: Record<string, string> = { A: 'A级', B: 'B级', C: 'C级', Waste: '废弃物' }
  return map[grade] || grade
}

function gradeTagType(grade: string) {
  const map: Record<string, string> = { A: 'success', B: '', C: 'warning', Waste: 'danger' }
  return map[grade] || 'info'
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.batch_id) params.batch_id = filters.batch_id
    const data = await cropApi.getHarvests(params)
    harvests.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function fetchSummary() {
  if (!filters.batch_id) return
  try {
    const data = await cropApi.getHarvestSummary(filters.batch_id)
    summary.value = data
  } catch {
    summary.value = {}
  }
}

function resetFilters() {
  filters.batch_id = undefined
  pagination.page = 1
  summary.value = {}
  fetchData()
}

function openCreateDialog() {
  Object.assign(formData, {
    batch_id: filters.batch_id,
    harvested_at: '',
    harvest_weight_kg: 0,
    grade: 'A',
    grade_weight_kg: 0,
    note: ''
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
    await cropApi.createHarvest({
      batch_id: formData.batch_id!,
      harvested_at: formData.harvested_at,
      harvest_weight_kg: formData.harvest_weight_kg,
      grade: formData.grade,
      grade_weight_kg: formData.grade_weight_kg,
      note: formData.note || undefined
    })
    ElMessage.success('采收记录已创建')
    dialogVisible.value = false
    fetchData()
    fetchSummary()
  } catch {
    // handled by interceptor
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  const [bRes] = await Promise.all([
    cropApi.getBatches({ page_size: 200 }),
    fetchData()
  ])
  batches.value = bRes.items
})
</script>

<style scoped lang="scss">
.harvest-page {
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
