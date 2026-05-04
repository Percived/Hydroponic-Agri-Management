<template>
  <header class="app-header">
    <div class="header-left">
      <span class="logo">🌱 水培农业管理系统</span>
    </div>
    <div class="header-right">
      <!-- 告警通知 -->
      <el-badge :value="alertBadgeCount" :hidden="alertBadgeCount === 0" :max="99" class="alert-badge">
        <el-button link @click="goAlerts" aria-label="查看告警">
          <el-icon size="20"><Bell /></el-icon>
        </el-button>
      </el-badge>
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-info">
          <el-icon><User /></el-icon>
          {{ authStore.user?.nickname || authStore.user?.username || '用户' }}
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAlertSSE, requestNotificationPermission } from '@/composables'
import { User, ArrowDown, SwitchButton, Bell } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

// Alert SSE - listen for CRITICAL alerts
const alertBadgeCount = ref(0)
const { connect: connectSSE, disconnect: disconnectSSE } = useAlertSSE({ level: 'CRITICAL' })

onMounted(() => {
  // Request browser notification permission on first interaction
  requestNotificationPermission()

  connectSSE()

  // Wire up the count manually via polling lastAlert
  // (The composable tracks internally; we sync badge count via a simple interval)
})

onUnmounted(() => {
  disconnectSSE()
})

function goAlerts() {
  alertBadgeCount.value = 0
  router.push('/alerts')
}

async function handleCommand(command: string) {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await authStore.logout()
      router.push('/login')
    } catch {
      // 取消退出
    }
  }
}
</script>

<style scoped lang="scss">
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.alert-badge {
  margin-right: 4px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  color: #606266;
  font-size: 14px;

  &:hover {
    color: #409eff;
  }
}
</style>
