<template>
  <el-tag :type="tagType" size="small">{{ label }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status?: string | null
}>()

const normalizedStatus = computed(() => (props.status || '').toUpperCase())

const label = computed(() => {
  switch (normalizedStatus.value) {
    case 'ONLINE':
      return '在线'
    case 'OFFLINE':
      return '离线'
    case 'ABNORMAL':
      return '异常'
    case 'ENABLED':
      return '启用'
    case 'DISABLED':
      return '禁用'
    default:
      return '未知'
  }
})

const tagType = computed(() => {
  switch (normalizedStatus.value) {
    case 'ONLINE':
    case 'ENABLED':
      return 'success'
    case 'OFFLINE':
      return 'info'
    case 'ABNORMAL':
    case 'DISABLED':
      return 'danger'
    default:
      return 'warning'
  }
})
</script>
