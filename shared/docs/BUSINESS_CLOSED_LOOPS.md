# 水培农业管理系统 - 业务逻辑闭环分析

> 核对日期: 2026-05-09
> 核对范围: 前后端全栈业务逻辑（以当前代码为准）

## 概述

系统中识别出 **12 个业务闭环**。本文逐个对照代码核验“真实入口/事件流/状态机/定时任务/SSE订阅”，并在每个闭环末尾给出面向落地与用户体验的改进建议。

---

## 一、已完整实现的闭环

### 1. 遥测数据管道（核心数据流）

这是驱动其他所有闭环的基础数据流。

```
设备MQTT上报 → IngressService.handleTelemetry()
  ├─ 验证设备是否存在（未知设备创建告警并丢弃）
  ├─ JSON 解析负载（支持单条/批量）
  ├─ 三重写入:
  │   ├─ MySQL（持久化存储）
  │   ├─ InfluxDB（时间序列存储）
  │   └─ 内存缓存（最新值快速读取）
  └─ 发布内部事件 "telemetry:received"
       ├─ SSE（telemetry_update）→ 前端实时图表更新
       └─ 策略调度器 → 阈值策略评估
```

**涉及文件:**
- `packages/backend/internal/platform/mqtt/ingress.go` (MQTT接收)
- `packages/backend/internal/telemetry/influx.go` (InfluxDB写入)
- `packages/backend/internal/telemetry/cache.go` (内存缓存)
- `packages/backend/internal/telemetry/handler.go` (查询API)
- `packages/backend/internal/platform/event/sse_handler.go` (内部事件→SSE事件映射/过滤)
- `packages/frontend/src/composables/useTelemetrySSE.ts` (前端SSE)

**核对结论（当前实现）:**
- 遥测入口为 `IngressService.handleTelemetry()`，写入 MySQL/Influx/内存缓存后发布 `telemetry:received`，SSE 对外事件名为 `telemetry_update`。
- 未知设备会创建“设备发现类告警”（当前实现的类型名为 `DEVICE_DISCOVERED`）并丢弃数据。
- SSE 遥测事件支持按 `device_codes`/`metric_codes` 过滤；payload 额外包含 `quality_flag`（前端类型已兼容可选字段）。

**差距与风险点:**
- “未知设备告警类型名/语义”与文档口径需要对齐（当前不是通用的 DEVICE_ERROR/OFFLINE）。
- SSE 遥测 payload 字段新增（如 `quality_flag`）后，历史文档未体现，容易造成前后端对齐误判。

**改进建议（UX + 落地）:**
- 统一遥测事件 payload schema：版本号 + 字段白名单，避免前端“静默缺字段/多字段”导致的图表异常。
- 未知设备处理增加“准入白名单/自动登记工作流”：告警里直接引导到资产入库页面（减少运维摩擦）。
- 对批量遥测解析失败提供可观测性（按设备/主题计数+最近错误样例），便于现场排障。

---

### 2. 策略驱动的自动化控制（核心自动化循环）

系统的"大脑"——将遥测数据与预定义策略连接，自动控制执行器。

```
两条触发路径:

路径A（事件驱动 - 阈值策略）:
  遥测事件到来 → evaluateThresholdPolicies()
    → 查询 enabled=true, policy_type="THRESHOLD" 且 metric_code 匹配的策略
    → 按优先级排序

路径B（定时器驱动 - 调度策略）:
  每30秒扫描 → evaluateScheduledPolicies()
    → 查询 enabled=true, policy_type="SCHEDULE" 且生效时间范围内的策略
    → 按优先级排序

条件评估:
  ├─ 冷却期检查（60秒内不重复执行）
  ├─ 阈值比较（支持滞后 hysteresis）
  ├─ 聚合函数（avg/max/min/last）+ 时间窗口
  ├─ 持续时间要求（RequiredDurationSec）
  └─ 条件满足 → 执行目标

目标执行:
  创建 ControlCommand(STATUS="PENDING")
  → JOIN actuator_channels → actuator_devices 获取 device_code
  → MQTT 发布到 hydroponic/<deviceCode>/cmd/<commandType>
  → 更新命令状态为 "SENT"
  → 发布 "command:dispatched" 事件
  → 设置冷却期

反馈确认:
  设备 MQTT ACK → handleAck()
  → 更新 control_commands 状态为 "ACKED"
  → 发布 "command:acked" 事件供 CommandWaiter 消费

执行记录:
  每次评估创建 PolicyExecution 记录
  (policy_id, trigger_source, decision, decision_reason)
```

