<template>
  <el-form :model="localValue" label-width="110px">
    <el-form-item label="监控指标">
      <el-select v-model="localValue.metric_code" style="width: 100%">
        <el-option label="温度" value="TEMP" />
        <el-option label="湿度" value="HUMIDITY" />
        <el-option label="pH值" value="PH" />
        <el-option label="电导率" value="EC" />
        <el-option label="CO2" value="CO2" />
        <el-option label="光照" value="LIGHT" />
      </el-select>
    </el-form-item>
    <el-form-item label="比较运算">
      <el-select v-model="localValue.operator" style="width: 100%">
        <el-option label=">" value=">" />
        <el-option label=">=" value=">=" />
        <el-option label="<" value="<" />
        <el-option label="<=" value="<=" />
        <el-option label="==" value="==" />
      </el-select>
    </el-form-item>
    <el-form-item label="阈值">
      <el-input-number v-model="localValue.threshold_value" :precision="2" style="width: 100%" />
    </el-form-item>
    <el-form-item label="抖动区间">
      <el-input-number v-model="localValue.hysteresis" :precision="2" style="width: 100%" />
    </el-form-item>
    <el-form-item label="检测窗口(秒)">
      <el-input-number v-model="localValue.window_sec" :min="1" style="width: 100%" />
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PolicyCondition } from '@/types'

const props = defineProps<{
  modelValue: PolicyCondition
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: PolicyCondition): void
}>()

const localValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})
</script>
