<template>
  <div class="recipes-page">
    <div class="page-header">
      <h1 class="page-title">营养配方</h1>
      <el-button type="primary" @click="openCreateRecipe">
        <el-icon><Plus /></el-icon>
        新增配方
      </el-button>
    </div>

    <div class="filter-section">
      <el-select v-model="filters.status" placeholder="状态" clearable style="width: 160px">
        <el-option label="草稿" value="DRAFT" />
        <el-option label="启用" value="ACTIVE" />
        <el-option label="归档" value="ARCHIVED" />
      </el-select>
      <el-button type="primary" @click="fetchData">查询</el-button>
      <el-button @click="resetFilters">重置</el-button>
    </div>

    <div class="table-container">
      <el-table :data="recipes" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="recipe_code" label="配方编号" width="150" />
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="effective_from" label="生效时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.effective_from) }}</template>
        </el-table-column>
        <el-table-column prop="effective_to" label="失效时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.effective_to) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditRecipe(row)">编辑</el-button>
            <el-button type="success" link @click="openTargetsDialog(row)">指标</el-button>
            <el-button type="danger" link @click="removeRecipe(row.id)">删除</el-button>
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

    <!-- 配方新增/编辑弹窗 -->
    <el-dialog v-model="recipeDialogVisible" :title="isEditRecipe ? '编辑配方' : '新增配方'" width="500px">
      <el-form ref="recipeFormRef" :model="recipeForm" :rules="recipeFormRules" label-width="100px">
        <el-form-item label="配方编号" prop="recipe_code">
          <el-input v-model="recipeForm.recipe_code" placeholder="请输入编号" maxlength="64" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="recipeForm.name" placeholder="请输入名称" maxlength="128" />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="recipeForm.version" placeholder="如 v1.0" maxlength="32" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="recipeForm.status" style="width: 100%">
            <el-option label="草稿" value="DRAFT" />
            <el-option label="启用" value="ACTIVE" />
            <el-option label="归档" value="ARCHIVED" />
          </el-select>
        </el-form-item>
        <el-form-item label="生效时间">
          <el-date-picker
            v-model="recipeForm.effective_from"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="失效时间">
          <el-date-picker
            v-model="recipeForm.effective_to"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recipeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="recipeSubmitLoading" @click="handleRecipeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 指标管理弹窗（含阶段指标+离子目标两个Tab） -->
    <el-dialog v-model="targetsDialogVisible" title="配方指标" width="800px" @closed="resetTargetsEdit">
      <el-tabs v-model="activeTargetTab" type="card">
        <!-- Tab 1: 阶段指标 -->
        <el-tab-pane label="阶段指标" name="stage">
          <div style="margin-bottom: 12px">
            <el-button type="primary" size="small" @click="openCreateStageTarget">新增阶段指标</el-button>
          </div>
          <el-table :data="stageTargets" v-loading="targetsLoading" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="growth_stage_id" label="生长阶段" width="90">
              <template #default="{ row }">
                {{ row.growth_stage_id ?? '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="metric_code" label="指标代码" width="110" />
            <el-table-column label="目标范围" min-width="160">
              <template #default="{ row }">
                <span v-if="row.target_min != null || row.target_max != null">
                  {{ row.target_min ?? '-' }} ~ {{ row.target_max ?? '-' }}
                </span>
                <span v-else>-</span>
                <span v-if="row.tolerance != null" class="tolerance">&plusmn;{{ row.tolerance }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="unit" label="单位" width="80" />
            <el-table-column prop="enabled" label="启用" width="70">
              <template #default="{ row }">
                <el-switch :model-value="row.enabled" disabled size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="110" fixed="right">
              <template #default="{ row, $index }">
                <el-button type="primary" link size="small" @click="openEditStageTarget(row, $index)">编辑</el-button>
                <el-button type="danger" link size="small" @click="removeStageTarget($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 离子目标 -->
        <el-tab-pane label="离子目标" name="ion">
          <div style="margin-bottom: 12px">
            <el-button type="primary" size="small" @click="openCreateIonTarget">新增离子目标</el-button>
          </div>
          <el-table :data="ionTargets" v-loading="targetsLoading" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="growth_stage_id" label="生长阶段" width="90">
              <template #default="{ row }">
                {{ row.growth_stage_id ?? '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="ion_code" label="离子代码" width="110" />
            <el-table-column label="范围 (mg/L)" min-width="160">
              <template #default="{ row }">
                <span v-if="row.target_min_mg_l != null || row.target_max_mg_l != null">
                  {{ row.target_min_mg_l ?? '-' }} ~ {{ row.target_max_mg_l ?? '-' }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="enabled" label="启用" width="70">
              <template #default="{ row }">
                <el-switch :model-value="row.enabled" disabled size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="110" fixed="right">
              <template #default="{ row, $index }">
                <el-button type="primary" link size="small" @click="openEditIonTarget(row, $index)">编辑</el-button>
                <el-button type="danger" link size="small" @click="removeIonTarget($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="targetsDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="targetsSaveLoading" @click="handleTargetsSave">保存指标</el-button>
      </template>
    </el-dialog>

    <!-- 阶段指标新增/编辑弹窗 -->
    <el-dialog v-model="stageTargetFormVisible" :title="isEditStageTarget ? '编辑阶段指标' : '新增阶段指标'" width="500px">
      <el-form ref="stageTargetFormRef" :model="stageTargetForm" :rules="stageTargetFormRules" label-width="120px">
        <el-form-item label="生长阶段ID">
          <el-input-number v-model="stageTargetForm.growth_stage_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="指标代码" prop="metric_code">
          <el-select v-model="stageTargetForm.metric_code" placeholder="选择指标" filterable style="width: 100%">
            <el-option v-for="m in metrics" :key="m.code" :label="`${m.name} (${m.code})`" :value="m.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标最小值">
          <el-input-number v-model="stageTargetForm.target_min" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="目标最大值">
          <el-input-number v-model="stageTargetForm.target_max" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="容差">
          <el-input-number v-model="stageTargetForm.tolerance" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="stageTargetForm.unit" placeholder="如 ℃, mS/cm" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="stageTargetForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stageTargetFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleStageTargetFormSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 离子目标新增/编辑弹窗 -->
    <el-dialog v-model="ionTargetFormVisible" :title="isEditIonTarget ? '编辑离子目标' : '新增离子目标'" width="500px">
      <el-form ref="ionTargetFormRef" :model="ionTargetForm" :rules="ionTargetFormRules" label-width="120px">
        <el-form-item label="生长阶段ID">
          <el-input-number v-model="ionTargetForm.growth_stage_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="离子代码" prop="ion_code">
          <el-input v-model="ionTargetForm.ion_code" placeholder="如 NO3, K, Ca" maxlength="32" />
        </el-form-item>
        <el-form-item label="最小值 (mg/L)">
          <el-input-number v-model="ionTargetForm.target_min_mg_l" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大值 (mg/L)">
          <el-input-number v-model="ionTargetForm.target_max_mg_l" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ionTargetForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ionTargetFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleIonTargetFormSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { recipeApi, metricApi } from '@/api'
import { formatDateTime, populateMetricNames } from '@/utils/format'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { NutrientRecipe, RecipeStatus, RecipeStageTarget, RecipeIonTarget, MetricDefinition, CreateStageTargetParams, CreateIonTargetParams } from '@/types'

// ── Recipes ──
const loading = ref(false)
const recipes = ref<NutrientRecipe[]>([])
const total = ref(0)

const filters = reactive({
  status: '' as string
})

const pagination = reactive({ page: 1, pageSize: 20 })

const recipeDialogVisible = ref(false)
const isEditRecipe = ref(false)
const recipeFormRef = ref<FormInstance>()
const recipeSubmitLoading = ref(false)
const editingRecipeId = ref<number | null>(null)

const emptyRecipeForm = () => ({
  recipe_code: '',
  name: '',
  version: '',
  status: 'DRAFT' as RecipeStatus,
  effective_from: '' as string,
  effective_to: '' as string
})

const recipeForm = reactive(emptyRecipeForm())

const recipeFormRules: FormRules = {
  recipe_code: [{ required: true, message: '请输入配方编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本', trigger: 'blur' }]
}

function statusName(s: string) {
  const map: Record<string, string> = { DRAFT: '草稿', ACTIVE: '启用', ARCHIVED: '归档' }
  return map[s] || s
}

function statusTagType(s: string) {
  const map: Record<string, string> = { DRAFT: 'info', ACTIVE: 'success', ARCHIVED: 'warning' }
  return map[s] || 'info'
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.status) params.status = filters.status
    const data = await recipeApi.getRecipes(params)
    recipes.value = data.items
    total.value = data.total
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.status = ''
  pagination.page = 1
  fetchData()
}

function openCreateRecipe() {
  isEditRecipe.value = false
  editingRecipeId.value = null
  Object.assign(recipeForm, emptyRecipeForm())
  recipeDialogVisible.value = true
}

function openEditRecipe(recipe: NutrientRecipe) {
  isEditRecipe.value = true
  editingRecipeId.value = recipe.id
  Object.assign(recipeForm, {
    recipe_code: recipe.recipe_code,
    name: recipe.name,
    version: recipe.version,
    status: recipe.status,
    effective_from: recipe.effective_from || '',
    effective_to: recipe.effective_to || ''
  })
  recipeDialogVisible.value = true
}

async function handleRecipeSubmit() {
  if (!recipeFormRef.value) return
  try {
    await recipeFormRef.value.validate()
  } catch {
    return
  }

  recipeSubmitLoading.value = true
  try {
    const payload = {
      recipe_code: recipeForm.recipe_code,
      name: recipeForm.name,
      version: recipeForm.version,
      status: (recipeForm.status || undefined) as RecipeStatus | undefined,
      effective_from: recipeForm.effective_from || undefined,
      effective_to: recipeForm.effective_to || undefined
    }
    if (isEditRecipe.value && editingRecipeId.value) {
      await recipeApi.updateRecipe(editingRecipeId.value, payload)
      ElMessage.success('配方已更新')
    } else {
      await recipeApi.createRecipe(payload)
      ElMessage.success('配方已创建')
    }
    recipeDialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  } finally {
    recipeSubmitLoading.value = false
  }
}

async function removeRecipe(id: number) {
  await ElMessageBox.confirm('确认删除该配方？', '提示', { type: 'warning' })
  await recipeApi.deleteRecipe(id)
  ElMessage.success('已删除')
  fetchData()
}

// ── Targets (tabs: stage + ion) ──
const targetsDialogVisible = ref(false)
const targetsLoading = ref(false)
const targetsSaveLoading = ref(false)
const stageTargets = ref<RecipeStageTarget[]>([])
const ionTargets = ref<RecipeIonTarget[]>([])
const metrics = ref<MetricDefinition[]>([])
const currentRecipeId = ref<number | null>(null)
const activeTargetTab = ref('stage')

// Stage target form
const stageTargetFormVisible = ref(false)
const isEditStageTarget = ref(false)
const editingStageIndex = ref<number | null>(null)
const stageTargetFormRef = ref<FormInstance>()

const emptyStageTargetForm = () => ({
  growth_stage_id: undefined as number | undefined,
  metric_code: '',
  target_min: undefined as number | undefined,
  target_max: undefined as number | undefined,
  tolerance: undefined as number | undefined,
  unit: '' as string,
  enabled: true as boolean
})

const stageTargetForm = reactive(emptyStageTargetForm())

const stageTargetFormRules: FormRules = {
  metric_code: [{ required: true, message: '请选择指标代码', trigger: 'change' }]
}

// Ion target form
const ionTargetFormVisible = ref(false)
const isEditIonTarget = ref(false)
const editingIonIndex = ref<number | null>(null)
const ionTargetFormRef = ref<FormInstance>()

const emptyIonTargetForm = () => ({
  growth_stage_id: undefined as number | undefined,
  ion_code: '',
  target_min_mg_l: undefined as number | undefined,
  target_max_mg_l: undefined as number | undefined,
  enabled: true as boolean
})

const ionTargetForm = reactive(emptyIonTargetForm())

const ionTargetFormRules: FormRules = {
  ion_code: [{ required: true, message: '请输入离子代码', trigger: 'blur' }]
}

async function loadMetrics() {
  try {
    const data = await metricApi.getMetrics({ page_size: LARGE_PAGE_SIZE })
    metrics.value = data.items
    populateMetricNames(data.items)
  } catch {
    metrics.value = []
  }
}

async function openTargetsDialog(recipe: NutrientRecipe) {
  currentRecipeId.value = recipe.id
  targetsDialogVisible.value = true
  targetsLoading.value = true
  try {
    const data = await recipeApi.getRecipeTargets(recipe.id)
    stageTargets.value = data.stage_targets || []
    ionTargets.value = data.ion_targets || []
  } catch {
    stageTargets.value = []
    ionTargets.value = []
  } finally {
    targetsLoading.value = false
  }
}

function resetTargetsEdit() {
  stageTargetFormVisible.value = false
  ionTargetFormVisible.value = false
}

// ── Stage target CRUD (local array) ──
function openCreateStageTarget() {
  isEditStageTarget.value = false
  editingStageIndex.value = null
  Object.assign(stageTargetForm, emptyStageTargetForm())
  stageTargetFormVisible.value = true
}

function openEditStageTarget(target: RecipeStageTarget, index: number) {
  isEditStageTarget.value = true
  editingStageIndex.value = index
  Object.assign(stageTargetForm, {
    growth_stage_id: target.growth_stage_id,
    metric_code: target.metric_code,
    target_min: target.target_min,
    target_max: target.target_max,
    tolerance: target.tolerance,
    unit: target.unit || '',
    enabled: target.enabled
  })
  stageTargetFormVisible.value = true
}

function handleStageTargetFormSubmit() {
  if (!stageTargetFormRef.value) return
  stageTargetFormRef.value.validate().then(() => {
    const item: CreateStageTargetParams & { id?: number } = {
      growth_stage_id: stageTargetForm.growth_stage_id || null,
      metric_code: stageTargetForm.metric_code,
      target_min: stageTargetForm.target_min ?? null,
      target_max: stageTargetForm.target_max ?? null,
      tolerance: stageTargetForm.tolerance ?? null,
      unit: stageTargetForm.unit || undefined,
      enabled: stageTargetForm.enabled
    }
    if (isEditStageTarget.value && editingStageIndex.value != null) {
      // Preserve existing id so backend knows it's an update
      const existing = stageTargets.value[editingStageIndex.value]
      if (existing) item.id = existing.id
      stageTargets.value.splice(editingStageIndex.value, 1, item as unknown as RecipeStageTarget)
    } else {
      stageTargets.value.push(item as unknown as RecipeStageTarget)
    }
    stageTargetFormVisible.value = false
  }).catch(() => { /* validation failed */ })
}

function removeStageTarget(index: number) {
  stageTargets.value.splice(index, 1)
}

// ── Ion target CRUD (local array) ──
function openCreateIonTarget() {
  isEditIonTarget.value = false
  editingIonIndex.value = null
  Object.assign(ionTargetForm, emptyIonTargetForm())
  ionTargetFormVisible.value = true
}

function openEditIonTarget(target: RecipeIonTarget, index: number) {
  isEditIonTarget.value = true
  editingIonIndex.value = index
  Object.assign(ionTargetForm, {
    growth_stage_id: target.growth_stage_id,
    ion_code: target.ion_code,
    target_min_mg_l: target.target_min_mg_l,
    target_max_mg_l: target.target_max_mg_l,
    enabled: target.enabled
  })
  ionTargetFormVisible.value = true
}

function handleIonTargetFormSubmit() {
  if (!ionTargetFormRef.value) return
  ionTargetFormRef.value.validate().then(() => {
    const item: CreateIonTargetParams & { id?: number } = {
      growth_stage_id: ionTargetForm.growth_stage_id || null,
      ion_code: ionTargetForm.ion_code,
      target_min_mg_l: ionTargetForm.target_min_mg_l ?? null,
      target_max_mg_l: ionTargetForm.target_max_mg_l ?? null,
      enabled: ionTargetForm.enabled
    }
    if (isEditIonTarget.value && editingIonIndex.value != null) {
      const existing = ionTargets.value[editingIonIndex.value]
      if (existing) item.id = existing.id
      ionTargets.value.splice(editingIonIndex.value, 1, item as unknown as RecipeIonTarget)
    } else {
      ionTargets.value.push(item as unknown as RecipeIonTarget)
    }
    ionTargetFormVisible.value = false
  }).catch(() => { /* validation failed */ })
}

function removeIonTarget(index: number) {
  ionTargets.value.splice(index, 1)
}

// ── Save all targets (bulk replace) ──
async function handleTargetsSave() {
  if (!currentRecipeId.value) return
  targetsSaveLoading.value = true
  try {
    const stagePayload: CreateStageTargetParams[] = stageTargets.value.map(t => ({
      id: t.id,
      growth_stage_id: t.growth_stage_id,
      metric_code: t.metric_code,
      target_min: t.target_min,
      target_max: t.target_max,
      tolerance: t.tolerance,
      unit: t.unit,
      enabled: t.enabled
    }))
    const ionPayload: CreateIonTargetParams[] = ionTargets.value.map(t => ({
      id: t.id,
      growth_stage_id: t.growth_stage_id,
      ion_code: t.ion_code,
      target_min_mg_l: t.target_min_mg_l,
      target_max_mg_l: t.target_max_mg_l,
      enabled: t.enabled
    }))
    await recipeApi.updateRecipeTargets(currentRecipeId.value, {
      stage_targets: stagePayload,
      ion_targets: ionPayload
    })
    ElMessage.success('指标已保存')
    // Refresh from server
    targetsLoading.value = true
    try {
      const data = await recipeApi.getRecipeTargets(currentRecipeId.value)
      stageTargets.value = data.stage_targets || []
      ionTargets.value = data.ion_targets || []
    } catch {
      stageTargets.value = []
      ionTargets.value = []
    } finally {
      targetsLoading.value = false
    }
  } catch {
    // handled by interceptor
  } finally {
    targetsSaveLoading.value = false
  }
}

onMounted(() => {
  loadMetrics()
  fetchData()
})
</script>

<style scoped lang="scss">
.recipes-page {
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
  .tolerance {
    color: var(--color-text-secondary);
    font-size: 12px;
    margin-left: 4px;
  }
}
</style>