**涉及文件:**
- `packages/backend/internal/policy/scheduler.go` (策略调度)
- `packages/backend/internal/policy/model.go` (数据模型)
- `packages/backend/internal/policy/subscriber.go` (历史/预留：当前主链路在 scheduler.go)
- `packages/backend/internal/command/handler.go` (命令下发)
- `packages/backend/internal/platform/mqtt/ingress.go` (ACK处理)
- `packages/frontend/src/views/controls/rules.vue` (策略管理UI)

**核对结论（当前实现）:**
- 策略调度器在 policy 模块路由注册时启动；事件驱动与 30s 定时扫描两条路径均已落地。
- 冷却期/滞后/聚合窗口/持续时间等判断逻辑在调度器内完整实现，并写入执行记录便于审计。

**差距与风险点:**
- `subscriber.go` 与 `scheduler.go` 同时存在，容易造成维护者误判“订阅入口”；建议在文档中明确“主链路”位置。
- 当 MQTT publish 不可用或失败时，命令状态与用户可见反馈需明确（避免 UI 显示已下发但设备未收到）。

**改进建议（UX + 落地）:**
- 给策略执行历史增加“命中原因可视化”：展示触发来源（device/channel/metric）、聚合值、阈值、滞后、持续时间等关键证据，便于现场调参。
- 策略冲突治理：同一执行器/通道在短时间被多个策略命中时，增加冲突检测与仲裁（优先级/互斥组/冷却共享）。
- 命令失败语义收敛：MQTT 不可用/发布失败时将命令标记为 FAILED，并把原因透传到前端（减少“假成功”）。

---

### 3. 命令生命周期（手动 + 自动）

三种命令下发模式：

```
模式A - 手动创建+发送:
  POST /commands → 创建 PENDING 命令
  POST /commands/:id/send → 标记 SENT + MQTT 发布

模式B - 同步调度（等待确认）:
  POST /commands/dispatch-and-wait
  → 创建命令 → 注册 CommandWaiter → MQTT 发布
  → 等待 ACK（10秒超时）
  → 返回 ACKED 或 TIMEOUT

模式C - 异步调度:
  POST /commands/dispatch-async
  → 创建命令 → MQTT 发布 → 立即返回 SENT
  → 发布 "command:dispatched" 事件

反馈链路:
  设备 ACK → handleAck()
  → "command:acked" 事件 → CommandWaiter 消费
  → MySQL 更新 STATUS="ACKED", acked_at=now
```

**涉及文件:**
- `packages/backend/internal/command/handler.go`
- `packages/backend/internal/command/waiter.go`
- `packages/backend/internal/command/model.go`
- `packages/backend/internal/command/routes.go`
- `packages/backend/internal/platform/mqtt/ingress.go` (ACK处理)

**核对结论（当前实现）:**
- 路由与三种下发模式均存在；ACK 通过内部事件 `command:acked` 唤醒等待者并更新命令状态。
- 命令状态枚举包含 PENDING/QUEUED/SENT/ACKED/TIMEOUT/FAILED（比旧文档描述更细）。

**差距与风险点:**
- ACK 事件的字段类型在不同解码路径下可能不一致（例如 `command_id` 数值类型），依赖 `map[string]interface{}` 取值存在脆弱点。

**改进建议（UX + 落地）:**
- 统一 ACK payload schema（强类型结构体 + 版本号）并在全链路使用同一解析方式，避免“偶发收不到 ACK”。
- 在前端命令列表/详情增加“生命周期时间线”（创建/发送/ACK/失败原因），并支持按设备/通道筛选，便于操作员追溯。
- 对同步等待接口提供“延长等待/二次查询”指引，避免 10 秒超时后用户误以为失败。

---

### 4. 告警生命周期

