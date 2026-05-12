# 交接文档

最后更新: 2026-05-12
当前分支: version2
当前重点: v2.3.2 气候联动触发源单通道化

## 最新变更 (2026-05-12)

### 营养液槽传感器绑定语义说明：`temp_sensor_channel_id` 约定绑定水温通道

- **契约说明**
  - 后端字段名维持 `temp_sensor_channel_id` 不变，避免迁移与 DTO 破坏式变更
  - 前端与接口文档统一将该字段解释为“水温传感器通道 ID”，实际应绑定 `metric_code=WATER_TEMP` 的采集通道
  - 当前后端 handler 仍未对 `metric_code` 做强校验，如需在服务端兜底，后续可在 nutrient tank create/update 增加校验

## 最新变更 (2026-05-10)

### 执行器命令单通道化：设备级 Topic 保持兼容，Payload 补目标通道

- **命令下发 payload 补齐目标通道元数据**
  - `internal/command/` 手动命令下发在保留设备级 MQTT topic 的前提下，向 payload 注入 `actuator_channel_id` 与 `channel_code`
  - `internal/policy/scheduler.go`、`internal/climate/profile_scheduler.go` 自动下发链路同步注入同样字段，避免手动/自动行为不一致
- **模拟器执行器改为优先单通道执行**
  - `cmd/simulator/actuator.go` 收到带目标通道的命令后，仅对匹配的执行器通道生效
  - 未携带目标通道时仍保持原有“整机广播”兼容行为
  - 若显式指定的通道不存在，模拟器不再误控全部通道，并返回 `invalid` ACK
- **测试**
  - 新增后端测试覆盖：验证 MQTT 下发 payload 包含目标执行器通道标识
  - 新增模拟器测试覆盖：验证命令仅作用于目标通道

### SSE 契约固化：DTO v1 + Devices/Commands 订阅端点

- **DTO v1（schema_version=1）**
  - `telemetry_update`：SSE data 增加 `schema_version`
  - `device_status`：新增 DTO v1（含 `reported_at`、可选 `reason`）
  - `command_dispatched` / `command_acked`：新增 DTO v1（含 `source_type/source_id`、`acked_at`）
- **新 SSE 端点**
  - `GET /api/devices/subscribe`（支持 `device_codes` 过滤）
  - `GET /api/commands/subscribe`（支持 `device_codes` 过滤；一条 SSE 输出 `command_dispatched` + `command_acked`）
- **Producer 统一发布**
  - MQTT ingress、offline detector、command handler、policy scheduler、climate scheduler 统一发布 DTO v1（避免 map 字段漂移）
  - policy/climate 的 MQTT publish 失败也会发布 `FAILED` 的 `command_dispatched` 事件用于可观测性
- **SSE handler 增强**
  - 支持 struct/map 的字段抽取与过滤；支持 multi-type subscribe

### 配置可靠投递：config_deliveries + ack_type=config + 重试

- **Schema**
  - 新增 `config_deliveries`：记录 `msg_id/trace_id/entity_rev/request_payload/status/retry/ack` 等投递权威字段
- **ConfigPusher**
  - `hydroponic/{deviceCode}/cmd/config` 推送改为 v1 payload（含 `schema_version/msg_id/trace_id/entity_rev/ttl_sec/require_ack/payload`）
  - MQTT 不连/发布失败不再静默跳过：会落库为 FAILED 并进入重试
- **Ingress ACK**
  - `/ack` 支持解析 `schema_version` + `ack_type=config`，按 `msg_id` 回写 `config_deliveries` 状态
  - legacy command ack 行为保持兼容
- **Retry Worker**
  - 周期扫描：SENT 超时转 FAILED(ACK_TIMEOUT)；FAILED 到点重发并递增 retry_count
- **Observability**
  - 新增查询 API：`GET /api/config-deliveries`、`GET /api/config-deliveries/:id`

### 批次阶段自动联动：crop_batch_stage 配置下发 + 定时调度

