<template>
  <div class="nutrient-recipe-detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="router.push('/nutrient/recipes')" text>
          <el-icon><ArrowLeft /></el-icon>返回营养配方
        </el-button>
        <h1 class="page-title">{{ recipe?.name || '配方详情' }}</h1>
        <el-tag v-if="recipe" size="large" :type="statusTagType">{{ statusLabel }}</el-tag>
      </div>
      <div class="header-right" v-if="recipe">
        <el-button v-if="canManage" type="primary" @click="openEditDialog">
          <el-icon><Edit /></el-icon>编辑
        </el-button>
        <el-button v-if="canManage && recipe.status === 'DRAFT'" type="success" @click="openPublishDialog">
          <el-icon><Finished /></el-icon>发布
        </el-button>
        <el-button v-if="canDelete" type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>删除
        </el-button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading && !recipe" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- Error -->
    <el-result
      v-else-if="errorMsg" icon="error" :title="errorMsg"
      sub-title="请检查网络连接或返回列表重试"
    >
      <template #extra>
        <el-button type="primary" @click="loadAllData">重新加载</el-button>
        <el-button @click="router.push('/nutrient/recipes')">返回列表</el-button>
      </template>
    </el-result>

    <!-- Content -->
    <template v-else-if="recipe">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <el-card shadow="never" class="info-card">
            <template #header><span class="card-title">配方信息</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="配方编码">{{ recipe.recipe_code }}</el-descriptions-item>
              <el-descriptions-item label="配方名称">{{ recipe.name }}</el-descriptions-item>
              <el-descriptions-item label="作物品种">{{ cropVarietyName(recipe.crop_variety_id) }}</el-descriptions-item>
              <el-descriptions-item label="版本">{{ recipe.version }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType">{{ statusLabel }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="生效时间">{{ recipe.effective_from ? formatDateTime(recipe.effective_from) : '-' }}</el-descriptions-item>
              <el-descriptions-item label="失效时间">{{ recipe.effective_to ? formatDateTime(recipe.effective_to) : '-' }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ recipe.published_at ? formatDateTime(recipe.published_at) : '-' }}</el-descriptions-item>
              <el-descriptions-item label="描述">{{ recipe.description || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDateTime(recipe.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatDateTime(recipe.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-row :gutter="16" class="stats-row">
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ stageTargets.length }}</div>
                <div class="stat-label">阶段指标</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ ionTargets.length }}</div>
                <div class="stat-label">离子目标</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="stat-card">
                <div class="stat-value">{{ recipe.version }}</div>
                <div class="stat-label">当前版本</div>
              </div>
            </el-col>
          </el-row>
        </el-tab-pane>

        <!-- Tab 2: 指标配置 -->
        <el-tab-pane label="指标配置" name="targets">
          <el-card shadow="never" class="section-card">
            <template #header><span class="card-title">阶段指标</span></template>
            <el-table :data="stageTargets" v-loading="targetsLoading" stripe size="small">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column label="生长阶段" width="140">
                <template #default="{ row }">{{ growthStageName(row.growth_stage_id) }}</template>
              </el-table-column>
              <el-table-column prop="metric_code" label="指标" width="120" />
              <el-table-column label="目标范围" min-width="200">
                <template #default="{ row }">
                  {{ row.target_min ?? '—' }} ~ {{ row.target_max ?? '—' }}
                  <span v-if="row.tolerance" class="hint">(±{{ row.tolerance }})</span>
                </template>
              </el-table-column>
              <el-table-column prop="unit" label="单位" width="80">
                <template #default="{ row }">{{ row.unit || '-' }}</template>
              </el-table-column>
              <el-table-column label="启用" width="70">
                <template #default="{ row }">
                  <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!targetsLoading && stageTargets.length === 0" description="暂无阶段指标" :image-size="40" />
          </el-card>

          <el-card shadow="never" class="section-card" style="margin-top: 16px">
            <template #header><span class="card-title">离子目标</span></template>
            <el-table :data="ionTargets" stripe size="small">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column label="生长阶段" width="140">
                <template #default="{ row }">{{ growthStageName(row.growth_stage_id) }}</template>
              </el-table-column>
              <el-table-column prop="ion_code" label="离子" width="100" />
              <el-table-column label="目标范围 (mg/L)" min-width="200">
                <template #default="{ row }">
                  {{ row.target_min_mg_l ?? '—' }} ~ {{ row.target_max_mg_l ?? '—' }}
                </template>
              </el-table-column>
              <el-table-column label="启用" width="70">
                <template #default="{ row }">
                  <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!targetsLoading && ionTargets.length === 0" description="暂无离子目标" :image-size="40" />
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else description="配方不存在" />

    <!-- Edit Dialog -->
    <el-dialog v-model="editDialogVisible" title="编辑配方" width="500px">
      <el-form ref="editFormRef" :model="editForm" :rules="editFormRules" label-width="120px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="作物品种">
          <el-select v-model="editForm.crop_variety_id" placeholder="选择作物品种" clearable filterable style="width: 100%">
            <el-option v-for="v in cropVarieties" :key="v.id" :label="`${v.name} (${v.code})`" :value="v.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="editForm.version" maxlength="32" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="生效时间">
          <el-date-picker v-model="editForm.effective_from" type="datetime" format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" style="width: 100%" />
        </el-form-item>
        <el-form-item label="失效时间">
          <el-date-picker v-model="editForm.effective_to" type="datetime" format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSubmitting" @click="handleEditSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- Publish Dialog -->
    <el-dialog v-model="publishDialogVisible" title="发布配方" width="400px">
      <el-form ref="publishFormRef" label-width="120px">
        <el-form-item label="发布版本" prop="version">
          <el-input v-model="publishVersion" placeholder="如 2.0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="publishDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="publishSubmitting" @click="handlePublish">确定发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { ArrowLeft, Edit, Delete, Finished } from '@element-plus/icons-vue'