完整的状态机和工作流追踪。

```
告警创建（多入口）:
  ├─ 设备发现: 未知设备发布遥测/心跳 → createUnknownDeviceAlert() → 去重
  ├─ 设备错误: hydroponic/<deviceCode>/errors → handleErrors() → 创建 DEVICE_ERROR 告警
  ├─ 设备离线: offline_detector 定期扫描 → 创建 DEVICE_OFFLINE 告警
  └─ 手动 API: POST /alerts → 操作员手动创建

告警处理:
  创建 → MySQL 持久化
  → 自动创建 AlertTimelineEvent(type="TRIGGERED")
  → 发布 "alert:created" 事件
     ├─ SSE（new_alert）→ 前端实时通知（CRITICAL 级别触发浏览器桌面通知）
     └─ 通知订阅者 → Webhook / Email / SMS

状态流转:
  OPEN → ACKNOWLEDGED → RESOLVED | IGNORED
  每次状态变更 → 创建时间线事件
  事务保证原子性

自动恢复:
  设备重新发送心跳 → 自动解决 DEVICE_OFFLINE 告警
  将告警状态更新为 RESOLVED
```

**涉及文件:**
- `packages/backend/internal/alert/model.go`
- `packages/backend/internal/alert/handler.go`
- `packages/backend/internal/alert/routes.go`
- `packages/backend/internal/platform/event/sse_handler.go`
- `packages/frontend/src/composables/useAlertSSE.ts`
- `packages/frontend/src/views/alerts/index.vue`

**核对结论（当前实现）:**
- 告警创建入口（未知设备/设备错误/离线检测/手动）均存在，且手动创建会写入 TRIGGERED 时间线并发布 `alert:created`。
- SSE 对外事件名为 `new_alert`（由 `alert:created` 映射）。
- 状态流转接口会写时间线事件，满足审计需求。

**差距与风险点（需更新到文档并建议优先修）:**
- 后端 SSE 目前对告警事件未做过滤，而前端订阅时传了 `device_id/level` 等参数，存在“用户以为过滤有效但实际无效”的体验落差。
- `new_alert` 的 SSE payload 字段与前端 `AlertEvent` 类型假设存在不一致风险（可能只有 `alert_id` 等最小字段），导致桌面通知 tag、列表刷新等出现异常。
- 心跳自动恢复离线告警时，当前实现只更新告警状态，未补写对应时间线事件，导致审计链路断点。

**改进建议（UX + 落地）:**
- SSE 告警事件支持过滤（至少 level/device_id 或 device_code），并与前端参数口径对齐；无效参数要显式报错或忽略说明。
- 统一告警 SSE payload：直接复用“告警详情 DTO”（包含 id、level、status、triggered_at、title/summary 等），避免前端二次查询造成的闪烁。
- 自动恢复也写入时间线（RESOLVED + reason=heartbeat），保证闭环审计完整，并可用于统计“自愈率”。

---

### 5. 设备健康监测（心跳 + 离线检测）

```
心跳接收:
  设备发布 hydroponic/<deviceCode>/heartbeat
  → 查找 device_code 归属
  → 未知设备 → createUnknownDeviceAlert()
  → 已知设备 → 更新 last_seen_at=now, status="ONLINE"
  → 自动解决任何打开的 DEVICE_OFFLINE 告警

状态上报:
  设备发布 hydroponic/<deviceCode>/status
  → 解析状态（ONLINE/OFFLINE/ERROR）
  → 更新 sensor_devices 和 actuator_devices
  → 发布 "device:status" 事件 → SSE → 前端

离线检测:
  定期扫描 last_seen_at 超过阈值的设备
  → 标记为 OFFLINE
  → 创建 DEVICE_OFFLINE 告警

仪表盘聚合:
  并行查询传感器/执行器的在线/总数
  → 计算离线数
  → 前端设备列表展示在线/离线状态
```

**涉及文件:**
- `packages/backend/internal/platform/mqtt/ingress.go` (心跳/状态处理)
- `packages/backend/internal/device/offline_detector.go` (离线检测)
- `packages/backend/internal/device/model.go`
- `packages/backend/internal/overview/handler.go` (仪表盘聚合)
- `packages/frontend/src/views/devices/list.vue`
- `packages/frontend/src/views/devices/detail.vue`