- **Schema**
  - `batch_stage_plans` 新增：`recipe_id` / `policy_id` / `climate_profile_id`（阶段级权威配置引用）
  - `crop_batches` 新增：`active_climate_profile_id`（与 existing active_recipe_id/active_policy_id 一致）
  - 新增 `batch_stage_runtime`：记录每个批次上次已应用的 `current_stage_plan_id`，用于去重与审计
- **BatchStageScheduler**
  - 周期扫描 RUNNING 批次，按 `stage_start_at <= now < stage_end_at` 判定当前阶段
  - 当检测到阶段切换时：更新 runtime + 更新 crop_batches.active_* + 对 batch 绑定的 actuator 设备推送 `config_type=crop_batch_stage`
  - 若 stage_plan 绑定了 `climate_profile_id`，会额外推送一次 `config_type=climate_profile`（profile 快照）

## 最新变更 (2026-05-09)

### 气候联动触发源改造：固定单一采集通道

- **Schema**
  - `climate_profiles` 新增 `trigger_sensor_channel_id`（固定单通道触发）
  - `climate_execution_logs` 新增 `trigger_sensor_channel_id` / `trigger_metric_code` / `collected_at`（可观测性增强）
- **`internal/climate/profile_scheduler.go`**
  - 自动调度从“按 trigger_metric_code 匹配”改为“按 telemetry:received 的 sensor_channel_id 精确匹配 profile”
  - 写入执行日志时补齐触发来源字段（通道/指标/采集时间）
  - 执行器动作按 `execution_order` 排序执行；MQTT 不可用或 publish 失败时，命令记录标记为 `FAILED`（不再写成 `SENT`）
- **`internal/climate/profile_handler.go`**
  - Profile 创建/更新增加触发源一致性校验：温室归属一致 + 指标与通道 metric_code 一致 + 通道必须 enabled
- **`internal/device/handler.go`**
  - 采集通道禁用（enabled=false）时，自动将引用该通道的气候 Profile 置为 enabled=false
  - 禁止删除采集通道（DELETE 直接返回冲突提示，要求改用禁用）
- **文档**
  - `shared/docs/API_SPEC.md`、`shared/docs/openapi.yaml` 对齐新增字段与删除策略

### 采集中心可用性提升：SSE 过滤 + 稳定性增强

- **`internal/platform/event/sse_handler.go`**
  - `/api/telemetry/subscribe` 支持 query 过滤：`device_codes`、`metric_codes`
  - SSE 增强：连接下发 `retry`；每条消息增加 `id`（优先使用 `collected_at`）
- **`internal/platform/http/middleware.go`**
  - `RequestLogger` 对 `/api/*/subscribe` 跳过响应体捕获，避免 SSE 长连接导致内存持续增长
  - 日志 query 对 `token` 参数脱敏
- **`internal/platform/mqtt/ingress.go`**
  - `telemetry:received` SSE payload 增加 `quality_flag`
- **文档**
  - `shared/docs/API_SPEC.md`、`shared/docs/openapi.yaml` 补齐 `/api/telemetry/subscribe` 参数说明

### 告警闭环一致性：SSE 过滤生效 + Payload 统一 + 自动恢复补时间线

- **`internal/platform/event/sse_handler.go`**
  - `/api/alerts/subscribe` 支持 query 过滤：`level`、`device_codes`（或 `device_code`）
- **告警事件发布方统一 payload**
  - `internal/alert/handler.go`、`internal/platform/mqtt/ingress.go`、`internal/device/offline_detector.go` 发布 `alert:created` 时统一使用 `schema_version=1` 且字段为 `id/type/level/status/triggered_at/...`（兼容前端消费）
  - 新增复用方法：`internal/alert/sse_data.go`（BuildAlertSSEDataV1）
- **自动恢复补审计**
  - 心跳触发自动解决离线告警时，新增写入 RESOLVED 时间线事件（EventPayload: reason=heartbeat）
- **测试**
  - 新增后端单测覆盖：SSE 告警 level 过滤；心跳自动恢复写时间线

### 命令闭环可靠性：ACK 强类型 + MQTT 失败语义收敛为 FAILED

- **`internal/platform/event/types.go`**
  - 新增 `CommandAckData`，用于 `command:acked` 内部事件的强类型承载
