<template>
  <el-form :model="editorData" label-width="120px">
    <el-form-item label="生长阶段" required>
      <el-input-number v-model="editorData.growth_stage_id" :min="1" />
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
  </el-form>
</template>

<script setup lang="ts">
import type { CreateBatchStagePlanRequest } from '@/types'

const editorData = defineModel<CreateBatchStagePlanRequest>({ required: true })
</script>
