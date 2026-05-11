<template>
  <el-form :model="editorData" label-width="120px">
    <el-form-item label="生长阶段" required>
      <el-select v-model="editorData.growth_stage_id" filterable style="width: 100%">
        <el-option v-for="s in stages" :key="s.id" :label="`${s.name} (${s.code})`" :value="s.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="开始时间" required>
      <el-date-picker v-model="editorData.stage_start_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" />
    </el-form-item>
    <el-form-item label="结束时间" required>
      <el-date-picker v-model="editorData.stage_end_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" />
    </el-form-item>
    <el-form-item label="目标EC下限">
      <el-input-number v-model="editorData.target_ec_min" :min="0" :precision="4" />
    </el-form-item>
    <el-form-item label="目标EC上限">
      <el-input-number v-model="editorData.target_ec_max" :min="0" :precision="4" />
    </el-form-item>
    <el-form-item label="目标pH下限">
      <el-input-number v-model="editorData.target_ph_min" :min="0" :max="14" :precision="4" />
    </el-form-item>
    <el-form-item label="目标pH上限">
      <el-input-number v-model="editorData.target_ph_max" :min="0" :max="14" :precision="4" />
    </el-form-item>
    <el-form-item label="营养配方">
      <el-select v-model="editorData.recipe_id" filterable clearable placeholder="可选" style="width: 100%">
        <el-option v-for="r in recipes" :key="r.id" :label="`${r.name} (${r.recipe_code})`" :value="r.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="控制策略">
      <el-select v-model="editorData.policy_id" filterable clearable placeholder="可选" style="width: 100%">
        <el-option v-for="p in policies" :key="p.id" :label="`${p.name} (${p.policy_code})`" :value="p.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="气候Profile">
      <el-select v-model="editorData.climate_profile_id" filterable clearable placeholder="可选" style="width: 100%">
        <el-option v-for="p in profiles" :key="p.id" :label="`${p.name} (${p.code})`" :value="p.id" />
      </el-select>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { climateApi, cropApi, policyApi, recipeApi } from '@/api'
import type { ClimateProfile, ControlPolicy, CreateBatchStagePlanRequest, GrowthStage, NutrientRecipe } from '@/types'

const editorData = defineModel<CreateBatchStagePlanRequest>({ required: true })

const stages = ref<GrowthStage[]>([])
const recipes = ref<NutrientRecipe[]>([])
const policies = ref<ControlPolicy[]>([])
const profiles = ref<ClimateProfile[]>([])

onMounted(async () => {
  const res = await cropApi.getGrowthStages({ page_size: 200 })
  stages.value = res.items

  const [recipeRes, policyRes, climateRes] = await Promise.all([
    recipeApi.getRecipes({ page: 1, page_size: 200 }),
    policyApi.getPolicies({ page: 1, page_size: 200 }),
    climateApi.getClimateProfiles({ page: 1, page_size: 200 })
  ])
  recipes.value = recipeRes.items
  policies.value = policyRes.items
  profiles.value = climateRes.items
})
</script>
