<template>
  <div class="batch-stage-page">
    <div class="page-header">
      <h1 class="page-title">阶段计划</h1>
    </div>

    <div class="panel">
      <el-form :inline="true">
        <el-form-item label="批次">
          <el-select v-model="selectedBatchId" filterable placeholder="选择批次" style="width: 300px" @change="loadStages">
            <el-option v-for="batch in batches" :key="batch.id" :label="`${batch.batch_no} (#${batch.id})`" :value="batch.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :disabled="!selectedBatchId" @click="openCreateDialog">新增阶段</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="panel">
      <el-alert
        v-if="conflictMessage"
        type="warning"
        :closable="false"
        show-icon
        :title="conflictMessage"
        class="conflict-alert"
      />
      <el-table :data="stages" stripe v-loading="loading" :row-class-name="stageRowClass">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="生长阶段" width="160">
          <template #default="{ row }">{{ growthStageName(row.growth_stage_id) }}</template>
        </el-table-column>
        <el-table-column label="配方" min-width="180">
          <template #default="{ row }">{{ recipeName(row.recipe_id) }}</template>
        </el-table-column>
        <el-table-column label="策略" min-width="180">
          <template #default="{ row }">{{ policyName(row.policy_id) }}</template>
        </el-table-column>
        <el-table-column label="气候Profile" min-width="200">
          <template #default="{ row }">{{ climateProfileName(row.climate_profile_id) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="stageStatusTag(row)" size="small">{{ stageStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stage_start_at" label="开始时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.stage_start_at) }}</template>
        </el-table-column>
        <el-table-column prop="stage_end_at" label="结束时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.stage_end_at) }}</template>
        </el-table-column>
        <el-table-column label="EC目标" width="120">
          <template #default="{ row }">{{ toRange(row.target_ec_min, row.target_ec_max) }}</template>
        </el-table-column>
        <el-table-column label="pH目标" width="120">
          <template #default="{ row }">{{ toRange(row.target_ph_min, row.target_ph_max) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="removeStage(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="editorVisible" :title="editingStageId ? '编辑阶段计划' : '新增阶段计划'" width="760px">
      <stage-plan-editor v-model="editorData" />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitStage">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { climateApi, cropApi, policyApi, recipeApi } from '@/api'
import StagePlanEditor from '@/components/batch/StagePlanEditor.vue'
import { formatDateTime } from '@/utils/format'
import { buildIdLabelMap, fallbackIdLabel, growthStageLabel } from '@/utils/labels'
import type { BatchStagePlan, ClimateProfile, ControlPolicy, CreateBatchStagePlanRequest, CropBatch, GrowthStage, NutrientRecipe } from '@/types'

const loading = ref(false)
const submitLoading = ref(false)
const batches = ref<CropBatch[]>([])
const selectedBatchId = ref<number>()
const stages = ref<BatchStagePlan[]>([])
const growthStages = ref<GrowthStage[]>([])
const recipes = ref<NutrientRecipe[]>([])
const policies = ref<ControlPolicy[]>([])
const climateProfiles = ref<ClimateProfile[]>([])

const growthStageLabelById = computed(() =>
  buildIdLabelMap(growthStages.value, s => s.id, growthStageLabel, '阶段')
)

function growthStageName(stageId?: number | null) {
  if (!stageId) return '-'
  return growthStageLabelById.value[stageId] || fallbackIdLabel('阶段', stageId)
}

const recipeLabelById = computed(() =>
  buildIdLabelMap(recipes.value, r => r.id, r => `${r.name} (${r.recipe_code})`, '配方')
)

function recipeName(recipeId?: number | null) {
  if (!recipeId) return '-'
  return recipeLabelById.value[recipeId] || fallbackIdLabel('配方', recipeId)
}

const policyLabelById = computed(() =>
  buildIdLabelMap(policies.value, p => p.id, p => `${p.name} (${p.policy_code})`, '策略')
)

