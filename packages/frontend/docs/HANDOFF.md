# 交接文档

最后更新: 2026-05-04
当前分支: main
当前重点: API 响应字段补齐 — 控制规则/告警列表设备名称与动作字段修复。

## 1. 近期变更（2026-05-04 会话）

### 控制规则页 & 告警页 — 数据映射修复
- **`api/control.ts`** — `getRules` 函数重写映射逻辑：
  - 从后端返回的 `action` JSON 中提取 `command_type` + `command_payload`
  - 映射后端新增的 `target_device_name` 字段
  - 兼容后端返回字段的大小写变体（PascalCase / snake_case）
- **`types/control.ts`** — `ControlRule` 新增 `action` 可选字段
- **`types/alert.ts`** — `Alert` 新增 `triggered_at`、`resolved_at` 字段

### 构建验证
- `npm run type-check` 通过
- 修复了 `CommandType` 枚举导入缺失问题

---

## 2. 历史变更

### Phase 2
- `views/devices/detail.vue` — 双 Tab 数据看板（ECharts + 遥测概览）
- `views/devices/list.vue` — 批量操作（启用/禁用、分组、采样间隔、命令、删除）
- `views/settings/notification-channels.vue` — 通知渠道 CRUD + Webhook 测试
- `views/settings/system-config.vue` — 系统配置管理（含脱敏）
- `components/layout/AppSidebar.vue` — 系统设置子菜单（ADMIN）
- `api/notification.ts`、`api/system-config.ts`、`api/device.ts` 新增接口

### 更早变更（Phase 1）
- MVP 前端交付：登录、设备列表/详情、设备分组、遥测、仪表盘、控制命令/规则、告警、用户管理、审计日志
- 路由守卫、JWT 鉴权、404 页面
- SSE composables（useAlertSSE, useTelemetrySSE）

---

## 3. 待办事项（前 5 项）

1. 引入前端自动化测试（vitest + vue-test-utils），覆盖登录、设备列表、路由守卫等关键路径。
2. 将 SSE 实时通道在仪表盘页面做可视化展示（实时推送 + 动态图表）。
3. 批量操作 UI 升级：进度条反馈、分批次状态展示、失败重试入口。
4. 移动端适配：侧边栏 → 底部导航栏、表格 → 卡片列表。
5. 抽取通用业务组件（DeviceCard、TelemetryCard、StatusBadge、ConfirmDialog）。

## 4. 阻碍 / 风险

- 自动化测试完全空白，所有功能验证依赖手动操作。
- SSE composables 已编写但未在任何视图中实际挂载消费。
- 前端未建立 CI/CD 流程。

## 5. 验证说明

- `npm run type-check` — TypeScript 类型检查（当前通过）
- `npm run build` — 生产构建（当前通过）

## 6. 下个会话如何继续

1. 阅读 `docs/PROJECT_STATUS.md`。
2. 阅读本文件（`docs/HANDOFF.md`）。
3. 从待办事项 #1 开始。

## 7. 快速填写模板（每次交接使用）

- 日期：2026-05-04
- 分支：main
- 已完成范围：控制规则 `getRules` 映射逻辑重写（action → command_type/command_payload）、Alert/ControlRule 类型字段补齐
- 待完成范围：自动化测试、SSE 集成、批量 UI 优化、移动端适配
- 风险：零自动化测试覆盖；SSE composables 未挂载
- 下个首要命令：`npm run type-check` / `npm run build`
