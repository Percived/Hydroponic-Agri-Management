<template>
  <el-form :model="formData" label-width="120px">
    <el-form-item label="批次编号" required>
      <el-input v-model="formData.batch_no" placeholder="如 BATCH-2026-001" />
    </el-form-item>
    <el-form-item label="温室" required>
      <el-select v-model="formData.greenhouse_id" filterable style="width: 100%">
        <el-option v-for="g in greenhouses" :key="g.id" :label="g.name" :value="g.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="作物品种" required>
      <el-select v-model="formData.crop_variety_id" filterable style="width: 100%">
        <el-option v-for="v in varieties" :key="v.id" :label="v.name" :value="v.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="定植密度(株/㎡)">
      <el-input-number v-model="formData.planting_density" :min="0" :precision="2" />
    </el-form-item>
    <el-form-item label="总株数">
      <el-input-number v-model="formData.total_plants" :min="0" />
    </el-form-item>
    <el-form-item label="开始时间">
      <el-date-picker v-model="formData.started_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" />
    </el-form-item>
    <el-form-item label="预计采收">
      <el-date-picker v-model="formData.expected_harvest_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" />
    </el-form-item>
    <el-form-item label="种植区">
      <el-select v-model="formData.growing_zone_id" clearable filterable style="width: 100%" placeholder="可选">
        <el-option v-for="z in zones" :key="z.id" :label="`${z.name} (${z.code})`" :value="z.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="配方版本">
      <el-input v-model="formData.recipe_version" placeholder="如 v1" />
    </el-form-item>
    <el-form-item label="策略版本">
      <el-input v-model="formData.policy_version" placeholder="如 v1" />
    </el-form-item>
    <el-form-item label="备注">
      <el-input v-model="formData.note" type="textarea" />
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { cropApi, greenhouseApi } from '@/api'
import type { CreateCropBatchRequest, CropVariety, Greenhouse, GrowingZone } from '@/types'

const formData = defineModel<CreateCropBatchRequest>({ required: true })

const greenhouses = ref<Greenhouse[]>([])
const varieties = ref<CropVariety[]>([])
const zones = ref<GrowingZone[]>([])

async function loadZones(greenhouseId?: number) {
  if (!greenhouseId) {
    zones.value = []
    return
  }
  try {
    const res = await greenhouseApi.getGreenhouseZones(greenhouseId)
    zones.value = res.items
  } catch {
    zones.value = []
  }
}

// Reload zones when greenhouse changes
watch(() => formData.value.greenhouse_id, (newId) => {
  formData.value.growing_zone_id = undefined
  loadZones(newId)
})

onMounted(async () => {
  const [ghResult, cvResult] = await Promise.all([
    greenhouseApi.getGreenhouses({ page_size: 200 }),
    cropApi.getCropVarieties({ page_size: 200 })
  ])
  greenhouses.value = ghResult.items
  varieties.value = cvResult.items

  // Load zones for initial greenhouse
  if (formData.value.greenhouse_id) {
    loadZones(formData.value.greenhouse_id)
  }
})
</script>