function policyName(policyId?: number | null) {
  if (!policyId) return '-'
  return policyLabelById.value[policyId] || fallbackIdLabel('策略', policyId)
}

const climateProfileLabelById = computed(() =>
  buildIdLabelMap(climateProfiles.value, p => p.id, p => `${p.name} (${p.code})`, 'Profile')
)

function climateProfileName(profileId?: number | null) {
  if (!profileId) return '-'
  return climateProfileLabelById.value[profileId] || fallbackIdLabel('Profile', profileId)
}

const editorVisible = ref(false)
const editingStageId = ref<number>()
const editorData = ref<CreateBatchStagePlanRequest>({
  batch_id: 0,
  growth_stage_id: 1,
  stage_start_at: new Date().toISOString(),
  stage_end_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
  target_ec_min: 1,
  target_ec_max: 2,
  target_ph_min: 5.5,
  target_ph_max: 6.5,
  recipe_id: undefined,
  policy_id: undefined,
  climate_profile_id: undefined
})

const conflictMessage = computed(() => {
  const sorted = [...stages.value].sort((a, b) => new Date(a.stage_start_at).getTime() - new Date(b.stage_start_at).getTime())
  for (let i = 1; i < sorted.length; i++) {
    const prevEnd = new Date(sorted[i - 1].stage_end_at).getTime()
    const currStart = new Date(sorted[i].stage_start_at).getTime()
    if (currStart < prevEnd) {
      return `阶段时间窗冲突：阶段 ${growthStageName(sorted[i - 1].growth_stage_id)} 与 ${growthStageName(sorted[i].growth_stage_id)} 存在重叠。`
    }
  }
  return ''
})

async function initBatches() {
  const result = await cropApi.getBatches({ page: 1, page_size: 200 })
  batches.value = result.items
  if (!selectedBatchId.value && batches.value.length > 0) {
    selectedBatchId.value = batches.value[0].id
    await loadStages()
  }
}

async function loadGrowthStages() {
  try {
    const res = await cropApi.getGrowthStages({ page: 1, page_size: 200 })
    growthStages.value = res.items
  } catch {
    growthStages.value = []
  }
}

async function loadRecipes() {
  try {
    const res = await recipeApi.getRecipes({ page: 1, page_size: 200 })
    recipes.value = res.items
  } catch {
    recipes.value = []
  }
}

async function loadPolicies() {
  try {
    const res = await policyApi.getPolicies({ page: 1, page_size: 200 })
    policies.value = res.items
  } catch {
    policies.value = []
  }
}

async function loadClimateProfiles() {
  try {
    const res = await climateApi.getClimateProfiles({ page: 1, page_size: 200 })
    climateProfiles.value = res.items
  } catch {
    climateProfiles.value = []
  }
}

