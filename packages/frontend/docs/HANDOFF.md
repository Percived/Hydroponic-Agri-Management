# 交接文档

最后更新: 2026-05-06
当前分支: main
当前重点: v0.7.0 架构落地 — 业务域闭环重构（资产/采集/策略/营养/告警/批次/植保/能耗/管理）

## 0. 近期变更（v0.7.0 架构重构）

### 全局架构变化

v0.7.0 完成了从 MVP 到业务域闭环的全量架构重构：

- **路由**：从旧版路径（`/devices`、`/greenhouses`、`/controls/rules`）迁移到业务域分组路径，共 26 条核心路由 + 11 条旧版重定向 + catch-all 404，涵盖 17 个 view 目录。
- **API 层**：20 个 API 模块，覆盖全部业务域，统一从 `src/api/index.ts` 导出。
- **类型层**：20 个类型模块，其中 `domain.ts` 为跨域共享枚举/接口，其余一一对应 API 模块。
- **侧边栏菜单**：从功能清单式重构为业务域分组式（资产中心、采集中心、策略控制、营养液管理、告警处置、批次管理、病虫害观察、能耗记录），ADMIN 专属项用 `v-if` 控制可见性。
- **设备管理**：传感器与执行器在 UI 中分拆为两个独立菜单项（`/assets/sensor-devices`、`/assets/actuator-devices`），路由均指向 `views/devices/list.vue`，通过路由路径区分类型。
- **控制域**：从单一的控制规则/命令模型演变为 Command + Policy 双模式，新增 `views/controls/rules.vue`（控制策略）、`views/climate/index.vue`（气候联动）。
- **新业务域**：气候联动（climate）、能耗记录（energy）、病虫害观察（pest）、营养液管理（nutrient）、营养配方（recipe）、批次管理（batches/crop）。
- **构建验证**：`npm run type-check` 通过。

### 文件数量快照

| 目录 | 文件数 | 说明 |
|------|--------|------|
| `src/api/` | 20 | alert, audit, auth, climate, control, crop, dashboard, device, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, request, telemetry, user |
| `src/types/` | 20 | alert, api, audit, climate, control, crop, dashboard, device, domain, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, telemetry, user |
| `src/views/` | 17 目录 | alerts, audit-logs, batches, climate, common, controls, dashboard, devices, energy, greenhouses, login, nutrient, pest, recipes, settings, telemetry, users |
| `src/components/` | 6 目录 | batch, charts, control, device, layout, telemetry |
| `src/stores/` | 3 文件 | auth.ts, greenhouse.ts, index.ts（2 个实际 store + 重导出） |
| `src/composables/` | 2 文件 | useAuth.ts, usePermission.ts |

---

## 1. 当前架构

### 1.1 路由表（26 + catch-all）

| 路由 | 名称 | View 目录 | 权限 |
|------|------|-----------|------|
| `/login` | Login | `login/` | 公开 |
| `/` | Dashboard | `dashboard/` | 全部 |
| `/assets/sensor-devices` | SensorDevices | `devices/` | 全部 |
| `/assets/actuator-devices` | ActuatorDevices | `devices/` | 全部 |
| `/assets/greenhouses` | Greenhouses | `greenhouses/` | ADMIN |
| `/assets/growing-zones` | GrowingZones | `greenhouses/` | ADMIN, OPERATOR |
| `/collection/realtime` | TelemetryRealtime | `telemetry/` | 全部 |
| `/collection/history` | TelemetryHistory | `telemetry/` | 全部 |
| `/collection/batch-trends` | BatchTrends | `telemetry/` | 全部 |
| `/strategy/policies` | ControlPolicies | `controls/` | ADMIN, OPERATOR |
| `/strategy/climate` | ClimateProfiles | `climate/` | ADMIN, OPERATOR |
| `/strategy/commands` | ControlCommands | `controls/` | ADMIN, OPERATOR |
| `/nutrient/tanks` | NutrientTanks | `nutrient/` | 全部 |
| `/nutrient/ion-tests` | IonTests | `nutrient/` | 全部 |
| `/nutrient/recipes` | NutrientRecipes | `recipes/` | 全部 |
| `/alerts/list` | Alerts | `alerts/` | 全部 |
| `/alerts/timeline` | AlertTimeline | `alerts/` | 全部 |
| `/batches/ledger` | BatchLedger | `batches/` | 全部 |
| `/batches/harvest` | HarvestRecords | `batches/` | 全部 |
| `/batches/stage-plans` | BatchStagePlans | `batches/` | ADMIN, OPERATOR |
| `/batches/review` | BatchReview | `batches/` | 全部 |
| `/pest/observations` | PestObservations | `pest/` | 全部 |
| `/energy/records` | EnergyRecords | `energy/` | 全部 |
| `/users` | Users | `users/` | ADMIN |
| `/audit-logs` | AuditLogs | `audit-logs/` | ADMIN |
| `/settings/notification-channels` | NotificationChannels | `settings/` | ADMIN |
| `/:pathMatch(.*)*` | NotFound | `views/` (根) | 公开 |

