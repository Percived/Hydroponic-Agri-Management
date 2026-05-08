# 项目状态

最后更新: 2026-05-07
负责人: 前端团队
版本: v0.8.0（SSE 实时推送就绪、仪表盘类型对齐）

## 1. 项目概述

基于 Vue 3 + TypeScript + Element Plus 构建的水培农植信息管理系统 Web 前端。
核心依赖：Vite、Pinia、Vue Router、Axios、ECharts。

## 2. 当前交付状态

总体评估：26 条路由全部实现，覆盖认证、仪表盘、资产中心、采集中心、策略控制、营养液、告警、批次、植保、能耗、管理、系统设置 12 个业务领域。TypeScript 类型检查除预存问题外通过（vue-tsc --noEmit），生产构建通过（npm run build）。

### v0.8.0 变更（2026-05-07）

- **SSE 实时推送就绪**：后端 SSE 端点（`/api/alerts/subscribe`、`/api/telemetry/subscribe`）已开通，前端 `useAlertSSE` 在 AppHeader 挂载，实时告警计数 + 浏览器通知可用。
- **仪表盘类型对齐**：`DashboardOverview` 新增 `devices_online/offline/total`、`device_type_distribution` 字段，与后端 v2.3.0 响应对齐；温室设备数修正为 `sensor_count + actuator_count`。

### 已完成路由表（26 条）

#### 认证
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/login` | Login | views/login/index.vue | 公开 | 登录 |

#### 仪表盘
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/` | Dashboard | views/dashboard/index.vue | 全部 | 首页 |

#### 资产中心
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/assets/sensor-devices` | SensorDevices | views/devices/list.vue | 全部 | 传感器设备 |
| `/assets/actuator-devices` | ActuatorDevices | views/devices/list.vue | 全部 | 执行器设备 |
| `/assets/greenhouses` | Greenhouses | views/greenhouses/index.vue | ADMIN | 温室管理 |
| `/assets/growing-zones` | GrowingZones | views/greenhouses/zones.vue | ADMIN/OPERATOR | 种植区管理 |

#### 采集中心
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/collection/realtime` | TelemetryRealtime | views/telemetry/realtime.vue | 全部 | 实时曲线 |
| `/collection/history` | TelemetryHistory | views/telemetry/history.vue | 全部 | 历史趋势 |
| `/collection/batch-trends` | BatchTrends | views/telemetry/batch-trends.vue | 全部 | 批次趋势 |

#### 策略控制
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/strategy/policies` | ControlPolicies | views/controls/rules.vue | ADMIN/OPERATOR | 控制策略 |
| `/strategy/climate` | ClimateProfiles | views/climate/index.vue | ADMIN/OPERATOR | 气候联动 |
| `/strategy/commands` | ControlCommands | views/controls/commands.vue | ADMIN/OPERATOR | 指令下发 |

#### 营养液
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/nutrient/tanks` | NutrientTanks | views/nutrient/tanks.vue | 全部 | 营养液槽 |
| `/nutrient/ion-tests` | IonTests | views/nutrient/ion-tests.vue | 全部 | 离子检测 |
| `/nutrient/recipes` | NutrientRecipes | views/recipes/index.vue | 全部 | 营养配方 |

#### 告警
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/alerts/list` | Alerts | views/alerts/index.vue | 全部 | 告警列表 |
| `/alerts/timeline` | AlertTimeline | views/alerts/timeline.vue | 全部 | 告警时间线 |

#### 批次管理
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/batches/ledger` | BatchLedger | views/batches/ledger.vue | 全部 | 批次台账 |
| `/batches/harvest` | HarvestRecords | views/batches/harvest.vue | 全部 | 采收记录 |
| `/batches/stage-plans` | BatchStagePlans | views/batches/stage-plans.vue | ADMIN/OPERATOR | 阶段计划 |
| `/batches/review` | BatchReview | views/batches/review.vue | 全部 | 批次复盘 |

#### 植保与能耗
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/pest/observations` | PestObservations | views/pest/observations.vue | 全部 | 病虫害观察 |
| `/energy/records` | EnergyRecords | views/energy/records.vue | 全部 | 能耗记录 |

#### 管理
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/users` | Users | views/users/index.vue | ADMIN | 用户管理 |
| `/audit-logs` | AuditLogs | views/audit-logs/index.vue | ADMIN | 审计日志 |

