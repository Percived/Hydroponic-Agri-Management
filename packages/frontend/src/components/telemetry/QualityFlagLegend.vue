<template>
  <div class="quality-legend">
    <el-tag v-for="item in items" :key="item.value" size="small" :type="item.type">{{ item.label }}</el-tag>
  </div>
</template>

<script setup lang="ts">
const items = [
  { value: 'normal', label: 'normal', type: 'success' },
  { value: 'missing', label: 'missing', type: 'info' },
  { value: 'out_of_range', label: 'out_of_range', type: 'warning' },
  { value: 'device_offline', label: 'device_offline', type: 'danger' }
] as const
</script>

<style scoped lang="scss">
.quality-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