- **`internal/platform/mqtt/ingress.go`**
  - `handleAck` 发布 `command:acked` 时使用 `CommandAckData`（避免 waiter 依赖 float64 类型）
- **`internal/command/waiter.go`**
  - 支持消费强类型 `CommandAckData`，并对 legacy map 形式做兼容
- **`internal/command/handler.go`**
  - `SendCommand/DispatchAndWait/DispatchAsync` 发生 MQTT 下发失败时，命令状态写为 `FAILED` 并写入一条 `FAILED` 回执（避免“PENDING 假滞留”）

### 通知安全：TestChannel 资源级权限校验

- **`internal/notification/handler.go`**
  - `TestChannel` 按 `id + user_id` 读取渠道，避免越权测试他人渠道

### 设备状态数据质量：status 值强校验

- **`internal/platform/mqtt/ingress.go`**
  - `handleStatus` 对 status 值做强校验（仅允许 ONLINE/OFFLINE/FAULT），非法值直接丢弃并告警日志

## 最新变更 (2026-05-08)

### 采集中心模块改进：遥测批量查询 + 多通道支持

- **`internal/telemetry/cache.go`** — 新增 `GetMulti(channelIDs []uint64) map[uint64]*CachedRecord` 批量读取方法
- **`internal/telemetry/handler.go`**:
  - 新增 `GetLatestBatch` — `GET /telemetry/channels/latest?ids=1,2,3` 批量获取通道最新值（缓存 + DB 回退，上限 100 个 ID）
  - `QueryTelemetry` — `sensor_channel_id` 参数支持逗号分隔多 ID（`WHERE sensor_channel_id IN (...)`，上限 50 个）
- **`internal/telemetry/dto.go`** — 新增 `LatestRecordResponse`、`LatestBatchResponse` 响应类型
- **`internal/telemetry/routes.go`** — 注册 `channels.GET("/latest", ...)` 路由（在 `:channelId/latest` 之前避免冲突）

### 气候模块 Profile 自动调度器（新增）

- **`internal/climate/profile_scheduler.go`** — 新增 `ProfileScheduler`，遵循 `policy/scheduler.go` 相同的事件驱动模式
  - 订阅 `telemetry:received` 事件
  - 根据 `trigger_metric_code` 匹配启用的气候配置
  - 按 `stage_level ASC` 评估阶段，找到最高匹配的阶段（带滞后回差）
  - 通过 MQTT 下发匹配阶段的执行器动作
  - 记录执行日志到 `climate_execution_logs`
  - 60 秒冷却期防止频繁切换
- **`internal/climate/handler.go`** — Handler 结构体新增 `hub *event.Hub` 字段
- **`internal/climate/routes.go`** — `RegisterRoutes` 中启动 `ProfileScheduler`

### 前端气候模块修复

- 表格列 `stage_count` → `stages_count`（对齐后端字段名）
- 移除 `==` 操作符选项（后端验证不支持）
- Profile 和 Action 表单新增 `enabled` 开关
- `ClimateProfile.enabled` 类型 `number` → `boolean`
- `ClimateStageAction.enabled` 类型 `number` → `boolean`
- `ClimateStageAction.command_payload` 类型 `Record<string, unknown>` → `string`
- Profile 更新 payload 移除 `greenhouse_id` 和 `code`（仅创建时需要）
- `execution_order` 默认值 `0` → `1`（后端验证 `min=1`）

## 0. 近期变更（v2.0.0 重构）

v2.0.0 是项目架构的根本性重构，涉及以下重大变化：

### 设备模块拆分（device -> sensor/actuator 独立实体）

旧架构使用单一 `devices` 表承载所有设备，v2.0.0 拆分为四张独立表：
- `sensor_devices` + `sensor_channels` — 传感器设备与采集通道（温度、湿度、pH、EC 等）
- `actuator_devices` + `actuator_channels` — 执行器设备与控制通道（水泵、风机、LED、阀门等）

