<template>
  <div class="batch-ledger-page">
    <div class="page-header">
      <h1 class="page-title">批次台账</h1>
      <el-button type="primary" @click="openCreateDialog">创建批次</el-button>
    </div>

    <div class="filter-section">
      <el-form :inline="true">
        <el-form-item label="温室">
          <el-select v-model="filters.greenhouse_id" clearable filterable style="width: 180px" placeholder="全部">
            <el-option v-for="g in greenhouses" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="作物品种">
          <el-select v-model="filters.crop_variety_id" clearable filterable style="width: 180px" placeholder="全部">
            <el-option v-for="v in varieties" :key="v.id" :label="v.name" :value="v.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable style="width: 150px">
            <el-option label="PLANNED" value="PLANNED" />
            <el-option label="RUNNING" value="RUNNING" />
            <el-option label="HARVESTING" value="HARVESTING" />
            <el-option label="COMPLETED" value="COMPLETED" />
            <el-option label="ABORTED" value="ABORTED" />
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
          <el-button type="primary" @click="fetchData">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="batches" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="batch_no" label="批次编号" width="160" />
        <el-table-column prop="greenhouse_id" label="温室ID" width="100" />
        <el-table-column prop="crop_variety_id" label="作物ID" width="100" />
        <el-table-column label="作物名称" width="120">
          <template #default="{ row }">{{ row.variety_name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="130">
          <template #default="{ row }">
            <el-select
              v-if="canControlDevice() && getLegalTransitions(row.status).length > 0"
              :model-value="row.status"
              size="small"
              style="width: 120px"
              @change="onStatusChange(row.id, $event)"
            >
              <el-option
                v-for="s in getLegalTransitions(row.status)"
                :key="s"
                :label="s"
                :value="s"
              />
            </el-select>
            <el-tag v-else size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="growing_zone_id" label="种植区" width="80" />
        <el-table-column prop="planting_density" label="定植密度(株/㎡)" width="120" />
        <el-table-column prop="total_plants" label="总株数" width="90" />
        <el-table-column label="预计采收" width="180">
          <template #default="{ row }">{{ formatDateTime(row.expected_harvest_at) }}</template>
        </el-table-column>
        <el-table-column prop="recipe_version" label="配方版本" width="100" />
        <el-table-column prop="policy_version" label="策略版本" width="100" />
        <el-table-column prop="started_at" label="开始时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
        </el-table-column>
        <el-table-column prop="ended_at" label="结束时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.ended_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="goDetail(row.id)">详情</el-button>
            <el-button type="danger" size="small" link @click="removeBatch(row.id)">删除</el-button>
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

    <el-dialog v-model="createVisible" title="创建批次" width="700px">
      <batch-form v-model="formData" />
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitCreate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cropApi, greenhouseApi } from '@/api'
import BatchForm from '@/components/batch/BatchForm.vue'
import { formatDateTime } from '@/utils/format'
import { usePermission } from '@/composables'
import type { CropBatch, CreateCropBatchRequest, CropVariety, Greenhouse } from '@/types'

const router = useRouter()
const { canControlDevice } = usePermission()

// Legal state transitions map
const LEGAL_TRANSITIONS: Record<string, string[]> = {
  PLANNED: ['RUNNING', 'ABORTED'],
  RUNNING: ['HARVESTING', 'ABORTED'],
  HARVESTING: ['COMPLETED', 'ABORTED'],
  COMPLETED: [],
  ABORTED: []
}

function getLegalTransitions(status: string): string[] {
  return LEGAL_TRANSITIONS[status] || []
}

function goDetail(id: number) {
  router.push(`/batches/${id}`)
}

const loading = ref(false)
const submitLoading = ref(false)
const batches = ref<CropBatch[]>([])
const total = ref(0)
const range = ref<[string, string] | null>(null)
const greenhouses = ref<Greenhouse[]>([])
const varieties = ref<CropVariety[]>([])
const filters = reactive({
  greenhouse_id: undefined as number | undefined,
  crop_variety_id: undefined as number | undefined,
  status: undefined as string | undefined
})
const pagination = reactive({ page: 1, pageSize: 20 })

const createVisible = ref(false)
const formData = ref<CreateCropBatchRequest>({
  batch_no: '',
  greenhouse_id: 1,
  crop_variety_id: 1,
  started_at: new Date().toISOString(),
  expected_harvest_at: '',
  recipe_version: '',
  policy_version: ''
})

async function fetchData() {
  loading.value = true
  try {
    const result = await cropApi.getBatches({
      greenhouse_id: filters.greenhouse_id,
      crop_variety_id: filters.crop_variety_id,
      status: filters.status,
      start_time: range.value?.[0],
      end_time: range.value?.[1],
      page: pagination.page,
      page_size: pagination.pageSize
    })
    batches.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.greenhouse_id = undefined
  filters.crop_variety_id = undefined
  filters.status = undefined
  range.value = null
  pagination.page = 1
  fetchData()
}

function openCreateDialog() {
  formData.value = {
    batch_no: `BATCH-${new Date().getFullYear()}-${Math.floor(Math.random() * 900 + 100)}`,
    greenhouse_id: 1,
    crop_variety_id: 1,
    started_at: new Date().toISOString(),
    expected_harvest_at: '',
    recipe_version: '',
    policy_version: ''
  }
  createVisible.value = true
}

async function submitCreate() {
  if (!formData.value.batch_no || !formData.value.started_at) {
    ElMessage.warning('请先填写必要字段')
    return
  }
  submitLoading.value = true
  try {
    await cropApi.createBatch({
      ...formData.value,
      expected_harvest_at: formData.value.expected_harvest_at || undefined,
      recipe_version: formData.value.recipe_version || undefined,
      policy_version: formData.value.policy_version || undefined
    })
    ElMessage.success('批次创建成功')
    createVisible.value = false
    await fetchData()
  } finally {
    submitLoading.value = false
  }
}

async function updateStatus(batchId: number, status: string) {
  try {
    await cropApi.transitionBatch(batchId, { status })
    ElMessage.success('状态已更新')
  } catch {
    // Revert on failure — refetch to restore original status
    fetchData()
  }
}

function onStatusChange(batchId: number, value: unknown) {
  updateStatus(batchId, value as string)
}

async function removeBatch(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该批次？', '删除确认', { type: 'warning' })
  } catch {
    return // user cancelled
  }
  try {
    await cropApi.deleteBatch(id)
    ElMessage.success('批次已删除')
    await fetchData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  Promise.all([
    fetchData(),
    greenhouseApi.getGreenhouses({ page_size: 200 }).then(res => greenhouses.value = res.items),
    cropApi.getCropVarieties({ page_size: 200 }).then(res => varieties.value = res.items)
  ])
})
</script>

<style scoped lang="scss">
.batch-ledger-page {
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
  .filter-section,
  .table-container {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
    padding: 16px;
  }
  .table-container {
    margin-top: 12px;
  }
  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
}
</style>
