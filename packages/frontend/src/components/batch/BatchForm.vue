<template>
  <el-form :model="formData" label-width="120px">
    <el-form-item label="批次编号" required>
      <el-input v-model="formData.batch_no" placeholder="如 BATCH-2026-001" />
    </el-form-item>
    <el-form-item label="温室" required>
      <el-input-number v-model="formData.greenhouse_id" :min="1" />
    </el-form-item>
    <el-form-item label="作物品种" required>
      <el-input-number v-model="formData.crop_variety_id" :min="1" />
    </el-form-item>
    <el-form-item label="定植密度">
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
      <el-input-number v-model="formData.growing_zone_id" :min="1" />
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
import type { CreateCropBatchRequest } from '@/types'

const formData = defineModel<CreateCropBatchRequest>({ required: true })
</script>
