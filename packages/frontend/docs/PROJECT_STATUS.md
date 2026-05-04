# 项目状态

最后更新: 2026-05-04
负责人: 前端团队
版本: v0.2（Phase 2 完成）

## 1. 项目概述

基于 Vue 3 + TypeScript + Element Plus 构建的水培农业管理系统 Web 前端。
核心依赖：Vite、Pinia、Vue Router、Axios、ECharts。

## 2. 当前交付状态

总体评估：Phase 2 功能已全部落地，构建通过，可进行手动集成测试。
TypeScript 类型检查通过（vue-tsc --noEmit），生产构建通过（npm run build）。

### 已实现页面（16 个路由）

| 页面 | 路由 | 权限 | Phase |
|------|------|------|-------|
| 登录页 | `/login` | 公开 | P1 |
| 首页仪表盘 | `/` | 全部 | P1 |
| 设备列表 | `/devices` | 全部 | P1 |
| 设备详情 | `/devices/:id` | 全部 | P1 |
| 温室管理 | `/greenhouses` | ADMIN | P1 |
| 设备分组 | `/device-groups` | 全部 | P1 |
| 遥测实时数据 | `/telemetry/realtime` | 全部 | P1 |
| 遥测历史数据 | `/telemetry/history` | 全部 | P1 |
| 控制命令 | `/controls/commands` | ADMIN/OPERATOR | P1 |
| 控制规则 | `/controls/rules` | ADMIN/OPERATOR | P1 |
| 告警中心 | `/alerts` | 全部 | P1 |
| 用户管理 | `/users` | ADMIN | P1 |
| 审计日志 | `/audit-logs` | ADMIN | P1 |
| 通知渠道 | `/settings/notification-channels` | ADMIN | P2 |
| 系统配置 | `/settings/system-config` | ADMIN | P2 |
| 404 | `/:pathMatch(.*)*` | 公开 | P1 |

### Phase 2 新增功能

- 设备详情页增设"数据看板"Tab：时间范围选择 + 指标多选 + ECharts 微型折线图 + 在线率 + 告警事件列表
- 设备列表页批量操作：多选设备后支持批量启用/禁用、批量修改分组、批量修改采样间隔、批量下发命令、批量删除（含原因记录）
- 通知渠道管理页：CRUD 表格 + 创建/编辑弹窗（支持 EMAIL/SMS/WEBHOOK 三种类型）+ Webhook 测试发送
- 系统配置页：配置键值表格 + 行内编辑 + 敏感字段（jwt_secret/db_password/mqtt_password）脱敏显示
- 侧边栏新增"系统设置"子菜单（仅 ADMIN 可见），包含通知渠道和系统配置入口

### 架构分层

```
src/
├── api/           # 14 个 API 模块（auth, device, device-group, greenhouse, telemetry, control, alert, audit, user, dashboard, notification, system-config, request, index）
├── assets/        # 全局样式（variables.scss, global.scss）
├── components/    # 公共组件（layout: AppLayout, AppHeader, AppSidebar）
├── composables/   # 4 个组合式函数（useAuth, usePermission, useAlertSSE, useTelemetrySSE）
├── router/        # 路由配置 + 路由守卫（鉴权 + 角色权限）
├── stores/        # 3 个 Pinia store（auth, device, greenhouse）
├── types/         # 10 个类型定义文件（api, user, device, greenhouse, telemetry, control, alert, audit, dashboard, notification）
├── utils/         # 工具函数（storage, format, constants, request）
└── views/         # 14 个页面组件（login, dashboard, devices/*, greenhouses, device-groups, telemetry/*, controls/*, alerts, users, audit-logs, settings/*, NotFound）
```

## 3. 已知缺口 / 风险（按优先级）

P0：

- 自动化测试完全缺失（无 vitest 或 Cypress 用例）。回归风险高，全靠手动验证。
- 前端未接入 E2E 或组件级测试框架。

P1：

- SSE 实时数据（Telemetry/Alert）在 composables 中有实现，但未在所有页面中充分集成展示。
- 批量操作 UI 目前通过 el-dropdown + el-dialog 实现，后续可能需要优化为大面板 + 进度条的体验。
- 移动端适配虽有规划（FRONTEND_PRD.md §4），但当前仅基于 Element Plus 默认响应式，未专门适配。

P2：

- 前端开发服务器端口（8082）与后端 API（3000）不一致，开发时依赖 Vite proxy 或 CORS。
- 部分页面加载未做骨架屏或 loading 优化。
- FRONTEND_PRD.md 中规划的 DeviceCard、TelemetryCard 等组件尚未抽取为通用组件。

## 4. 后续步骤（按顺序）

1. 引入自动化测试框架（vitest + vue-test-utils），为首批关键页面添加冒烟测试。
2. 完善 SSE 实时数据在仪表盘/设备详情页的集成展示（实时推送 + 动态图表更新）。
3. 优化批量操作体验：进度条、分批处理状态、失败重试。
4. 移动端适配实现（侧边栏 → 底部导航，表格 → 卡片列表）。
5. 抽取通用业务组件（DeviceCard, TelemetryCard, BatchOperationDialog）。

## 5. 运维命令

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

## 6. 模块索引（快速导航）

- 入口：`src/main.ts`
- 路由配置：`src/router/index.ts`
- API 模块：`src/api/`
- 类型定义：`src/types/`
- 状态管理：`src/stores/`
- 权限控制：`src/composables/usePermission.ts`
- 鉴权逻辑：`src/composables/useAuth.ts`、`src/stores/auth.ts`
- 布局组件：`src/components/layout/`
- API 文档：`../../shared/docs/API_SPEC.md`、`../../shared/docs/openapi.yaml`
- 产品需求：`docs/FRONTEND_PRD.md`
- 前端规范：`.claude/CLAUDE.md`

## 7. 新会话上下文包

在开启新模型/会话时，首先分享以下文件：

1. `docs/PROJECT_STATUS.md`
2. `docs/HANDOFF.md`
3. 你的直接目标（一句话）

## 8. 更新规则

- 内容变更时始终更新 `最后更新` 日期。
- 保持本文件稳定且处于摘要级别。
- 将短期详情放入 `docs/HANDOFF.md`。