async function loadStages() {
  if (!selectedBatchId.value) return
  loading.value = true
  try {
    const result = await cropApi.getBatchStagePlans({ batch_id: selectedBatchId.value })
    stages.value = (result.items || []).sort(
      (a, b) => new Date(a.stage_start_at).getTime() - new Date(b.stage_start_at).getTime()
    )
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingStageId.value = undefined
  editorData.value = {
    batch_id: selectedBatchId.value || 0,
    growth_stage_id: 1,
    stage_start_at: new Date().toISOString(),
    stage_end_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    target_ec_min: 1,
    target_ec_max: 2,
    target_ph_min: 5.5,
    target_ph_max: 6.5,
    recipe_id: undefined,
    policy_id: undefined,
    climate_profile_id: undefined
  }
  editorVisible.value = true
}

function openEditDialog(stage: BatchStagePlan) {
  editingStageId.value = stage.id
  editorData.value = {
    batch_id: stage.batch_id,
    growth_stage_id: stage.growth_stage_id,
    recipe_id: stage.recipe_id ?? undefined,
    policy_id: stage.policy_id ?? undefined,
    climate_profile_id: stage.climate_profile_id ?? undefined,
    stage_start_at: stage.stage_start_at,
    stage_end_at: stage.stage_end_at,
    target_ec_min: stage.target_ec_min ?? undefined,
    target_ec_max: stage.target_ec_max ?? undefined,
    target_ph_min: stage.target_ph_min ?? undefined,
    target_ph_max: stage.target_ph_max ?? undefined
  }
  editorVisible.value = true
}

function validateStageInput(payload: CreateBatchStagePlanRequest) {
  if (!payload.growth_stage_id || !payload.stage_start_at || !payload.stage_end_at) return '请填写阶段ID和时间窗'
  const start = new Date(payload.stage_start_at).getTime()
  const end = new Date(payload.stage_end_at).getTime()
  if (start >= end) return '阶段结束时间必须晚于开始时间'
  const overlap = stages.value.some((s) => {
    if (editingStageId.value && s.id === editingStageId.value) return false
    const sStart = new Date(s.stage_start_at).getTime()
    const sEnd = new Date(s.stage_end_at).getTime()
    return Math.max(start, sStart) < Math.min(end, sEnd)
  })
  if (overlap) return '阶段时间窗与现有阶段冲突'
  return ''
}

async function submitStage() {
  if (!selectedBatchId.value) return
  const validation = validateStageInput(editorData.value)
  if (validation) {
    ElMessage.error(validation)
    return
  }

  submitLoading.value = true
  try {
    if (editingStageId.value) {
      await cropApi.updateBatchStagePlan(editingStageId.value, editorData.value)
      ElMessage.success('阶段计划已更新')
    } else {
      await cropApi.createBatchStagePlan(editorData.value)
      ElMessage.success('阶段计划已创建')
    }
    editorVisible.value = false
    await loadStages()
  } finally {
    submitLoading.value = false
  }
}

async function removeStage(stageId: number) {
  await ElMessageBox.confirm('确认删除该阶段计划？', '提示', { type: 'warning' })
  await cropApi.deleteBatchStagePlan(stageId)
  ElMessage.success('已删除')
  await loadStages()
}

function toRange(min?: number | null, max?: number | null) {
  if (min == null && max == null) return '-'
  return `${min ?? '-'} ~ ${max ?? '-'}`
}

// Stage status helpers
const now = new Date()

function stageStatusTag(stage: BatchStagePlan) {
  const start = new Date(stage.stage_start_at)
  const end = new Date(stage.stage_end_at)
  if (now < start) return 'info'
  if (now > end) return 'success'
  return ''
}

function stageStatusText(stage: BatchStagePlan) {
  const start = new Date(stage.stage_start_at)
  const end = new Date(stage.stage_end_at)
  if (now < start) return '未开始'
  if (now > end) return '已完成'
  return '进行中'
}

function stageRowClass({ row }: { row: BatchStagePlan }) {
  const start = new Date(row.stage_start_at)
  const end = new Date(row.stage_end_at)
  if (now < start) return 'stage-pending'
  if (now > end) return 'stage-completed'
  return 'stage-active'
}

onMounted(() => {
  initBatches()
  loadGrowthStages()
  loadRecipes()
  loadPolicies()
  loadClimateProfiles()
})
</script>

<style scoped lang="scss">
.batch-stage-page {
  .page-header {
    margin-bottom: 16px;
  }
  .page-title {
    margin: 0;
    font-size: 22px;
    font-weight: 700;
  }
  .panel {
    background: var(--bg-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-card);
    padding: 16px;
    margin-bottom: 12px;
  }
  .conflict-alert {
    margin-bottom: 10px;
  }
  :deep(.stage-completed) {
    background-color: rgba(103, 194, 58, 0.06);
  }
  :deep(.stage-active) {
    background-color: rgba(64, 158, 255, 0.06);
  }
  :deep(.stage-pending) {
    background-color: rgba(144, 147, 153, 0.04);
  }
}
</style>