import { recipeApi, cropApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, fallbackIdLabel, growthStageLabel } from '@/utils/labels'
import { LARGE_PAGE_SIZE } from '@/utils/constants'
import type { NutrientRecipe, RecipeStageTarget, RecipeIonTarget, GrowthStage, CropVariety } from '@/types'
import { Role } from '@/types'

const route = useRoute()
const router = useRouter()

const recipeId = computed(() => Number(route.params.id))
const { hasRole, canControlDevice } = usePermission()
const canManage = computed(() => canControlDevice())
const canDelete = computed(() => hasRole(Role.ADMIN))

const loading = ref(false)
const errorMsg = ref('')
const recipe = ref<NutrientRecipe | null>(null)
const activeTab = ref('info')

const stageTargets = ref<RecipeStageTarget[]>([])
const ionTargets = ref<RecipeIonTarget[]>([])
const targetsLoading = ref(false)

const growthStages = ref<GrowthStage[]>([])
const cropVarieties = ref<CropVariety[]>([])

const growthStageLabelById = computed(() =>
  buildIdLabelMap(growthStages.value, s => s.id, growthStageLabel, '阶段')
)

const varietyLabelById = computed(() => {
  const map: Record<number, string> = {}
  for (const v of cropVarieties.value) {
    map[v.id] = `${v.name} (${v.code})`
  }
  return map
})

function growthStageName(stageId?: number | null) {
  if (!stageId) return '-'
  return growthStageLabelById.value[stageId] || fallbackIdLabel('阶段', stageId)
}

function cropVarietyName(varietyId?: number | null) {
  if (!varietyId) return '-'
  return varietyLabelById.value[varietyId] || fallbackIdLabel('作物', varietyId)
}

const statusLabel = computed(() => {
  if (!recipe.value) return ''
  const map: Record<string, string> = { DRAFT: '草稿', ACTIVE: '已发布', ARCHIVED: '已归档' }
  return map[recipe.value.status] || recipe.value.status
})

const statusTagType = computed(() => {
  if (!recipe.value) return 'info'
  const map: Record<string, string> = { DRAFT: 'warning', ACTIVE: 'success', ARCHIVED: 'info' }
  return map[recipe.value.status] || 'info'
})

async function loadReferenceData() {
  try {
    const [stages, varieties] = await Promise.allSettled([
      cropApi.getGrowthStages({ page_size: LARGE_PAGE_SIZE }),
      cropApi.getCropVarieties({ page_size: LARGE_PAGE_SIZE })
    ])
    if (stages.status === 'fulfilled') growthStages.value = stages.value.items
    if (varieties.status === 'fulfilled') cropVarieties.value = varieties.value.items
  } catch { /* ignore */ }
}