#### 系统设置
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/settings/notification-channels` | NotificationChannels | views/settings/notification-channels.vue | ADMIN | 通知渠道 |

#### 404
| 路由 | 名称 | 视图文件 | 权限 | 标题 |
|------|------|----------|------|------|
| `/:pathMatch(.*)*` | NotFound | views/NotFound.vue | 公开 | 页面不存在 |

### 兼容重定向（Legacy Redirects，共 10 条）

旧路由 `→` 新路由：
- `/devices` → `/assets/sensor-devices`
- `/devices/:id` → `/assets/sensor-devices`
- `/greenhouses` → `/assets/greenhouses`
- `/device-groups` → `/assets/greenhouses`
- `/telemetry/realtime` → `/collection/realtime`
- `/telemetry/history` → `/collection/history`
- `/controls/commands` → `/strategy/commands`
- `/controls/rules` → `/strategy/policies`
- `/alerts` → `/alerts/list`
- `/alerts/workflow` → `/alerts/timeline`

## 3. 架构

```
src/
├── api/           # 20 个 API 模块（alert, audit, auth, climate, control, crop, dashboard, device, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, request, telemetry, user）
├── assets/        # 全局样式（variables.scss, global.scss）
├── components/    # 6 个组件目录
│   ├── batch/     # BatchForm, StagePlanEditor, BatchReviewBoard
│   ├── charts/    # MetricTrendChart, BatchEventOverlay
│   ├── control/   # PolicyConditionEditor
│   ├── device/    # DeviceStatusBadge
│   ├── layout/    # AppLayout, AppHeader, AppSidebar, index
│   └── telemetry/ # QualityFlagLegend
├── composables/   # 4 个组合式函数（useAuth, usePermission, useAlertSSE, useTelemetrySSE）
├── router/        # 路由配置 + 路由守卫（鉴权 + 角色权限）
├── stores/        # 3 个 Pinia store（auth, greenhouse + index）
├── types/         # 20 个类型定义文件（alert, api, audit, climate, control, crop, dashboard, device, domain, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, telemetry, user）
├── utils/         # 工具函数（constants, format, storage + index）
└── views/         # 17 个视图目录（alerts, audit-logs, batches, climate, common, controls, dashboard, devices, energy, greenhouses, login, nutrient, pest, recipes, settings, telemetry, users）+ NotFound.vue
```

## 4. 已知缺口 / 风险（按优先级）

### P0 - 阻塞级

- 自动化测试完全缺失（无 vitest 或 Cypress 用例）。回归风险高，全靠手动验证。
- 前端未接入 E2E 或组件级测试框架。

### P1 - 高优先级

- SSE 实时数据后端已就绪（`GET /api/alerts/subscribe`、`GET /api/telemetry/subscribe`），前端 composables（`useAlertSSE`、`useTelemetrySSE`）已在 AppHeader 中挂载，实时告警计数可用。telemetry 实时曲线页面仍使用 polling 作为降级方案。
- 批量操作 UI 目前通过 el-dropdown + el-dialog 实现，后续可能需要优化为大面板 + 进度条的体验。
- 移动端适配虽有规划（FRONTEND_PRD.md §4），但当前仅基于 Element Plus 默认响应式，未专门适配。

### P2 - 中优先级

- 前端开发服务器（Vite 默认 5173 / 实际配置 8082）与后端 API（3000）不一致，开发时依赖 Vite proxy 或 CORS。
- 部分页面加载未做骨架屏或 loading 优化。
- FRONTEND_PRD.md 中规划的 DeviceCard、TelemetryCard 等组件尚未抽取为通用组件。

## 5. 后续步骤（按顺序）

1. 引入自动化测试框架（vitest + vue-test-utils），为首批关键页面添加冒烟测试。
2. 完善 SSE 实时数据在仪表盘/设备详情页的集成展示（实时推送 + 动态图表更新），telemetry 实时曲线从 polling 迁移至 SSE。
3. 优化批量操作体验：进度条、分批处理状态、失败重试。
4. 移动端适配实现（侧边栏 → 底部导航，表格 → 卡片列表）。
5. 抽取通用业务组件（DeviceCard, TelemetryCard, BatchOperationDialog）。
6. 修复 recipes/index.vue 和 alerts/index.vue 中的类型错误（RecipeTarget、AlertStats 接口不匹配）。

## 6. 运维命令

```bash
cd packages/frontend

# 安装依赖
npm install

# 启动开发服务器（端口 8082）
npm run dev

# TypeScript 类型检查
npm run type-check

# 生产构建
npm run build

# 预览生产构建
npm run preview

# ESLint 检查
npm run lint
```

## 7. 模块索引（快速导航）

- 入口文件：`src/main.ts`
- 路由配置：`src/router/index.ts`
- API 模块：`src/api/`
- API 统一导出：`src/api/index.ts`
- Axios 实例封装：`src/api/request.ts`
- 类型定义：`src/types/`
- 类型统一导出：`src/types/index.ts`
- 状态管理：`src/stores/`
- 权限控制：`src/composables/usePermission.ts`
- 鉴权逻辑：`src/composables/useAuth.ts`、`src/stores/auth.ts`
- 实时数据：`src/composables/useAlertSSE.ts`、`src/composables/useTelemetrySSE.ts`
- 布局组件：`src/components/layout/`
- 侧边栏配置：`src/components/layout/AppSidebar.vue`
- API 文档：`../../shared/docs/API_SPEC.md`、`../../shared/docs/openapi.yaml`
- 产品需求：`docs/FRONTEND_PRD.md`
- 前端规范：`.claude/CLAUDE.md`

## 8. 更新规则

- 内容变更时始终更新 `最后更新` 日期。
- 保持本文件稳定且处于摘要级别。
- 将短期详情放入 `docs/HANDOFF.md`。
