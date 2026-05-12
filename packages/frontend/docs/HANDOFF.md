# 交接文档

最后更新: 2026-05-12
当前分支: version2
当前重点: v0.8.1 — 气候联动触发源单通道化

## 最新变更 (2026-05-12)

### 审计日志页对齐后端最小可用审计返回

- **`src/views/audit-logs/index.vue`**
  - 移除当前后端未提供的 `IP地址` 列，避免固定显示空值
  - 保持现有动作筛选与列表结构不变，直接展示后端返回的 `detail` 文本
- **`src/types/audit.ts`**
  - `username` 调整为可空
  - 删除 `ip_address` 字段依赖，新增可选 `request_id`
- **后端配套变化**
  - `/api/audit-logs` 现在会返回 `username`
  - `detail` 统一为可直接渲染的紧凑 JSON 字符串
  - 审计记录已覆盖登录、用户、设备、策略、告警，以及每次执行器通道命令

### 策略控制页：SCHEDULE 改为结构化计划编辑

- **`src/views/controls/rules.vue`**
  - `SCHEDULE` 不再通过“无条件（仅定时执行）”开关表达定时语义
  - 新增结构化计划编辑：`单次执行 / 每日执行 / 每周执行`
  - 表单新增 `schedule_mode/run_once_at/time_of_day/weekdays_mask/timezone`
  - `effective_from / effective_to` 文案保留为生效窗口，不再表达执行时刻
  - 列表新增“计划描述”列，用于展示单次/每日/每周计划摘要
- **`src/types/policy.ts`**
  - `ControlPolicy`、`CreatePolicyRequest` 补齐 `SCHEDULE` 调度字段
  - `enabled` 放宽为 `number | boolean`，兼容后端 `bool` 返回

### 营养液槽温度绑定语义调整：前端改为只选择水温通道

- **`src/views/nutrient/tanks.vue`**
  - “温度传感器通道”文案改为“水温传感器通道”
  - 绑定下拉项只保留 `metric_code=WATER_TEMP` 的采集通道，避免误绑环境温度通道
  - 展开监测卡片文案同步改为“水温”
- **`src/types/nutrient.ts`**
  - 保留后端既有字段名 `temp_sensor_channel_id`，但前端类型注释明确其语义为“水温通道绑定”
- **`tests/nutrient/tanks-water-temp.test.mjs`**
  - 新增最小回归测试，校验营养液槽页面的绑定文案与 `WATER_TEMP` 筛选条件

## 最新变更 (2026-05-10)

### SSE 契约固化：schema_version 校验 + 新 SSE DTO 类型

- **`src/composables/useTelemetrySSE.ts`**
  - 接收 `telemetry_update` 时校验 `schema_version===1`，不匹配则断开并标记 error
- **`src/composables/useAlertSSE.ts`**
  - 接收 `new_alert` 时校验 `schema_version===1`，不匹配则断开
- **`src/types/telemetry.ts`**
  - `TelemetrySSEEvent` 增加 `schema_version`
- **`src/types/alert.ts`**
  - `Alert` 增加可选 `schema_version/device_code/batch_id` 字段以对齐 SSE data
- **`src/types/device.ts`**
  - 新增 `DeviceStatusSSEDataV1`
- **`src/types/control.ts`**
  - 新增 `CommandDispatchedSSEDataV1`、`CommandAckSSEDataV1`

## 最新变更 (2026-05-09)

### 气候联动触发源改造：固定单一采集通道

- **`src/views/climate/index.vue`**
  - Profile 表单新增：温室→采集设备→指标→采集通道四级选择
  - 创建/更新 Profile 提交 `trigger_sensor_channel_id`
  - 列表与执行日志弹窗补充展示触发通道/指标/采集时间字段
  - 删除按钮仅 ADMIN 展示（后端限制 DELETE 为 ADMIN）
- **`src/types/climate.ts`**
  - `CreateClimateProfileRequest` / `CreateClimateProfileWithStagesRequest` 增加必填 `trigger_sensor_channel_id`
  - `ClimateExecutionLog` 增加可选字段：`trigger_sensor_channel_id/trigger_metric_code/collected_at`
