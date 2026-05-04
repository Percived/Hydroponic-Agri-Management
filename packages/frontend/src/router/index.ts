import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { Role } from '@/types'
import { getToken } from '@/utils/storage'

// 路由配置
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { requiresAuth: false, title: '登录' }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/index.vue'),
    meta: { requiresAuth: true, title: '首页' }
  },
  {
    path: '/devices',
    name: 'DeviceList',
    component: () => import('@/views/devices/list.vue'),
    meta: { requiresAuth: true, title: '设备列表' }
  },
  {
    path: '/devices/:id',
    name: 'DeviceDetail',
    component: () => import('@/views/devices/detail.vue'),
    meta: { requiresAuth: true, title: '设备详情' }
  },
  {
    path: '/greenhouses',
    name: 'Greenhouses',
    component: () => import('@/views/greenhouses/index.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN], title: '温室管理' }
  },
  {
    path: '/device-groups',
    name: 'DeviceGroups',
    component: () => import('@/views/device-groups/index.vue'),
    meta: { requiresAuth: true, title: '设备分组' }
  },
  {
    path: '/telemetry/realtime',
    name: 'TelemetryRealtime',
    component: () => import('@/views/telemetry/realtime.vue'),
    meta: { requiresAuth: true, title: '实时数据' }
  },
  {
    path: '/telemetry/history',
    name: 'TelemetryHistory',
    component: () => import('@/views/telemetry/history.vue'),
    meta: { requiresAuth: true, title: '历史数据' }
  },
  {
    path: '/alerts',
    name: 'Alerts',
    component: () => import('@/views/alerts/index.vue'),
    meta: { requiresAuth: true, title: '告警中心' }
  },
  {
    path: '/controls/commands',
    name: 'ControlCommands',
    component: () => import('@/views/controls/commands.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '控制命令' }
  },
  {
    path: '/controls/rules',
    name: 'ControlRules',
    component: () => import('@/views/controls/rules.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '控制规则' }
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/users/index.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN], title: '用户管理' }
  },
  {
    path: '/audit-logs',
    name: 'AuditLogs',
    component: () => import('@/views/audit-logs/index.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN], title: '审计日志' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { requiresAuth: false, title: '页面不存在' }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to) => {
  // 设置页面标题
  document.title = (to.meta.title as string) || '水培农业管理系统'

  // 不需要认证的页面直接放行
  if (!to.meta.requiresAuth) {
    return true
  }

  // 检查是否登录
  const token = getToken()
  if (!token) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }

  // 角色权限检查
  const requiredRoles = to.meta.roles as Role[] | undefined
  if (requiredRoles && requiredRoles.length > 0) {
    // 动态获取用户角色
    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()
    if (!authStore.hasAnyRole(requiredRoles)) {
      return { name: 'DeviceList' }
    }
  }

  return true
})

export default router