对应 API 路径：
- `/api/sensor-devices` — 传感器设备 CRUD（5 个端点）
- `/api/sensor-channels` — 传感器通道 CRUD（5 个端点）
- `/api/actuator-devices` — 执行器设备 CRUD（5 个端点）
- `/api/actuator-channels` — 执行器通道 CRUD（5 个端点）

### 控制模块拆分（control -> command + policy）

旧 `internal/control` 模块承载了命令下发与策略规则双重职责，v2.0.0 拆分为两个独立模块：
- **command** (`internal/command/`) — 命令下发与回执追踪：`/api/commands`
  - 模型：`ControlCommand`、`CommandReceipt`
  - 命令状态机：`PENDING / QUEUED / SENT / ACKED / TIMEOUT / FAILED`
  - 子资源：`/api/commands/{id}/receipts`（回执查询与创建）
  - 动作：`send`、`ack`
- **policy** (`internal/policy/`) — 控制策略引擎：`/api/policies`、`/api/policy-executions`
  - 模型：`ControlPolicy`、`PolicyCondition`、`PolicyTarget`、`PolicyExecution`
  - 策略类型：`THRESHOLD / SCHEDULE / DURATION`
  - 动作：`publish`、`archive`、`execute`
  - 子资源：`/api/policies/{id}/conditions`、`/api/policies/{id}/targets`

### 10+ 新领域模块

v2.0.0 新增了一批完整的业务领域模块，覆盖水培农业全流程：

| 模块 | 目录 | API 前缀 | 说明 |
|------|------|---------|------|
| climate | `internal/climate/` | `/api/climate-profiles`, `/api/climate-execution-logs` | 气候环境配置（多阶段控制） |
| crop | `internal/crop/` | `/api/crop-varieties`, `/api/growth-stages`, `/api/batches`, `/api/batch-stage-plans`, `/api/harvests` | 作物品种、生长阶段、种植批次 |
| energy | `internal/energy/` | `/api/energy-records` | 能耗记录与汇总 |
| greenhouse | `internal/greenhouse/` | `/api/greenhouses`, `/api/growing-zones` | 温室与种植区管理 |
| metric | `internal/metric/` | `/api/metrics` | 指标定义字典 |
| nutrient | `internal/nutrient/` | `/api/nutrient-tanks`, `/api/solution-changes`, `/api/ion-tests`, `/api/concentrate-inventory`, `/api/concentrate-usage-logs` | 营养液管理（DWC 核心） |
| pest | `internal/pest/` | `/api/pest-observations`, `/api/treatment-records` | 病虫害观察与治理 |
| recipe | `internal/recipe/` | `/api/recipes`, `/api/recipe-bindings` | 营养液配方与批次绑定 |
| review | `internal/review/` | `/api/reviews` | 批次复盘快照 |

### 迁移归并

- 主迁移文件：`migrations/merged/all.up.sql`（整合全部 schema 初始化与种子数据，可一次性离线执行）
- 此合并文件为**唯一权威数据源**，不存在独立编号迁移文件；新开发者直接执行此文件即可初始化全量数据库
- 含演示种子数据（覆盖园区/温室/三层设备、测点与采样、批次与阶段、配方与目标、策略与回执、告警处置闭环、复盘快照）

### v2.3.0 阶段三四变更（2026-05-07）

**SSE 实时推送修通**：
- 新增 SSE HTTP 端点：`GET /api/alerts/subscribe`、`GET /api/telemetry/subscribe`（支持 query-string token 鉴权）
- SSE handler 位于 `internal/platform/event/sse_handler.go`，自动映射内部事件类型到前端兼容格式
- 前端 `useAlertSSE.ts`、`useTelemetrySSE.ts` composables 已就绪

**序列化风格统一**：
- `overview` 模块改用强类型 `DashboardResponse` DTO 替代 `gin.H{}`
- 新增 `devices_online/offline/total` 聚合字段与 `device_type_distribution`
- 时间字段统一为 ISO 8601 字符串格式

**策略自动调度器**：
- 新增 `internal/policy/scheduler.go`，支持两种触发模式：
  - 事件驱动：订阅 `telemetry:received` → 匹配 THRESHOLD 策略 → 自动执行
  - 定时扫描：每 30s 拉取 SCHEDULE 策略 → 评估 → 执行