**核对结论（当前实现）:**
- 心跳与状态 topic 均由 MQTT ingress 处理；离线检测为后台定时任务（默认 30s 扫描）并会创建离线告警与发布 `device:status`。

**差距与风险点:**
- 设备 status 上报值缺少强校验，异常值可能污染状态字段并影响 UI 判断。

**改进建议（UX + 落地）:**
- 增加离线/恢复去抖（debounce）与阈值配置化：避免边缘网络抖动导致频繁离线告警。
- 设备详情页强化“最后心跳/最近状态变更/原因”展示，并支持一键跳转到相关告警/命令历史。

---

### 6. 仪表盘概览（系统健康汇总）

```
触发: 用户访问仪表盘页面或周期性刷新

并发查询（WaitGroup）:
  ├─ 在线传感器数量
  ├─ 传感器总数
  ├─ 在线执行器数量
  ├─ 执行器总数
  ├─ 打开的告警数量
  ├─ 严重告警数量
  ├─ 今日告警数量
  ├─ 温室汇总（JOIN 多表）
  └─ 最近5条命令

结果组装 → DashboardResponse DTO
  ├─ 设备在线/离线/总数
  ├─ 告警统计
  ├─ 温室概览（平均温度、湿度）
  └─ 最近命令列表

前端渲染:
  关键指标卡片 → 温室概览卡片 → 最近命令表格
  → 操作员据此导航到各模块进行操作
```

**涉及文件:**
- `packages/backend/internal/overview/handler.go`
- `packages/backend/internal/overview/dto.go`
- `packages/frontend/src/views/dashboard/index.vue`

**核对结论（当前实现）:**
- 仪表盘聚合在后端通过 WaitGroup 并发查询后组装 DTO 返回；前端渲染为关键指标卡片与列表。

**差距与风险点（实现与旧文档不一致）:**
- 旧文档写“8路并发”，当前代码实际启动的并发任务数与错误通道容量不一致，存在极端情况下写满阻塞的风险（属于潜在稳定性问题）。

**改进建议（UX + 落地）:**
- 后端并发聚合改为：按实际 goroutine 数设置 error channel 缓冲，或改用 errgroup；并在超时/部分失败时返回“部分可用”的字段级错误，避免整页空白。
- 前端仪表盘给出“数据更新时间/数据来源”提示，减少用户对实时性与延迟的误解。

---

## 二、部分实现的闭环

### 7. 气候控制（配置 → 调度 → 执行）

```
配置阶段:
  操作员定义气候联动 Profile（含触发通道、阶段阈值、执行动作）→ /api/climate/...

调度（已集成启动）:
  ProfileScheduler 启动后订阅 "telemetry:received"
  → 按 profile.trigger_sensor_channel_id 过滤
  → 评估当前 stage（阈值+滞后）
  → 命中后创建命令并 MQTT 下发

动作执行:
  关联 actuator_channel 和 command_type
  → MQTT 下发命令到设备

⚠️ 缺失:
  - 当前气候联动以“单一触发通道 + 阈值/滞后”为核心，不等同于“多指标目标范围控制”（温/湿/CO2/光照闭环）完整版
  - 与作物批次阶段计划（#9/#10）尚无自动联动
```

**涉及文件:**
- `packages/backend/internal/climate/profile_handler.go`
- `packages/backend/internal/climate/stage_handler.go`
- `packages/backend/internal/climate/action_handler.go`
- `packages/backend/internal/climate/profile_scheduler.go` (事件驱动调度 + 命令下发)
- `packages/backend/internal/climate/model.go`

**核对结论（当前实现）:**
- ProfileScheduler 已随 climate 模块注册而启动；主链路为 `telemetry:received → scheduler → MQTT 下发 → 命令落库`，已形成可运行闭环。

**改进建议（UX + 落地）:**
- 增加“联动执行记录”页：按 profile 展示命中次数、最近触发值、下发命令、ACK/失败原因，便于调参与验收。
- 冲突与互斥：同一执行器通道被多个 profile/actions 命中时增加互斥组与优先级策略，避免“抖动控制”。
- 与批次阶段联动：允许批次进入某阶段时自动启用/切换对应气候 profile（把 #9/#10 的缺口闭合）。

