<template>
  <div class="batch-event-overlay">
    <el-timeline>
      <el-timeline-item
        v-for="(event, idx) in events"
        :key="`${event.type}-${idx}-${event.time}`"
        :timestamp="event.time"
        :type="event.type === 'alert' ? 'warning' : 'primary'"
      >
        {{ event.label }}
      </el-timeline-item>
    </el-timeline>
    <el-empty v-if="events.length === 0" description="当前时间窗无事件" />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  events: Array<{
    type: 'alert' | 'control'
    time: string
    label: string
  }>
}>()
</script>

<style scoped lang="scss">
.batch-event-overlay {
  max-height: 360px;
  overflow: auto;
}
</style>