- 含冷却期（60s）和冲突跳过机制（`COOLDOWN`/`CONDITION_NOT_MET`/`FAILED`）

**配置推送基础设施**：
- 新增 `internal/platform/mqtt/config_pusher.go`，支持通过 MQTT `cmd/config` topic 推送配置到设备
- 已接入 climate handler（profile/stage/action CRUD 后自动推送）
- 待固件定义配置协议格式后启用

**Handler 文件拆分**：
- 单文件不超过 400 行；所有 >1000 行的 handler 已拆分：
  - climate: `handler.go` + `profile_handler.go` + `stage_handler.go` + `action_handler.go`
  - policy: `execution_handler.go` + `policy_handler.go` + `condition_handler.go` + `target_handler.go`
  - nutrient: `handler.go` + `tank_handler.go` + `solution_handler.go` + `ion_test_handler.go` + `inventory_handler.go`
  - crop: `handler.go` + `variety_handler.go` + `stage_handler.go` + `batch_handler.go` + `harvest_handler.go`
- 所有方法仍共享同一 `Handler` 结构体接收者，路由注册不变

**enabled 字段类型统一**：
- 9 个模型的 `Enabled uint8` → `Enabled bool`（`gorm:"default:true"`）
- 涉及：ClimateProfile, ClimateStageAction, ControlPolicy, PolicyCondition, PolicyTarget, RecipeStageTarget, RecipeIonTarget, SensorChannel, ActuatorChannel
- DTO 中 `*uint8` → `*bool`，MySQL TINYINT 向后兼容无需迁移

### API 路径重构（扁平命名）

所有 API 路径采用扁平命名规范，不再使用嵌套式层级：
- `/api/sensor-devices`（非 `/api/devices/sensors`）
- `/api/actuator-devices`（非 `/api/devices/actuators`）
- `/api/commands`（非 `/api/controls/commands`）
- `/api/policies`（非 `/api/controls/policies`）
- `/api/greenhouses`、`/api/growing-zones`

### 共享契约更新

- `shared/docs/API_SPEC.md` 更新至 v2.0.0，包含 **193 个端点**，分布至 21 个模块路由组
- `shared/docs/openapi.yaml` 同步更新

---

## 1. 当前架构

### 1.1 18 个领域模块总览

项目根路径：`packages/backend/`

```
internal/
├── alert/          # 告警管理与处置闭环
├── audit/          # 审计日志
├── auth/           # JWT 认证 + RBAC（ADMIN/OPERATOR/VIEWER）
├── climate/        # 气候环境配置（多阶段控制）
├── command/        # 控制命令下发与回执追踪
├── crop/           # 作物品种、生长阶段、种植批次
├── device/         # 传感器/执行器设备与通道管理
├── energy/         # 能耗记录与汇总
├── greenhouse/     # 温室与种植区管理
├── metric/         # 指标定义字典与通道测点绑定
├── notification/   # 通知渠道（EMAIL/SMS/WEBHOOK）
├── nutrient/       # 营养液管理（液箱/换液/离子检测/浓缩液）
├── overview/       # 仪表盘聚合查询
├── pest/           # 病虫害观察与治理
├── platform/       # 基础设施（config/db/di/errors/http/influx/logger/mqtt/response/event）
├── policy/         # 控制策略引擎（阈值/定时/持续时长）
├── recipe/         # 营养液配方与批次绑定
├── review/         # 批次复盘快照
└── telemetry/      # 遥测数据采集与查询（InfluxDB + MySQL 双写）
```

### 1.2 模块职责与 API 路径对照