async function loadRecipe() {
  if (!recipeId.value) {
    errorMsg.value = '无效的配方 ID'
    return
  }
  try {
    recipe.value = await recipeApi.getRecipeDetail(recipeId.value)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载配方失败'
    errorMsg.value = msg
  }
}

async function loadTargets() {
  if (!recipeId.value) return
  targetsLoading.value = true
  try {
    const data = await recipeApi.getRecipeTargets(recipeId.value)
    stageTargets.value = data.stage_targets
    ionTargets.value = data.ion_targets
  } catch {
    stageTargets.value = []
    ionTargets.value = []
  } finally {
    targetsLoading.value = false
  }
}

async function loadAllData() {
  errorMsg.value = ''
  loading.value = true
  await Promise.allSettled([loadRecipe(), loadReferenceData(), loadTargets()])
  loading.value = false
}

// Edit dialog
const editDialogVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editSubmitting = ref(false)

const editForm = reactive({
  name: '',
  crop_variety_id: undefined as number | undefined,
  version: '',
  description: '' as string,
  effective_from: null as string | null,
  effective_to: null as string | null
})

const editFormRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

function openEditDialog() {
  if (!recipe.value) return
  editForm.name = recipe.value.name
  editForm.crop_variety_id = recipe.value.crop_variety_id ?? undefined
  editForm.version = recipe.value.version
  editForm.description = recipe.value.description || ''
  editForm.effective_from = recipe.value.effective_from || null
  editForm.effective_to = recipe.value.effective_to || null
  editDialogVisible.value = true
}

async function handleEditSubmit() {
  if (!editFormRef.value) return
  try {
    await editFormRef.value.validate()
  } catch { return }
  if (!recipeId.value) return

  editSubmitting.value = true
  try {
    await recipeApi.updateRecipe(recipeId.value, {
      name: editForm.name,
      crop_variety_id: editForm.crop_variety_id,
      version: editForm.version || undefined,
      description: editForm.description || undefined,
      effective_from: editForm.effective_from || undefined,
      effective_to: editForm.effective_to || undefined
    })
    ElMessage.success('配方已更新')
    editDialogVisible.value = false
    await loadRecipe()
    await loadTargets()
  } catch { /* handled by interceptor */ }
  finally { editSubmitting.value = false }
}

// Publish dialog
const publishDialogVisible = ref(false)
const publishFormRef = ref<FormInstance>()
const publishSubmitting = ref(false)
const publishVersion = ref('')

function openPublishDialog() {
  publishVersion.value = recipe.value?.version || ''
  publishDialogVisible.value = true
}

async function handlePublish() {
  if (!publishVersion.value.trim()) {
    ElMessage.error('请输入发布版本')
    return
  }
  publishSubmitting.value = true
  try {
    await recipeApi.publishRecipe(recipeId.value, { version: publishVersion.value })
    ElMessage.success('配方已发布')
    publishDialogVisible.value = false
    await loadRecipe()
  } catch { /* handled by interceptor */ }
  finally { publishSubmitting.value = false }
}

// Delete
async function handleDelete() {
  await ElMessageBox.confirm('确认删除该配方？', '提示', { type: 'warning' })
  try {
    await recipeApi.deleteRecipe(recipeId.value)
    ElMessage.success('已删除')
    router.push('/nutrient/recipes')
  } catch { /* handled by interceptor */ }
}

onMounted(() => {
  loadAllData()
})
</script>

<style scoped lang="scss">
.nutrient-recipe-detail-page {
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
  .section-card { margin-bottom: 0; }
  .card-title { font-size: 15px; font-weight: 600; }
  .stats-row {
    margin-bottom: 16px;
    .stat-card {
      text-align: center;
      padding: 20px 16px;
      background: var(--bg-card);
      border-radius: var(--radius-md);
      box-shadow: var(--shadow-card);
      .stat-value { font-size: 28px; font-weight: 700; color: var(--color-primary); }
      .stat-label { font-size: 13px; color: var(--color-text-secondary); margin-top: 4px; }
    }
  }
  .hint { color: var(--color-text-secondary); font-size: 12px; }
}
</style>