- **`src/views/devices/detail.vue`**
  - 采集通道删除入口移除（后端禁止 DELETE）；通过启用开关进行停用

### 营养液管理可用性优化

- **`src/views/nutrient/tanks.vue`**
  - 新增/编辑液槽：传感器通道下拉项由 `channel_code (ID:channel_id)` 改为 `channel_code (设备名/设备编码)`，减少用户记忆成本

### 列表外键字段可读化

- **`src/utils/labels.ts`**
  - 新增 `*_id → 名称` 的通用 label 工具（温室/种植区/批次/作物/阶段/设备/通道等）
- **`src/views/greenhouses/zones.vue`**
  - 表格列：`温室ID` → `温室`（显示温室名称）
- **`src/views/nutrient/tanks.vue`**
  - 表格列：`种植区ID` → `种植区`（显示种植区名称）
- **`src/views/batches/ledger.vue`**
  - 表格列：温室/作物品种/种植区外键改为名称展示
- **`src/views/pest/observations.vue`**
  - 表格列：温室/种植区/批次外键改为名称展示
- **`src/views/energy/records.vue`**
  - 表格列：温室/批次外键改为名称展示
- **`src/views/nutrient/ion-tests.vue`**
  - 表格列：液槽外键改为液槽编号展示
- **`src/views/recipes/index.vue`**
  - 阶段指标/离子目标：生长阶段由 ID 改为名称展示；新增/编辑改为下拉选择生长阶段
- **`src/views/batches/stage-plans.vue`**
  - 表格列：生长阶段由 ID 改为名称展示（冲突提示同步名称化）
- **`src/views/batches/harvest.vue`**
  - 表格列：批次由 ID 改为名称展示
- **`src/views/climate/index.vue`**
  - Profiles/Logs：温室与触发通道由 ID 改为名称展示；Actions：执行器通道展示 `channel_code (所属设备)`
- **`src/views/alerts/index.vue`**
  - 通道列由 `#id` 改为 `channel_code (所属设备)`（按需缓存查询）
- **`src/views/audit-logs/index.vue`**
  - `目标ID` 列更名为 `目标`（展示 `#id`，避免“XXID”表头）

### 资产中心权限与启停

- **`src/views/devices/list.vue`**
  - 资产中心设备列表：新增/删除/提交按钮按角色控制（VIEWER 只读；ADMIN/OPERATOR 可新增；删除仅 ADMIN）
- **`src/views/devices/detail.vue`**
  - 设备详情：编辑设备/通道增改按角色控制（ADMIN/OPERATOR）
  - 通道启停：`enabled` 支持开关切换（调用 `PUT /api/sensor-channels/:id`、`PUT /api/actuator-channels/:id`）
  - 修复编辑设备/编辑通道弹窗打开后被 `resetFields()` 清空的问题
- **`src/views/greenhouses/index.vue`**
  - 温室启停：`status` 支持开关切换（ENABLED/DISABLED，调用 `PUT /api/greenhouses/:id`）
- **`src/views/greenhouses/zones.vue`**
  - 种植区启停：`status` 支持开关切换（ENABLED/DISABLED，调用 `PUT /api/growing-zones/:id`）
  - 删除取消不再抛异常；编辑/新增按角色控制（ADMIN/OPERATOR），删除仅 ADMIN
- **`src/views/alerts/index.vue`**
  - 补齐 `AlertStats.ignored_count` 默认值，保证类型检查通过
- **构建说明**
  - `npm run type-check` / `npm run build` 在部分环境下可能出现 node 内存不足，需要设置 `NODE_OPTIONS=--max-old-space-size=4096`

### 采集中心可用性提升：SSE 生命周期 + 过滤生效 + 错误可见

- **`src/composables/useTelemetrySSE.ts`**
  - 新增连接状态：`status`（disconnected/connecting/connected/error）与 `lastError`
  - 订阅 URL 支持过滤参数：`device_codes`、`metric_codes`