| # | 模块 | API 路径前缀 | 端点数 | 职责 |
|---|------|------------|--------|------|
| 1 | auth | `/api/auth`, `/api/users`, `/api/roles` | 9 | 登录/登出/用户 CRUD/角色 CRUD |
| 2 | overview | `/api/overview` | 1 | 仪表盘聚合（设备在线率、告警统计、遥测摘要） |
| 3 | greenhouse | `/api/greenhouses`, `/api/growing-zones` | 11 | 温室 CRUD、种植区 CRUD、温室嵌套分区 |
| 4 | device | `/api/sensor-devices`, `/api/sensor-channels`, `/api/actuator-devices`, `/api/actuator-channels` | 20 | 传感器/执行器设备与通道 CRUD |
| 5 | metric | `/api/metrics` | 2 | 指标定义字典查询 |
| 6 | telemetry | `/api/telemetry` | 5 | 遥测采集/实时查询/历史查询/删除 |
| 7 | command | `/api/commands` | 8 | 命令 CRUD、发送/确认、回执管理 |
| 8 | policy | `/api/policies`, `/api/policy-executions` | 22 | 策略 CRUD、条件/目标子资源、发布/归档/执行 |
| 9 | alert | `/api/alerts` | 7 | 告警 CRUD、统计、状态变更、时间线 |
| 10 | notification | `/api/notification-channels` | 5 | 通知渠道 CRUD + 测试发送 |
| 11 | audit | `/api/audit-logs` | 1 | 审计日志查询 |
| 12 | crop | `/api/crop-varieties`, `/api/growth-stages`, `/api/batches`, `/api/batch-stage-plans`, `/api/harvests` | 22 | 品种/生长阶段/批次/阶段计划/收获管理 |
| 13 | recipe | `/api/recipes`, `/api/recipe-bindings` | 13 | 营养液配方 CRUD、配方目标值、批次绑定 |
| 14 | climate | `/api/climate-profiles`, `/api/climate-execution-logs` | 17 | 环境配置多阶段管理、阶段动作、执行日志 |
| 15 | nutrient | `/api/nutrient-tanks`, `/api/solution-changes`, `/api/ion-tests`, `/api/concentrate-inventory`, `/api/concentrate-usage-logs` | 20 | 液箱/换液/离子检测/浓缩液库存与消耗 |
| 16 | energy | `/api/energy-records` | 8 | 能耗记录 CRUD、温室/批次/汇总查询 |
| 17 | pest | `/api/pest-observations`, `/api/treatment-records` | 15 | 病虫害观察 CRUD、治理记录 CRUD |
| 18 | review | `/api/reviews` | 7 | 复盘快照 CRUD、自动生成、批次快照查询 |

**合计：193 个 API 端点**

### 1.3 基础设施层（platform/）

```
internal/platform/
├── config/         # Viper 配置加载（configs/config.yaml）
├── db/             # MySQL 连接池初始化
├── di/             # 统一依赖注入结构体 Deps
├── errors/         # 业务错误码定义
├── event/          # 事件中心（EventHub）
├── http/           # Gin 路由组装（router.go）+ 中间件（CORS/RequestID/Logger）
├── influx/         # InfluxDB 客户端初始化
├── logger/         # slog 日志初始化
├── mqtt/           # MQTT 客户端初始化
└── response/       # 统一 JSON 响应封装
```

### 1.4 每个领域模块的标准文件结构

每个模块遵循统一约定。大型 handler（>400 行）已拆分为多个子 handler 文件：

```
internal/<module>/
├── model.go              # GORM 数据模型
├── dto.go                # 请求/响应 DTO
├── handler.go            # HTTP 处理器（Handler struct + NewHandler + 共享辅助函数）
├── routes.go             # 路由注册
├── *_handler.go          # 按资源拆分的子 handler（>400 行模块专用）
└── scheduler.go          # 自动调度器（policy 模块专用）
```

### 1.5 技术栈

| 层 | 技术 |
|----|------|
| 语言 | Go 1.24 |
| HTTP | Gin 1.11 |
| ORM | GORM 1.31（MySQL 驱动） |
| 时序数据库 | InfluxDB 2.7 |
| 消息中间件 | EMQX 5.6（MQTT，paho.mqtt.golang） |
| 鉴权 | JWT（golang-jwt/v5） |
| 校验 | go-playground/validator/v10 |
| 配置 | Viper |
| API 文档 | Swagger（swaggo） |
| 事件 | 内置 EventHub（platform/event） |

### 1.6 数据流

