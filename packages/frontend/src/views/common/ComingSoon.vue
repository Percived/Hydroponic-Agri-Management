<template>
  <div class="coming-soon-page">
    <el-result icon="info" :title="title" :sub-title="subTitle">
      <template #extra>
        <el-tag type="warning">规划中</el-tag>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const title = computed(() => (route.meta.title as string) || '页面建设中')
const subTitle = computed(() => (route.meta.description as string) || '该能力位于后续迭代阶段，当前版本先提供入口占位。')
</script>

<style scoped lang="scss">
.coming-soon-page {
  background: var(--bg-card);
  border-radius: var(--radius-md);
  padding: 24px;
}
</style>
