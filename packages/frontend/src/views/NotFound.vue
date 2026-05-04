<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面不存在</p>
    <el-button type="primary" @click="goHome">返回首页</el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.push('/')
}
</script>

<style scoped lang="scss">
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #e8f8f4 0%, #f4fbf9 50%, #f2f3f5 100%);

  h1 {
    font-size: 96px;
    font-weight: 800;
    background: linear-gradient(135deg, var(--color-primary-dark), var(--color-primary-light));
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    margin-bottom: 16px;
    line-height: 1.1;
  }

  p {
    font-size: 18px;
    color: var(--color-text-regular);
    margin-bottom: 24px;
  }
}
</style>