旧版重定向（保留向后兼容）：`/devices`、`/devices/:id`、`/greenhouses`、`/device-groups`、`/telemetry/realtime`、`/telemetry/history`、`/controls/commands`、`/controls/rules`、`/alerts`、`/alerts/workflow`。

### 1.2 侧边栏菜单结构（`components/layout/AppSidebar.vue`）

```
首页                                /
资产中心                            子菜单
  传感器设备                        /assets/sensor-devices
  执行器设备                        /assets/actuator-devices
  温室管理（ADMIN）                 /assets/greenhouses
  种植区管理（OPERATOR+）           /assets/growing-zones
采集中心                            子菜单
  实时曲线                          /collection/realtime
  历史趋势                          /collection/history
  批次趋势                          /collection/batch-trends
策略控制（OPERATOR+）               子菜单
  控制策略                          /strategy/policies
  气候联动                          /strategy/climate
  指令下发                          /strategy/commands
营养液管理                          子菜单
  营养液槽                          /nutrient/tanks
  离子检测                          /nutrient/ion-tests
  营养配方                          /nutrient/recipes
告警处置                            子菜单
  告警列表                          /alerts/list
  时间线                            /alerts/timeline
批次管理                            子菜单
  批次台账                          /batches/ledger
  采收记录                          /batches/harvest
  阶段计划（OPERATOR+）             /batches/stage-plans
  批次复盘                          /batches/review
病虫害观察                          /pest/observations
能耗记录                            /energy/records
用户管理（ADMIN）                   /users
审计日志（ADMIN）                   /audit-logs
通知渠道（ADMIN）                   /settings/notification-channels
```

菜单权限控制逻辑：
- `isAdmin`（`hasRole(Role.ADMIN)`）：控制温室管理、用户管理、审计日志、通知渠道可见性。
- `canOperate`（`hasRole(ADMIN) || hasRole(OPERATOR)`）：控制种植区管理、阶段计划可见性。
- `canControl`（`canControlDevice()`）：控制整个"策略控制"子菜单可见性。

### 1.3 API 模块清单

| 文件 | 导出名 | 对接后端模块 |
|------|--------|-------------|
| `request.ts` | `get`, `post`, `put`, `del` | 通用请求封装（Axios 实例） |
| `auth.ts` | `authApi` | 认证/登录 |
| `dashboard.ts` | `dashboardApi` | 仪表盘统计 |
| `device.ts` | `deviceApi` | 设备管理 |
| `greenhouse.ts` | `greenhouseApi` | 温室/种植区 |
| `telemetry.ts` | `telemetryApi` | 遥测数据 |
| `alert.ts` | `alertApi` | 告警 |
| `control.ts` | `commandApi` | 控制指令 |
| `user.ts` | `userApi` | 用户管理 |
| `audit.ts` | `auditApi` | 审计日志 |
| `notification.ts` | `notificationApi` | 通知渠道 |
| `metric.ts` | `metricApi` | 测点字典 |
| `crop.ts` | `cropApi` | 作物信息 |
| `recipe.ts` | `recipeApi` | 营养配方 |
| `policy.ts` | `policyApi` | 控制策略 |
| `climate.ts` | `climateApi` | 气候联动 |
| `nutrient.ts` | `nutrientApi` | 营养液槽/检测 |
| `energy.ts` | `energyApi` | 能耗记录 |
| `pest.ts` | `pestApi` | 病虫害观察 |
| `index.ts` | 汇总重导出 | 统一入口 |

### 1.4 类型模块清单

| 文件 | 对应模块 | 说明 |
|------|---------|------|
| `api.ts` | 全局 | `ApiResponse<T>`、分页、错误码 |
| `domain.ts` | 跨域 | `Role` 枚举等跨模块共享类型 |
| 其余 18 个 | 一一对应 API 模块 | 每个域的数据模型与 DTO |

### 1.5 组件目录

| 目录 | 说明 | 典型组件 |
|------|------|---------|
| `batch/` | 批次管理组件 | BatchForm, StagePlanEditor, BatchReviewBoard |
| `charts/` | 图表组件 | MetricTrendChart, BatchEventOverlay |
| `control/` | 控制引擎组件 | PolicyConditionEditor, ScheduleEditor, CommandStatusTimeline |
| `device/` | 设备组件 | DeviceStatusBadge, DeviceCard |
| `layout/` | 布局组件 | AppLayout, AppHeader, AppSidebar |
| `telemetry/` | 遥测组件 | QualityFlagLegend |