- **`src/views/telemetry/overview.vue`**
  - 页面挂载自动连接 SSE，离开页面断开
  - 顶部显示连接状态，连接失败可手动重连
  - 级联加载/通道加载增加竞态保护（避免快速切换时旧请求覆盖新状态）
  - 加载失败显示错误提示与“重试”
- **`src/views/telemetry/trends.vue`**
  - 级联加载增加竞态保护
  - 加载失败显示错误提示与“重试”
- **`src/types/telemetry.ts`**
  - `quality_flag` 枚举对齐后端：`normal/missing/out_of_range/device_offline`

### 采集中心趋势图修复

- **`src/views/telemetry/overview.vue`**
  - 修复 SSE 趋势缓冲区重复追加问题：只有通道最新事件真正变化时才写入趋势数组
  - 趋势点追加改为幂等处理，避免 EventSource 重连或其他通道更新时把旧点重复写回同一条曲线
  - 历史趋势回填改为统一排序和去重，保证图表输入始终按采集时间升序
- **`src/views/telemetry/trendBuffer.ts`** (新增):
  - 抽离 `appendTrendPoint` / `normalizeTrendPoints`，统一处理趋势点去重、排序、缓冲区裁剪
- **`tests/telemetry/trend-buffer.test.ts`** (新增):
  - 回归覆盖重复 SSE 点不重复追加
  - 回归覆盖迟到点插入后仍保持时间升序
  - 回归覆盖历史数据归一化时去重与排序

### 告警SSE订阅对齐：过滤参数生效 + 类型统一

- **`src/composables/useAlertSSE.ts`**
  - 订阅参数使用 `level` + `device_codes`（可选），对齐后端 SSE 过滤
  - `lastAlert` 类型改为复用 `src/types/alert.ts` 的 `Alert`（移除重复的 AlertEvent 定义）
- **`src/composables/index.ts`**
  - 移除 `AlertEvent` 类型导出（改为导出 `UseAlertSSEOptions/UseAlertSSEReturn`）

### 通知渠道类型对齐：补齐 IN_APP 枚举

- **`src/types/notification.ts`**
  - `ChannelType` 与后端枚举对齐，新增 `IN_APP`；展示名补齐 `站内通知`

## 最新变更 (2026-05-08)

### 采集中心模块改进：2页替代3页 + SSE 实时总览 + 多通道趋势分析

- **路由变更**: 3条采集中心路由合并为2条:
  - `/collection/overview` → `views/telemetry/overview.vue` (实时总览，新)
  - `/collection/trends` → `views/telemetry/trends.vue` (趋势分析，新)
  - 旧路径 `/collection/realtime`, `/collection/history`, `/collection/batch-trends` 及其他 legacy 路径重定向到新路由
- **`src/views/telemetry/overview.vue`** (新建):
  - 温室→种植区→设备(多选)级联过滤 + 指标多选
  - 传感器卡片网格：设备名/通道码/指标名/数值+单位/质量标识/在线状态/SSE更新高亮
  - 可折叠趋势图区域 (MetricTrendChart 复用)
  - 加载/空状态覆盖：骨架卡片、el-empty 多态提示
- **`src/views/telemetry/trends.vue`** (新建):
  - 温室→种植区→通道(多选搜索)级联 + 指标/时间范围/批次/质量标识过滤
  - 统计摘要(avg/max/min)、多指标对比图表 + 批次事件时间线、数据明细表格(分页)
  - 指标名称动态缓存 populating (metricApi + populateMetricNames)
- **`src/types/telemetry.ts`** — 新增 `TelemetryLatestItem`, `TelemetryLatestBatchResponse`, `ChannelSnapshot`, `TelemetrySSEEvent` 类型
- **`src/api/telemetry.ts`** — 新增 `getChannelsLatest(channelIds: number[])` 批量查询函数
- **`src/composables/useTelemetrySSE.ts`** — 新增 `channelValues: Map<number, TelemetrySSEEvent>` 按 sensor_channel_id 索引 SSE 事件；兼容后端实际数据格式 `{sensor_channel_id, metric_code, value, collected_at, device_code}`
- **`src/components/layout/AppSidebar.vue`** — 采集中心子菜单从3项改为2项：实时总览、趋势分析
- **删除文件**: `src/views/telemetry/realtime.vue`, `history.vue`, `batch-trends.vue`

