<template>
  <el-row :gutter="16">
    <el-col :span="16">
      <el-card>
        <template #header>环境曲线 + 事件叠加</template>
        <metric-trend-chart :series="series" :events="events" />
      </el-card>
    </el-col>
    <el-col :span="8">
      <el-card>
        <template #header>事件时间线</template>
        <batch-event-overlay :events="timelineEvents" />
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import MetricTrendChart from '@/components/charts/MetricTrendChart.vue'
import BatchEventOverlay from '@/components/charts/BatchEventOverlay.vue'

defineProps<{
  series: Array<{ name: string; data: Array<{ time: string; value: number }> }>
  events: Array<{ time: string; value: number; label: string; eventType: 'alert' | 'control' }>
  timelineEvents: Array<{ type: 'alert' | 'control'; time: string; label: string }>
}>()
</script>
