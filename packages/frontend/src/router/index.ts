import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { Role } from '@/types'
import { getToken } from '@/utils/storage'

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
  // ── 资产中心 ──
  {
    path: '/assets/sensor-devices',
    name: 'SensorDevices',
    component: () => import('@/views/devices/list.vue'),
    meta: { requiresAuth: true, title: '传感器设备' }
  },
  {
    path: '/assets/actuator-devices',
    name: 'ActuatorDevices',
    component: () => import('@/views/devices/list.vue'),
    meta: { requiresAuth: true, title: '执行器设备' }
  },
  {
    path: '/assets/greenhouses',
    name: 'Greenhouses',
    component: () => import('@/views/greenhouses/index.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN], title: '温室管理' }
  },
  {
    path: '/assets/growing-zones',
    name: 'GrowingZones',
    component: () => import('@/views/greenhouses/zones.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '种植区管理' }
  },
  // ── 采集中心 ──
  {
    path: '/collection/realtime',
    name: 'TelemetryRealtime',
    component: () => import('@/views/telemetry/realtime.vue'),
    meta: { requiresAuth: true, title: '实时曲线' }
  },
  {
    path: '/collection/history',
    name: 'TelemetryHistory',
    component: () => import('@/views/telemetry/history.vue'),
    meta: { requiresAuth: true, title: '历史趋势' }
  },
  {
    path: '/collection/batch-trends',
    name: 'BatchTrends',
    component: () => import('@/views/telemetry/batch-trends.vue'),
    meta: { requiresAuth: true, title: '批次趋势' }
  },
  // ── 策略控制 ──
  {
    path: '/strategy/policies',
    name: 'ControlPolicies',
    component: () => import('@/views/controls/rules.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '控制策略' }
  },
  {
    path: '/strategy/climate',
    name: 'ClimateProfiles',
    component: () => import('@/views/climate/index.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '气候联动' }
  },
  {
    path: '/strategy/commands',
    name: 'ControlCommands',
    component: () => import('@/views/controls/commands.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '指令下发' }
  },
  // ── 营养液 ──
  {
    path: '/nutrient/tanks',
    name: 'NutrientTanks',
    component: () => import('@/views/nutrient/tanks.vue'),
    meta: { requiresAuth: true, title: '营养液槽' }
  },
  {
    path: '/nutrient/ion-tests',
    name: 'IonTests',
    component: () => import('@/views/nutrient/ion-tests.vue'),
    meta: { requiresAuth: true, title: '离子检测' }
  },
  {
    path: '/nutrient/recipes',
    name: 'NutrientRecipes',
    component: () => import('@/views/recipes/index.vue'),
    meta: { requiresAuth: true, title: '营养配方' }
  },
  // ── 告警处置 ──
  {
    path: '/alerts/list',
    name: 'Alerts',
    component: () => import('@/views/alerts/index.vue'),
    meta: { requiresAuth: true, title: '告警列表' }
  },
  {
    path: '/alerts/timeline',
    name: 'AlertTimeline',
    component: () => import('@/views/alerts/timeline.vue'),
    meta: { requiresAuth: true, title: '告警时间线' }
  },
  // ── 批次管理 ──
  {
    path: '/batches/ledger',
    name: 'BatchLedger',
    component: () => import('@/views/batches/ledger.vue'),
    meta: { requiresAuth: true, title: '批次台账' }
  },
  {
    path: '/batches/harvest',
    name: 'HarvestRecords',
    component: () => import('@/views/batches/harvest.vue'),
    meta: { requiresAuth: true, title: '采收记录' }
  },
  {
    path: '/batches/stage-plans',
    name: 'BatchStagePlans',
    component: () => import('@/views/batches/stage-plans.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '阶段计划' }
  },
  {
    path: '/batches/review',
    name: 'BatchReview',
    component: () => import('@/views/batches/review.vue'),
    meta: { requiresAuth: true, title: '批次复盘' }
  },
  // ── 植保 ──
  {
    path: '/pest/observations',
    name: 'PestObservations',
    component: () => import('@/views/pest/observations.vue'),
    meta: { requiresAuth: true, title: '病虫害观察' }
  },
  // ── 能耗 ──
  {
    path: '/energy/records',
    name: 'EnergyRecords',
    component: () => import('@/views/energy/records.vue'),
    meta: { requiresAuth: true, title: '能耗记录' }
  },
  // ── Admin ──
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
    path: '/settings/notification-channels',
    name: 'NotificationChannels',
    component: () => import('@/views/settings/notification-channels.vue'),
    meta: { requiresAuth: true, roles: [Role.ADMIN], title: '通知渠道' }
  },
  // ── Legacy redirects ──
  { path: '/devices', redirect: '/assets/sensor-devices' },
  { path: '/devices/:id', redirect: '/assets/sensor-devices' },
  { path: '/greenhouses', redirect: '/assets/greenhouses' },
  { path: '/device-groups', redirect: '/assets/greenhouses' },
  { path: '/telemetry/realtime', redirect: '/collection/realtime' },
  { path: '/telemetry/history', redirect: '/collection/history' },
  { path: '/controls/commands', redirect: '/strategy/commands' },
  { path: '/controls/rules', redirect: '/strategy/policies' },
  { path: '/alerts', redirect: '/alerts/list' },
  { path: '/alerts/workflow', redirect: '/alerts/timeline' },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { requiresAuth: false, title: '页面不存在' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  document.title = (to.meta.title as string) || '水培农植信息管理系统'

  if (!to.meta.requiresAuth) {
    return true
  }

  const token = getToken()
  if (!token) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }

  const requiredRoles = to.meta.roles as Role[] | undefined
  if (requiredRoles && requiredRoles.length > 0) {
    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()
    if (!authStore.hasAnyRole(requiredRoles)) {
      return { name: 'SensorDevices' }
    }
  }

  return true
})

export default router
