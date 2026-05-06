<template>
  <aside class="app-sidebar">
    <el-menu
      :default-active="activeMenu"
      router
      background-color="transparent"
      class="sidebar-menu"
    >
      <el-menu-item index="/">
        <el-icon><HomeFilled /></el-icon>
        <span>首页</span>
      </el-menu-item>
      <el-sub-menu index="assets">
        <template #title>
          <el-icon><Monitor /></el-icon>
          <span>资产中心</span>
        </template>
        <el-menu-item index="/assets/sensor-devices">传感器设备</el-menu-item>
        <el-menu-item index="/assets/actuator-devices">执行器设备</el-menu-item>
        <el-menu-item v-if="isAdmin" index="/assets/greenhouses">温室管理</el-menu-item>
        <el-menu-item v-if="canOperate" index="/assets/growing-zones">种植区管理</el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="collection">
        <template #title>
          <el-icon><TrendCharts /></el-icon>
          <span>采集中心</span>
        </template>
        <el-menu-item index="/collection/realtime">实时曲线</el-menu-item>
        <el-menu-item index="/collection/history">历史趋势</el-menu-item>
        <el-menu-item index="/collection/batch-trends">批次趋势</el-menu-item>
      </el-sub-menu>
      <el-sub-menu v-if="canControl" index="strategy">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>策略控制</span>
        </template>
        <el-menu-item index="/strategy/policies">控制策略</el-menu-item>
        <el-menu-item index="/strategy/climate">气候联动</el-menu-item>
        <el-menu-item index="/strategy/commands">指令下发</el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="nutrient">
        <template #title>
          <el-icon><CoffeeCup /></el-icon>
          <span>营养液管理</span>
        </template>
        <el-menu-item index="/nutrient/tanks">营养液槽</el-menu-item>
        <el-menu-item index="/nutrient/ion-tests">离子检测</el-menu-item>
        <el-menu-item index="/nutrient/recipes">营养配方</el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="alerts">
        <template #title>
          <el-icon><Bell /></el-icon>
          <span>告警处置</span>
        </template>
        <el-menu-item index="/alerts/list">告警列表</el-menu-item>
        <el-menu-item index="/alerts/timeline">时间线</el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="batches">
        <template #title>
          <el-icon><Grid /></el-icon>
          <span>批次管理</span>
        </template>
        <el-menu-item index="/batches/ledger">批次台账</el-menu-item>
        <el-menu-item index="/batches/harvest">采收记录</el-menu-item>
        <el-menu-item v-if="canOperate" index="/batches/stage-plans">阶段计划</el-menu-item>
        <el-menu-item index="/batches/review">批次复盘</el-menu-item>
      </el-sub-menu>
      <el-menu-item index="/pest/observations">
        <el-icon><Warning /></el-icon>
        <span>病虫害观察</span>
      </el-menu-item>
      <el-menu-item index="/energy/records">
        <el-icon><DataLine /></el-icon>
        <span>能耗记录</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/users">
        <el-icon><User /></el-icon>
        <span>用户管理</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/audit-logs">
        <el-icon><Document /></el-icon>
        <span>审计日志</span>
      </el-menu-item>
      <el-menu-item v-if="isAdmin" index="/settings/notification-channels">
        <el-icon><Message /></el-icon>
        <span>通知渠道</span>
      </el-menu-item>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Grid, TrendCharts, Setting, Bell, User, Document, HomeFilled, CoffeeCup, Warning, DataLine, Message } from '@element-plus/icons-vue'
import { usePermission } from '@/composables/usePermission'
import { Role } from '@/types'

const route = useRoute()
const { canControlDevice, hasRole } = usePermission()

const activeMenu = computed(() => {
  return route.path
})

const canControl = computed(() => canControlDevice())
const canOperate = computed(() => hasRole(Role.ADMIN) || hasRole(Role.OPERATOR))
const isAdmin = computed(() => hasRole(Role.ADMIN))
</script>

<style scoped lang="scss">
.app-sidebar {
  width: 200px;
  background: #fff;
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  overscroll-behavior: contain;
}

.sidebar-menu {
  border-right: none;
  height: 100%;
  padding: 8px 0;

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    margin: 2px 8px;
    border-radius: 8px;
    transition: background-color var(--transition-fast), color var(--transition-fast);

    &:hover {
      background-color: var(--color-primary-bg-light);
      color: var(--color-primary);
    }
  }

  :deep(.el-menu-item.is-active) {
    background: linear-gradient(135deg, var(--color-primary), var(--color-primary-light));
    color: #fff;
    border-radius: 8px;

    .el-icon {
      color: #fff;
    }
  }

  :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
    color: var(--color-primary);
  }
}
</style>
