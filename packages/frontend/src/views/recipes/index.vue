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

    <!-- 指标列表弹窗（嵌套） -->
    <el-dialog v-model="targetsDialogVisible" title="配方指标" width="700px">
      <div style="margin-bottom: 12px">
        <el-button type="primary" size="small" @click="openCreateTarget">新增指标</el-button>
      </div>
      <el-table :data="targets" v-loading="targetsLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="growth_stage_id" label="生长阶段ID" width="110" />
        <el-table-column prop="metric_code" label="指标代码" width="120" />
        <el-table-column label="目标值范围" min-width="180">
          <template #default="{ row }">
            <span v-if="row.target_min != null || row.target_max != null">
              {{ row.target_min ?? '-' }} ~ {{ row.target_max ?? '-' }}
            </span>
            <span v-else>{{ row.target_value ?? '-' }}</span>
            <span v-if="row.tolerance != null" class="tolerance">&plusmn;{{ row.tolerance }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditTarget(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="removeTarget(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="targetsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 指标新增/编辑弹窗 -->
    <el-dialog v-model="targetFormVisible" :title="isEditTarget ? '编辑指标' : '新增指标'" width="500px">
      <el-form ref="targetFormRef" :model="targetForm" :rules="targetFormRules" label-width="120px">
        <el-form-item label="生长阶段ID">
          <el-input-number v-model="targetForm.growth_stage_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="指标代码" prop="metric_code">
          <el-input v-model="targetForm.metric_code" placeholder="如 TEMP, EC, PH" />
        </el-form-item>
        <el-form-item label="目标最小值">
          <el-input-number v-model="targetForm.target_min" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="目标最大值">
          <el-input-number v-model="targetForm.target_max" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="精确目标值">
          <el-input-number v-model="targetForm.target_value" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="容差">
          <el-input-number v-model="targetForm.tolerance" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="targetForm.unit" placeholder="如 ℃, mS/cm" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="targetFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="targetSubmitLoading" @click="handleTargetSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { recipeApi } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { NutrientRecipe, RecipeStatus, RecipeTarget } from '@/types'

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

// ── Targets (nested) ──
const targetsDialogVisible = ref(false)
const targetsLoading = ref(false)
const targets = ref<RecipeTarget[]>([])
const currentRecipeId = ref<number | null>(null)

const targetFormVisible = ref(false)
const isEditTarget = ref(false)
const targetFormRef = ref<FormInstance>()
const targetSubmitLoading = ref(false)
const editingTargetId = ref<number | null>(null)

const emptyTargetForm = () => ({
  growth_stage_id: undefined as number | undefined,
  metric_code: '',
  target_min: undefined as number | undefined,
  target_max: undefined as number | undefined,
  target_value: undefined as number | undefined,
  tolerance: undefined as number | undefined,
  unit: '' as string
})

const targetForm = reactive(emptyTargetForm())

const targetFormRules: FormRules = {
  metric_code: [{ required: true, message: '请输入指标代码', trigger: 'blur' }]
}

async function openTargetsDialog(recipe: NutrientRecipe) {
  currentRecipeId.value = recipe.id
  targetsDialogVisible.value = true
  targetsLoading.value = true
  try {
    const data = await recipeApi.getRecipeTargets(recipe.id)
    targets.value = data.items
  } catch {
    targets.value = []
  } finally {
    targetsLoading.value = false
  }
}

function openCreateTarget() {
  isEditTarget.value = false
  editingTargetId.value = null
  Object.assign(targetForm, emptyTargetForm())
  targetFormVisible.value = true
}

function openEditTarget(target: RecipeTarget) {
  isEditTarget.value = true
  editingTargetId.value = target.id
  Object.assign(targetForm, {
    growth_stage_id: target.growth_stage_id,
    metric_code: target.metric_code,
    target_min: target.target_min,
    target_max: target.target_max,
    target_value: target.target_value,
    tolerance: target.tolerance,
    unit: target.unit || ''
  })
  targetFormVisible.value = true
}

async function handleTargetSubmit() {
  if (!targetFormRef.value) return
  try {
    await targetFormRef.value.validate()
  } catch {
    return
  }

  if (!currentRecipeId.value) return
  targetSubmitLoading.value = true
  try {
    const payload = {
      growth_stage_id: targetForm.growth_stage_id || undefined,
      metric_code: targetForm.metric_code,
      target_min: targetForm.target_min,
      target_max: targetForm.target_max,
      target_value: targetForm.target_value,
      tolerance: targetForm.tolerance,
      unit: targetForm.unit || undefined
    }
    if (isEditTarget.value && editingTargetId.value) {
      await recipeApi.updateRecipeTarget(currentRecipeId.value, editingTargetId.value, payload)
      ElMessage.success('指标已更新')
    } else {
      await recipeApi.createRecipeTarget(currentRecipeId.value, payload)
      ElMessage.success('指标已创建')
    }
    targetFormVisible.value = false
    // Refresh targets
    targetsLoading.value = true
    try {
      const data = await recipeApi.getRecipeTargets(currentRecipeId.value)
      targets.value = data.items
    } catch {
      targets.value = []
    } finally {
      targetsLoading.value = false
    }
  } catch {
    // handled by interceptor
  } finally {
    targetSubmitLoading.value = false
  }
}

async function removeTarget(targetId: number) {
  if (!currentRecipeId.value) return
  await ElMessageBox.confirm('确认删除该指标？', '提示', { type: 'warning' })
  await recipeApi.deleteRecipeTarget(currentRecipeId.value, targetId)
  ElMessage.success('已删除')
  targetsLoading.value = true
  try {
    const data = await recipeApi.getRecipeTargets(currentRecipeId.value)
    targets.value = data.items
  } catch {
    targets.value = []
  } finally {
    targetsLoading.value = false
  }
}

onMounted(() => {
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
