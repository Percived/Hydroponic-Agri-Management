# 交接文档

最后更新: 2026-05-04
当前分支: main
当前重点: Phase 2 完成 — 设备数据看板、批量操作、通知渠道、系统配置。

## 1. 近期变更（Phase 2）

### 设备模块
- `views/devices/detail.vue` — 重构为双 Tab 结构："基本信息" + "数据看板"
  - 数据看板：时间范围选择器 + 指标多选 + 每指标卡片（avg/max/min/告警数）+ ECharts 微型折线图（告警时间点红色标记）
  - 底部：在线率进度条 + 告警事件列表
  - 图表实例通过 `chartInstances` Record 管理，watch 响应指标选择变化
- `views/devices/list.vue` — 新增批量操作
  - 选择列（`el-table-column type="selection"`）
  - 批量操作下拉菜单：批量启用/禁用、修改分组、修改采样间隔、下发命令、批量删除（含原因）
  - 每项操作弹出对应 dialog，批量命令调用 `POST /api/controls/batch-commands`
- `api/device.ts` — 新增 `getTelemetrySummary()`、`batchUpdateDevices()`、`batchDeleteDevices()`、`batchCommands()` 及对应 TypeScript 接口

### 通知模块（新）
- `views/settings/notification-channels.vue` — 通知渠道 CRUD 页面
  - 表格展示所有渠道，支持创建/编辑/删除
  - 创建/编辑弹窗：渠道类型（EMAIL/SMS/WEBHOOK）、名称、JSON 配置编辑器、最低告警级别
  - WEBHOOK 类型：测试发送按钮 → `POST /api/notification-channels/:id/test`
- `api/notification.ts` — `getChannels()`、`createChannel()`、`updateChannel()`、`deleteChannel()`、`testChannel()`
- `types/notification.ts` — `NotificationChannel`、`CreateChannelRequest`、`UpdateChannelRequest`、`ChannelType` 枚举

### 系统配置模块（新）
- `views/settings/system-config.vue` — 系统配置页面
  - 表格列出配置键值，点击可编辑行内修改
  - 敏感字段（jwt_secret/db_password/mqtt_password）显示为脱敏值，不可点击编辑
  - 保存/取消行内操作
- `api/system-config.ts` — `getSystemConfigs()`、`updateSystemConfig()`、`SystemConfigItem` 等类型

### 布局与路由
- `components/layout/AppSidebar.vue` — 新增 "系统设置" `el-sub-menu`（仅 ADMIN 可见），含"通知渠道"和"系统配置"两个子项
- `router/index.ts` — 新增 `/settings/notification-channels`（ADMIN）和 `/settings/system-config`（ADMIN）两个路由

### API 与类型导出
- `api/index.ts` — 新增 `notificationApi` 和 `systemConfigApi` 导出
- `types/index.ts` — 新增 `'./notification'` 导出

---

## 2. 历史变更

- 新增温室管理页面（`views/greenhouses/index.vue`）：温室 CRUD 表格 + 删除确认
- 新增温室相关 API 模块（`api/greenhouse.ts`）和类型（`types/greenhouse.ts`）
- 新增温室 Pinia store（`stores/greenhouse.ts`）
- 侧边栏新增"温室管理"菜单项（仅 ADMIN 可见）
- 路由新增 `/greenhouses`（ADMIN）

- 初始 MVP 前端交付（Phase 1）：
  - 登录页（`views/login/index.vue`）
  - 设备列表 + 详情（`views/devices/list.vue`、`detail.vue`）
  - 设备分组管理（`views/device-groups/index.vue`）
  - 遥测实时数据（`views/telemetry/realtime.vue`）
  - 遥测历史数据（`views/telemetry/history.vue`）
  - 首页仪表盘（`views/dashboard/index.vue`）
  - 控制命令 + 规则（`views/controls/commands.vue`、`rules.vue`）
  - 告警中心（`views/alerts/index.vue`）
  - 用户管理（`views/users/index.vue`）
  - 审计日志（`views/audit-logs/index.vue`）
  - 路由守卫（鉴权 + RBAC）
  - JWT Token 存储与自动附加
  - 404 页面
  - SSE composables（useAlertSSE, useTelemetrySSE）

- 前端项目初始化：
  - `FRONTEND_PRD.md` — 产品需求文档
  - `.claude/CLAUDE.md` — 前端编码规范
  - `docs/plans/2026-04-20-mvp-frontend-design.md` — MVP 设计决策
  - `docs/plans/2026-04-21-p0-features-design.md` — P0 功能设计

---

## 3. 待办事项（前 5 项）

1. 引入前端自动化测试（vitest + vue-test-utils），覆盖登录、设备列表、路由守卫等关键路径。
2. 将 SSE 实时通道在仪表盘页面做可视化展示（实时推送 + 动态图表）。
3. 批量操作 UI 升级：进度条反馈、分批次状态展示、失败重试入口。
4. 移动端适配：侧边栏 → 底部导航栏、表格 → 卡片列表（参考 FRONTEND_PRD.md §4）。
5. 抽取通用业务组件（DeviceCard、TelemetryCard、StatusBadge、ConfirmDialog）。

## 4. 阻碍 / 风险

- 自动化测试完全空白，所有功能验证依赖手动操作。
- SSE composables（useAlertSSE、useTelemetrySSE）已编写但未在任何视图中实际挂载消费。
- 前端未建立 CI/CD 流程，无 lint-staged 或 pre-commit hook。
- 部分新页面未在移动端尺寸下验证过布局表现。

## 5. 验证说明

- `npm run type-check` — TypeScript 类型检查（当前通过）
- `npm run build` — 生产构建（当前通过）
- 开发时如使用 `npm run dev`（端口 8082），需确认 Vite proxy 配置指向后端 API 端口。

## 6. 下个会话如何继续

1. 阅读 `docs/PROJECT_STATUS.md`。
2. 阅读本文件（`docs/HANDOFF.md`）。
3. 从待办事项 #1 开始，除非优先级发生变化。

## 7. 快速填写模板（每次交接使用）

- 日期：2026-05-04
- 分支：main
- 已完成范围：Phase 2 — 设备数据看板（ECharts + 遥测概览）、批量设备操作、通知渠道 CRUD + Webhook 测试、系统配置管理、侧边栏新增系统设置子菜单
- 待完成范围：自动化测试、SSE 可视化集成、批量操作 UI 优化、移动端适配
- 风险：零自动化测试覆盖；SSE composables 未挂载消费
- 下个首要命令：`npm run type-check` / `npm run build`
