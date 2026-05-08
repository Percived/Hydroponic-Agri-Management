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
      <el-table :data="tanks" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="液槽编号" width="160" />
        <el-table-column prop="growing_zone_id" label="种植区ID" width="120" />
        <el-table-column label="总容积(L)" width="120">
          <template #default="{ row }">{{ row.total_volume_liter }}</template>
        </el-table-column>
        <el-table-column label="当前容积(L)" width="120">
          <template #default="{ row }">{{ row.current_volume_liter ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑液槽' : '新增液槽'" width="500px">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="120px">
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
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { nutrientApi, greenhouseApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { NutrientTank, GrowingZone } from '@/types'

const loading = ref(false)
const tanks = ref<NutrientTank[]>([])
const total = ref(0)
const zones = ref<GrowingZone[]>([])

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
  total_volume_liter: 100
})

const formRules: FormRules = {
  growing_zone_id: [{ required: true, message: '请输入种植区ID', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入液槽编号', trigger: 'blur' },
    { min: 1, max: 64, message: '编号长度为 1-64 个字符', trigger: 'blur' }
  ],
  total_volume_liter: [{ required: true, message: '请输入总容积', trigger: 'blur' }]
}

function statusName(status: string) {
  const map: Record<string, string> = { IN_USE: '使用中', MAINTENANCE: '维护中', DISABLED: '停用' }
  return map[status] || status
}

function statusTagType(status: string) {
  const map: Record<string, string> = { IN_USE: 'success', MAINTENANCE: 'warning', DISABLED: 'info' }
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
    total_volume_liter: 100
  })
  dialogVisible.value = true
}

function openEditDialog(tank: NutrientTank) {
  isEdit.value = true
  editingId.value = tank.id
  Object.assign(formData, {
    growing_zone_id: tank.growing_zone_id,
    code: tank.code,
    total_volume_liter: tank.total_volume_liter
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
      growing_zone_id: formData.growing_zone_id!,
      code: formData.code,
      total_volume_liter: formData.total_volume_liter
    }
    if (isEdit.value && editingId.value) {
      await nutrientApi.updateNutrientTank(editingId.value, payload)
      ElMessage.success('液槽已更新')
    } else {
      await nutrientApi.createNutrientTank(payload)
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
    const data = await greenhouseApi.getGrowingZones({ page_size: LARGE_PAGE_SIZE })
    zones.value = data.items
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchData()
  loadZones()
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
}
</style>
