<template>
  <aside class="app-sidebar">
    <el-menu
      :default-active="activeMenu"
      router
      class="sidebar-menu"
    >
      <el-menu-item index="/">
        <el-icon><HomeFilled /></el-icon>
        <span>首页</span>
      </el-menu-item>
      <el-menu-item index="/devices">
        <el-icon><Monitor /></el-icon>
        <span>设备管理</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/greenhouses">
        <el-icon><OfficeBuilding /></el-icon>
        <span>温室管理</span>
      </el-menu-item>
      <el-menu-item index="/device-groups">
        <el-icon><Grid /></el-icon>
        <span>设备分组</span>
      </el-menu-item>
      <el-sub-menu index="telemetry">
        <template #title>
          <el-icon><TrendCharts /></el-icon>
          <span>遥测数据</span>
        </template>
        <el-menu-item index="/telemetry/realtime">实时数据</el-menu-item>
        <el-menu-item index="/telemetry/history">历史数据</el-menu-item>
      </el-sub-menu>
      <el-sub-menu v-if="canControl" index="controls">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>控制中心</span>
        </template>
        <el-menu-item index="/controls/commands">控制命令</el-menu-item>
        <el-menu-item index="/controls/rules">控制规则</el-menu-item>
      </el-sub-menu>
      <el-menu-item index="/alerts">
        <el-icon><Bell /></el-icon>
        <span>告警中心</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/users">
        <el-icon><User /></el-icon>
        <span>用户管理</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/audit-logs">
        <el-icon><Document /></el-icon>
        <span>审计日志</span>
      </el-menu-item>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Grid, TrendCharts, Setting, Bell, User, Document, HomeFilled, OfficeBuilding } from '@element-plus/icons-vue'
import { usePermission } from '@/composables/usePermission'
import { Role } from '@/types'

const route = useRoute()
const { canControlDevice, hasRole } = usePermission()

const activeMenu = computed(() => {
  return route.path
})

const canControl = computed(() => canControlDevice())
const isAdmin = computed(() => hasRole(Role.ADMIN))
</script>

<style scoped lang="scss">
.app-sidebar {
  width: 200px;
  background: #fff;
  border-right: 1px solid #e6e6e6;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.sidebar-menu {
  border-right: none;
  height: 100%;
}
</style>
