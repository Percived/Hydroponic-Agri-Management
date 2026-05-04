<template>
  <div class="app-layout">
    <a class="skip-to-content" href="#main-content">跳到主内容</a>
    <AppHeader />
    <div class="app-main">
      <AppSidebar />
      <main id="main-content" class="app-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'
</script>

<style scoped lang="scss">
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.skip-to-content {
  position: absolute;
  top: -100%;
  left: 0;
  z-index: 9999;
  padding: 8px 16px;
  background: #409eff;
  color: #fff;
  text-decoration: none;
  font-size: 14px;
  border-radius: 0 0 4px 0;

  &:focus {
    top: 0;
  }
}

.app-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.app-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  overscroll-behavior: contain;
  background: #f5f7fa;
}
</style>