### 气候模块前端修复

- `src/views/climate/index.vue`:
  - 表格: `stage_count` → `stages_count`（对齐后端 `stages_count` 字段）
  - 阶段操作符选择: 移除 `==` 选项（后端验证不支持）
  - Profile 表单: 新增 `enabled` 开关
  - Action 表单: 新增 `enabled` 开关
  - Profile 编辑提交: 不再发送 `greenhouse_id`、`code`（`UpdateClimateProfileRequest` 无此字段）
  - `execution_order` 默认值 `0` → `1`（后端 `min=1`）
  - Action 编辑: `command_payload_str` 直接使用后端返回的字符串
- `src/types/climate.ts`:
  - `ClimateProfile.enabled`: `number` → `boolean`
  - `ClimateStageAction.enabled`: `number` → `boolean`
  - `ClimateStageAction.command_payload`: `Record<string, unknown>` → `string`
  - `CreateClimateProfileRequest.enabled` 新增 `?: boolean`
  - `CreateClimateStageActionRequest.enabled` 新增 `?: boolean`

## 0. 近期变更（v0.8.0）

### 后端 Phase 3+4 同步

- **SSE 实时推送就绪**：后端新增 `GET /api/alerts/subscribe`、`GET /api/telemetry/subscribe`，前端 composables 已在 AppHeader 挂载，实时告警计数正常。
- **仪表盘类型对齐**：`DashboardOverview` 新增 `devices_online/offline/total`、`device_type_distribution` 字段，温室设备数修正为 `sensor_count + actuator_count`。
- **后端 handler 拆分**：climate、policy、nutrient、crop 单文件拆为 3-5 个子文件，前端无需变更。
- **后端 enabled 类型统一**：`uint8` → `bool`，前端 DTO 类型自动兼容。

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
2. **SSE 深度集成**：telemetry 实时曲线从 polling 迁移至 SSE（当前 polling 作为降级方案），Dashboard 页面接入 telemetry SSE 实时更新图表。
3. **批量操作 UI 增强**：进度条反馈、分批状态展示、失败重试入口。
4. **移动端适配**：侧边栏切换为底部导航栏、表格切换为卡片列表。
5. **通用组件抽取**：提取 DeviceCard、TelemetryCard、StatusBadge、ConfirmDialog 等为共享组件。
6. **类型修复**：recipes/index.vue（RecipeTarget/RecipeTargetsResponse 接口不匹配）、alerts/index.vue（AlertStats.ignored_count 缺失）。

---

## 4. 风险 / 阻碍

- **零自动化测试覆盖**：所有功能验证依赖手动操作，回归风险高。
- **SSE 可用性**：后端 SSE 端点已就绪，前端 AppHeader 已挂载告警 SSE；telemetry 实时曲线仍用 polling，待迁移。
- **无 CI/CD 流程**：前端未接入任何持续集成/持续部署管线。
- **移动端未适配**：仅依赖 Element Plus 默认响应式，未做专门适配。
- **类型瑕疵**：recipes 与 alerts 页面存在预存类型错误（vue-tsc 报 8 个 error）。

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

- **日期**：2026-05-07
- **分支**：main
- **已完成范围**：v0.7.0 全量架构重构 + v0.8.0 SSE 实时推送就绪 + 仪表盘类型对齐后端 v2.3.0。AppHeader 挂载告警 SSE（实时计数、浏览器通知），DashboardOverview 新增 devices_* 聚合字段与 device_type_distribution。
- **待完成范围**：自动化测试、SSE 深度集成（telemetry polling → SSE）、批量 UI 优化、移动端适配、通用组件抽取、类型修复（recipes/alerts）
- **风险**：零自动化测试覆盖；telemetry SSE 未挂载消费（polling 为降级）；无 CI/CD
- **下个首要命令**：`npm run type-check && npm run build`