### 1.6 Pinia Stores

| Store | 文件 | 职责 |
|-------|------|------|
| `useAuthStore` | `auth.ts` | JWT token、用户信息、登录/登出、角色判断 |
| `useGreenhouseStore` | `greenhouse.ts` | 当前选中温室/种植区状态 |

### 1.7 请求流程

```
组件/Store → api/<module>.ts → request.ts (Axios 实例)
    ↓ 请求拦截器                     ↓ 响应拦截器
  附加 Authorization Bearer       code !== 0 → 业务异常
                                 HTTP 401 → clearAuth → /login
                                 HTTP 403 → 权限拒绝
```

---

## 2. 历史变更摘要

### Phase 0-5（v0.7.0 主体）
- 路由与菜单从功能清单式重构为 10 大业务域闭环分组。
- 传感器/执行器设备在 UI 分拆展示。
- 控制模型演进为 Command + Policy 双模式，新增气候联动域。
- 新增 6 个业务域：气候联动（climate）、能耗记录（energy）、病虫害观察（pest）、营养液管理（nutrient）、营养配方（recipe）、批次管理（batches/crop）。
- 告警闭环：列表增强（来源/处置状态筛选）、处置流程（指派/接管/关闭）、时间线聚合。
- 批次全生命周期：台账、采收记录、阶段计划、批次复盘。
- 采集中心三视图：实时曲线、历史趋势、批次趋势（三线联动）。
- 策略引擎：阈值策略 + 时序计划的 CRUD、冲突检测、发布流程，指令回执状态机与时间线。

### Phase 1-2（MVP 时期）
- 登录页、仪表盘、设备列表/详情、设备分组、遥测实时/历史、控制命令/规则、告警列表、用户管理、审计日志。
- 路由守卫（鉴权 + 角色权限）、404 页面。
- 通知渠道 CRUD、系统配置管理、批量设备操作。

---

## 3. 待办事项

1. **自动化测试**：引入 vitest + vue-test-utils，覆盖登录、设备列表、路由守卫、告警列表等关键路径。
2. **SSE 实时集成**：当前 composables 层有 SSE 实现，但未在视图中实际挂载消费，需接入仪表盘/设备详情/实时曲线页面。
3. **批量操作 UI 增强**：进度条反馈、分批状态展示、失败重试入口。
4. **移动端适配**：侧边栏切换为底部导航栏、表格切换为卡片列表。
5. **通用组件抽取**：提取 DeviceCard、TelemetryCard、StatusBadge、ConfirmDialog 等为共享组件。

---

## 4. 风险 / 阻碍

- **零自动化测试覆盖**：所有功能验证依赖手动操作，回归风险高。
- **SSE 未集成**：composables 中 `useAlertSSE`、`useTelemetrySSE` 已编写但未在任何视图中实际挂载消费。
- **无 CI/CD 流程**：前端未接入任何持续集成/持续部署管线。
- **移动端未适配**：仅依赖 Element Plus 默认响应式，未做专门适配。

---

## 5. 验证命令

```bash
cd packages/frontend

npm run type-check    # TypeScript 类型检查（vue-tsc --noEmit），当前通过
npm run build        # 生产构建（含类型检查），当前通过
npm run lint         # ESLint 检查
npm run dev          # 开发服务器启动（默认 localhost:5173，代理后端 /api）
```

---

## 6. 下个会话指引

1. 阅读 `docs/PROJECT_STATUS.md` 了解整体交付状态。
2. 阅读本文件（`docs/HANDOFF.md`）了解最近变更与当前架构。
3. 优先从待办事项 #1（自动化测试）或 #2（SSE 集成）开始。
4. 任何代码变更后，同步更新本文件。

---

## 7. 快速填写模板

- **日期**：2026-05-06
- **分支**：main
- **已完成范围**：v0.7.0 全量架构重构 — 26 路由 / 17 view 目录 / 20 API 模块 / 20 类型模块 / 业务域闭环菜单 / 6 大新业务域（climate, energy, pest, nutrient, recipe, crop）/ 策略引擎 / 告警闭环 / 批次全生命周期
- **待完成范围**：自动化测试、SSE 实时集成、批量 UI 优化、移动端适配、通用组件抽取
- **风险**：零自动化测试覆盖；SSE composables 未挂载消费；无 CI/CD
- **下个首要命令**：`npm run type-check`、`npm run build`