```
设备 (MQTT) -> EMQX Broker -> Backend (MQTT Client) -> InfluxDB (时序数据)
                                                      -> MySQL   (元数据 + 双写明细)
浏览器  <->  Frontend (Vue SPA)  <->  Backend API (Gin)  <->  MySQL  +  InfluxDB
```

### 1.7 环境变量与配置

| 变量前缀 | 用途 | 默认配置源 |
|----------|------|-----------|
| `HAMB_*` | 覆盖 configs/config.yaml 中的任意配置项 | `configs/config.yaml` |

关键配置项（部署前应变更）：
- `auth.jwt_secret` — JWT 签名密钥（默认 `change-me`）
- `influx.token` — InfluxDB 管理 Token（默认 `your-token`）

---

## 2. 关键设计决策

### 2.1 传感器/执行器设备拆分

v2.0.0 将设备分为两大类，各自承载独立的表结构：

- **SensorDevice + SensorChannel**：关注数据采集。通道包含 `metric_code`、`unit`、`range_min/max`、`sampling_interval_sec` 等采集相关字段。
- **ActuatorDevice + ActuatorChannel**：关注控制执行。通道包含 `actuator_type`（PUMP/AERATOR/FAN/VALVE/SHADE/LED/HEATER/CO2_GEN/FOGGER）、`current_state`（ON/OFF）、`rated_power_watt` 等执行相关字段。

设计理由：传感器和执行器在业务上语义完全不同，拆分为独立实体使遥测采集与控制命令下发路径各司其职，避免字段混杂和语义歧义。

### 2.2 Command + Policy 模式分离

控制域拆分为两个独立模块：

- **command** (`/api/commands`)：原子命令下发层。负责单条命令的创建、MQTT 发送、设备回执追踪。适用场景：手动操作单台设备。
- **policy** (`/api/policies`)：规则引擎层。负责策略定义（阈值/定时/持续时长）、条件组合、目标指定、自动执行记录。适用场景：自动化环境调控。

设计理由：命令是"怎么做"，策略是"什么时候做"。分离后命令模块保持轻量（单设备、一次操作），策略模块承载复杂规则编排。

### 2.3 扁平 API 路径命名

所有 API 路径采用扁平风格，以实体名为前缀：

```
/api/sensor-devices     （非 /api/devices/sensors）
/api/actuator-devices   （非 /api/devices/actuators）
/api/sensor-channels    （非 /api/channels/sensors）
/api/commands           （非 /api/controls/commands）
/api/policies           （非 /api/controls/policies）
/api/policy-executions  （非 /api/policies/executions）
/api/notification-channels（非 /api/channels/notifications）
```

设计理由：避免深层嵌套导致的路由歧义（如 `/api/channels` 是传感器通道还是通知渠道？），每个实体独占一个前缀，语义清晰。

### 2.4 无物理外键约束

所有 MySQL 表之间不创建 FOREIGN KEY 约束，依赖以下机制保证数据一致性：
- 逻辑关联字段（如 `greenhouse_id`、`actuator_device_id`、`batch_id`）
- 数据库索引（在关联字段上创建 INDEX 加速查询）
- 应用层校验（handler 中验证关联实体是否存在）

设计理由：避免物理外键带来的级联操作风险、迁移顺序依赖和运维成本，在应用层保持灵活可控。

### 2.5 MySQL + InfluxDB 双存储

| 数据类型 | 存储 | 用途 |
|----------|------|------|
| 遥测时序数据 | InfluxDB（主） + MySQL `telemetry_samples`（辅助） | InfluxDB 作为高性能时序存储，MySQL 作为降级回退与关联查询 |
| 业务元数据 | MySQL | 设备、温室、批次、策略、配方等所有领域实体 |
| 历史查询 | InfluxDB 优先，失败降级至 MySQL | `GET /api/telemetry/query` 支持 `source` 参数指定数据源 |

### 2.6 统一 API 响应格式