---

### 8. MQTT 配置推送（系统 → 设备配置同步）

```
触发: 配置变更（当前仅气候 profile 变更已接入）

处理:
  PushToDevice(deviceCode, cfgType, action, entityID, payload)
  → 组装 ConfigPushPayload {config_type, action, entity_id, payload}
  → 发布到 hydroponic/<deviceCode>/cmd/config

通过执行器通道推送:
  PushToActuatorChannel(actuatorChannelID, ...)
  → JOIN actuator_channels → actuator_devices 获取 device_code
  → 调用 PushToDevice()

⚠️ 缺失:
  - 设备端配置确认回执（config-ack）链路未实现
  - 目前仅气候 profile 的变更触发推送，且 payload 为空（仅通知变更，不传完整配置）
  - 无投递记录/重试/幂等控制，难以做“可靠配置同步”
```

**涉及文件:**
- `packages/backend/internal/platform/mqtt/config_pusher.go`
- `packages/backend/internal/platform/mqtt/topics.go`
- `packages/backend/internal/climate/profile_handler.go` (profile变更触发推送)

**改进建议（UX + 落地）:**
- 明确定义“配置同步协议”：cfgType/action/entity_id + schema_version + payload；设备端回执 topic 与字段要强类型化并可审计。
- 引入投递流水与重试（deliveries/outbox），并在前端展示“配置已下发/设备已确认/失败重试中”，把不可见的不确定性变成可管理的状态。
- 扩展接入点：策略（policy）、批次阶段（batch stage）、配方目标（recipe targets）等均通过统一 pusher 下发，减少设备端多协议成本。

---

### 9. 阶段到设备配置联动

```
触发: 操作员在 StagePlanEditor 中创建/更新阶段转换计划

处理:
  定义转换事件（触发条件、目标环境参数、目标营养参数、要执行的动作）
  → 阶段激活或转换发生时
  → PushToDevice(deviceCode, "crop_batch", "update", batchID, stageConfig)
  → 设备接收新设定值并调整执行器

⚠️ 缺失:
  - 当前仅有阶段计划（时间窗+目标值）CRUD/UI，尚无“到期自动切换/自动激活下一阶段”的调度器
  - crop 模块未接入 ConfigPusher，阶段配置未实际推送到设备
```

**涉及文件:**
- `packages/backend/internal/crop/stage_handler.go`
- `packages/backend/internal/crop/harvest_handler.go` (BatchStagePlan CRUD)
- `packages/backend/internal/platform/mqtt/config_pusher.go`
- `packages/frontend/src/components/batch/StagePlanEditor.vue`

**改进建议（UX + 落地）:**
- 新增批次阶段调度器：按时间窗推进当前阶段，推进时写入批次事件日志并可回滚（人工覆盖）。
- 阶段配置推送落地：明确“批次绑定的设备/通道”与“下发 payload 结构”，复用 #8 的可靠投递机制。
- 前端提供“下一阶段预览/到期提醒/一键手动推进”并展示设备确认状态，减少操作员盯盘压力。

---

## 三、需人工干预的半自动闭环

### 10. 作物批次生命周期

```
操作员创建品种 → 创建批次(选品种/温室/区域)
  → 定义阶段规划(幼苗→营养生长→开花→成熟)
  → 设定阶段/计划（当前以手动维护为主）
  → 批次状态转换（当前为手动触发）
  → 收获记录(实际产量/质量等级)
  → 台账分析(计划 vs 实际对比)

流程特点:
  - 核心是“记录 + 复盘”，自动联动尚未闭合
  - 批次状态转换为显式接口触发，未与阶段计划到期自动推进联动
```

**涉及文件:**
- `packages/backend/internal/crop/variety_handler.go`
- `packages/backend/internal/crop/batch_handler.go`
- `packages/backend/internal/crop/stage_handler.go`
- `packages/backend/internal/crop/harvest_handler.go`
- `packages/frontend/src/views/batches/detail.vue` (状态转换入口)
- `packages/frontend/src/views/batches/harvest.vue`
- `packages/frontend/src/views/batches/ledger.vue`