所有 API 返回统一的 JSON 信封：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_xxxx"
}
```

业务错误码体系：

| code | 含义 | HTTP |
|------|------|------|
| 0 | 成功 | 200/201 |
| 10001 | 参数校验失败 | 400 |
| 10002 | 未登录或 Token 无效 | 401 |
| 10003 | 权限不足 | 403 |
| 10004 | 资源不存在 | 404 |
| 10005 | 资源冲突 | 409 |
| 10006 | 频率限制 | 429 |
| 10007 | 设备离线 | 409 |
| 10008 | 规则冲突 | 409 |
| 10009 | 数据超出物理范围 | 422 |
| 10010 | 设备编码重复 | 409 |

### 2.7 RBAC 角色体系

| 角色 | 权限范围 |
|------|---------|
| ADMIN | 全量访问（用户管理、设备编辑、策略发布、系统配置） |
| OPERATOR | 查询 + 设备控制 + 告警处置 + 策略执行 |
| VIEWER | 仅查询类接口 |

---

## 3. 待办事项（前 5 项，按优先级排列）

1. 为 v2.0.0 新增模块补齐自动化测试覆盖（climate、command、crop、energy、nutrient、pest、policy、recipe、review 模块当前测试覆盖几乎为零）。
2. 检查所有更新/删除 handler 的 `RowsAffected` 返回值，对不存在的资源返回 404（当前多个 handler 未做此检查，可能返回 200）。
3. 通知模块 `dispatch` 逻辑与告警引擎的集成仍未完成 —— `evaluateAndTrigger` 创建告警后未实际调用 `go dispatchNotifications(alert)`。
4. ~~策略自动评估调度~~ — 已完成（`internal/policy/scheduler.go`，v2.3.0）。
5. 对齐并验证 MQTT/InfluxDB 在生产环境中的配置与启动行为。

---

## 4. 阻碍 / 风险

- **测试覆盖不足**：18 个模块中仅 alert、device、telemetry 有核心路径测试，其余 10 个新模块（climate/command/crop/energy/greenhouse/metric/nutrient/pest/recipe/review）无专门测试。大规模重构后的回归风险较高。
- **多模块 RowsAffected 缺失**：部分 handler 对不存在的资源更新/删除仍返回成功，前端可能误判操作结果。
- **通知 dispatch 未集成**：告警创建后通知渠道 dispatch 尚未串联，告警闭环不完整。
- **配置推送依赖固件**：MQTT 配置推送基础设施已就绪，但协议格式需与固件方对齐后启用。
- **迁移文件统一**：`migrations/merged/all.up.sql` 为唯一权威数据源，可直接执行初始化全量数据库。

---

## 5. 验证说明

- 执行 `go build ./...` 确保所有模块编译通过。
- 执行 `go test ./...` 运行全部测试（注意：部分测试依赖 MySQL 连接，需在 Docker 环境启动后执行）。
- 使用 `docker compose up -d` 启动基础设施（MySQL + InfluxDB + EMQX）。
- 运行 `migrations/merged/all.up.sql` 以初始化/重建全量数据库 schema。
- 使用 `go run cmd/api/main.go` 启动后端服务，访问 `http://localhost:3000/healthz` 验证存活状态。
- API 文档可通过 `http://localhost:3000/docs/index.html`（Swagger UI）查看。

---

## 6. 下个会话如何继续

1. 阅读 `docs/PROJECT_STATUS.md` —— 了解版本状态、已知缺口与风险。
2. 阅读本文件（`docs/HANDOFF.md`）—— 了解当前架构全貌与近期变更。
3. 从待办事项 #1 开始（补齐新模块测试），除非优先级发生变化。

---

## 7. 快速填写模板（每次交接使用）

- 日期：2026-05-07
- 分支：main
- 已完成范围：v2.0.0 架构重构 + v2.3.0 阶段三四（SSE 实时推送修通、序列化风格统一、策略自动调度器、配置推送基础设施、Handler 文件拆分、enabled 字段类型统一）。193 个 API 端点 + 2 个 SSE 订阅端点。
- 待完成范围：补齐新模块自动化测试、RowsAffected 全量修复、通知 dispatch 集成、固件配置协议对齐。
- 风险：新增模块测试覆盖几乎为零、通知 dispatch 未与告警引擎集成、配置推送协议待固件配合。
- 下个首要命令：`docker compose up -d && go build ./... && go run cmd/api/main.go`