**核对结论（当前实现）:**
- 批次 CRUD/状态转换/阶段计划/采收与台账页面均已存在；现阶段仍以“人工驱动闭环”为主，自动化联动（到期推进、气候/营养联动）未落地。

**改进建议（UX + 落地）:**
- 把“批次状态/阶段计划/配方绑定/气候 profile 绑定”纳入同一批次视图，用向导式流程减少遗漏配置。
- 增加批次级快照与审计：阶段切换时记录当前配置（配方/策略/气候 profile 版本）用于复盘对比。
- 与 #7/#8/#9 打通后，形成“阶段→配置下发→设备确认→效果监控”的完整自动化闭环。

---

### 11. 营养管理

```
配方定义(recipe模块) → 水箱管理(创建A/B罐/混合罐/pH调节罐)
  → 溶液更换(记录EC/pH/化学品用量)
  → 离子测试(N/P/K/Ca/Mg/Fe等)
  → 与配方目标对比
  → 差异分析和调整建议
  → 操作员手动调整 → 重新测试

流程特点:
  - 纯手动操作循环
  - 无自动营养投加闭环
  - 无基于测试结果的自动配方调整
```

**涉及文件:**
- `packages/backend/internal/nutrient/tank_handler.go`
- `packages/backend/internal/nutrient/solution_handler.go`
- `packages/backend/internal/nutrient/ion_test_handler.go`
- `packages/backend/internal/nutrient/inventory_handler.go`
- `packages/backend/internal/recipe/handler.go`
- `packages/frontend/src/views/nutrient/tanks.vue`
- `packages/frontend/src/views/nutrient/ion-tests.vue`
- `packages/frontend/src/views/recipes/index.vue`

**核对结论（当前实现）:**
- 后端已提供 tanks、solution-changes、ion-tests、inventory、usage-logs 等记录型接口；前端当前主要覆盖 tanks 与 ion-tests（其余能力尚未形成完整 UI 闭环）。

**改进建议（UX + 落地）:**
- 补齐“换液/补液/库存/用量日志”前端页面，使后端能力可被真实使用，否则会形成“有接口无闭环”。
- 增加服务端“差异分析 DTO”：离子检测 vs 配方 targets 输出 delta、风险等级与建议（可先只给建议，不自动投加）。
- 若要走向自动化：在差异分析后可选生成 dosing 命令，复用现有命令系统（并加上安全阈值与人工确认）。

---

### 12. 通知分发（告警 → 外部系统）

```
配置: 操作员创建通知通道(Webhook/Email/SMS/InApp)
  → 设置 MinAlertLevel(INFO/WARN/CRITICAL)
  → 启用/禁用切换

分发: 订阅 "alert:created" 事件
  → 查询所有启用通道
  → 级别过滤(告警级别 >= 通道 MinAlertLevel)
  → 路由到通道实现:
     ├─ Webhook:  HTTP POST + 可选 HMAC-SHA256 签名 ✓
     ├─ Email:    仅记录日志 ✗
     ├─ SMS:      仅记录日志 ✗
     └─ InApp:    仅记录日志 ✗

⚠️ 缺失:
  - Email、SMS、InApp 通道仅存框架
  - 无通知发送状态追踪和重试机制
```

**涉及文件:**
- `packages/backend/internal/notification/handler.go`
- `packages/backend/internal/notification/subscriber.go`
- `packages/backend/internal/notification/model.go`
- `packages/backend/internal/notification/routes.go`
- `packages/frontend/src/views/settings/notification-channels.vue`
- `packages/frontend/src/api/notification.ts`
- `packages/frontend/src/types/notification.ts`

**核对结论（当前实现）:**
- 通知订阅在 notification 模块注册时启动；Webhook 会真实投递，其它通道仍为占位实现。
- 前端通道类型与后端枚举存在不一致（后端包含 IN_APP，但前端未暴露/不可配置）。

**改进建议（UX + 落地）:**
- 通知可靠性闭环：引入投递记录（deliveries）与重试策略，并在前端展示“已发送/失败/重试中/最终失败”。
- 权限与安全：测试/编辑通道必须按 user_id/RBAC 做资源级校验，避免越权操作。
- InApp 落地：可复用现有 SSE/告警中心，提供“通知中心”列表与已读状态，满足无外部渠道时的闭环需求。

---

## 跨模块事件流全景

```
设备（MQTT）
    │
    ▼
MQTT IngressService
    │
    ├── handleTelemetry()
    │   ├── MySQL (持久化)
    │   ├── InfluxDB (时间序列)
    │   ├── 内存缓存 (最新值)
    │   └── Event Hub: "telemetry:received"
    │       ├── 策略调度器 → 阈值检查 → 命令创建 → MQTT下发 → 设备执行
    │       ├── 气候调度器 → 触发通道命中 → 命令创建 → MQTT下发 → 设备执行
    │       └── SSE 端点 → 前端实时更新
    │
    ├── handleHeartbeat()
    │   ├── 更新 last_seen_at
    │   └── 解决 DEVICE_OFFLINE 告警
    │
    ├── handleStatus()
    │   ├── 更新设备状态
    │   └── Event Hub: "device:status"
    │
    ├── handleErrors()
    │   ├── 创建 DEVICE_ERROR 告警
    │   └── Event Hub: "alert:created"
    │
    └── handleAck()
        ├── Event Hub: "command:acked" → CommandWaiter 消费
        └── 更新命令状态为 "ACKED"

告警生命周期:
    Event Hub: "alert:created"
    ├── SSE（new_alert）→ 前端通知（CRITICAL 级别触发桌面通知）
    ├── 通知订阅者 → Webhook / Email / SMS
    └── 操作员确认/解决 → 状态更新 → 时间线追踪

用户操作（HTTP API）:
    前端（Vue）→ REST API → Gin 处理器 → MySQL 读写 → 响应
    前端 SSE: EventSource → /api/alerts/subscribe 和 /api/telemetry/subscribe
```

---

## 闭环状态总览

| # | 闭环名称 | 类型 | 状态 | 关键缺失 |
|---|---------|------|------|---------|
| 1 | 遥测数据管道 | 自动化 | ✅ 完整 | - |
| 2 | 策略自动化控制 | 自动化 | ✅ 完整 | - |
| 3 | 命令生命周期 | 自动化 | ✅ 完整 | - |
| 4 | 告警生命周期 | 自动化 | ✅ 核心完整 | SSE过滤/统一payload/自动恢复写时间线 |
| 5 | 设备健康监测 | 自动化 | ✅ 完整 | - |
| 6 | 仪表盘概览 | 查询聚合 | ✅ 完整 | - |
| 7 | 气候控制 | 自动化控制 | ✅ 已落地（现有模型） | 多指标目标控制/批次联动 |
| 8 | MQTT配置推送 | 自动化 | 🔶 部分 | ACK/可靠投递/payload与扩展接入点 |
| 9 | 阶段→设备配置联动 | 半自动 | 🔸 缺失 | 自动调度/推送链路 |
| 10 | 作物批次生命周期 | 手动 | 🔸 人工 | 到期推进/策略-气候-营养联动 |
| 11 | 营养管理 | 手动 | 🔸 人工 | 差异分析建议/（可选）自动投加 |
| 12 | 通知分发 | 自动化 | 🔶 部分 | Email/SMS/InApp落地 + 投递追踪重试 |

---

## 关键架构要点

1. **Event Hub 是系统的神经中枢** — 发布/订阅模式实现模块间解耦，使遥测→策略→命令→执行的端到端自动化成为可能，无需模块间直接耦合。

2. **最核心的自动化闭环是 #2（策略自动化控制）** — 它实现了感知→决策→执行→反馈的完整闭环，是系统的"大脑"。

3. **#8、#9 是实现“批次阶段自动联动”的关键缺口** — 补齐可靠配置推送与阶段调度后，批次/气候/营养可以从“记录型”迈向“自动化执行+可审计”。

4. **系统整体以监控为主，自动化执行为辅** — 作物全生命周期、营养管理、通知分发等环节仍以操作员手动操作为主，自动化程度有较大提升空间。
