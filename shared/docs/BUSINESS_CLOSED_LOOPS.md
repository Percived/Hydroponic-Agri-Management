# 水培农业管理系统 - 业务逻辑闭环分析

> 分析日期: 2026-05-08
> 分析范围: 前后端全栈业务逻辑

## 概述

系统中识别出 **12 个业务闭环**，按完成度分为三类：完整实现、部分实现、需人工干预的半自动闭环。

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
  └─ 发布 "telemetry:received" 事件
       ├─ SSE → 前端实时图表更新
       └─ 策略调度器 → 阈值策略评估
```

**涉及文件:**
- `packages/backend/internal/platform/mqtt/ingress.go` (MQTT接收)
- `packages/backend/internal/telemetry/influx.go` (InfluxDB写入)
- `packages/backend/internal/telemetry/cache.go` (内存缓存)
- `packages/backend/internal/telemetry/handler.go` (查询API)
- `packages/frontend/src/composables/useTelemetrySSE.ts` (前端SSE)

**闭环特性:** 设备遥测数据进入系统 → 经过验证和存储 → SSE 实时推送前端 + 历史查询 API → 操作员监控和诊断系统状态。

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
- `packages/backend/internal/policy/subscriber.go` (事件订阅)
- `packages/backend/internal/command/handler.go` (命令下发)
- `packages/backend/internal/platform/mqtt/ingress.go` (ACK处理)
- `packages/frontend/src/views/controls/rules.vue` (策略管理UI)

**闭环特性:** 传感器遥测（或调度定时器）触发策略评估 → 条件与阈值比较 → 匹配时通过 MQTT 下发命令到执行器 → 设备执行并 ACK 确认 → 执行记录持久化供审核 → 操作员查看执行历史调整策略。

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

**闭环特性:** 命令创建(手动/自动) → MQTT 下发 → 设备执行 → ACK 回执 → 状态更新。同步模式显式等待完成；异步模式发布事件供追踪。

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
     ├─ SSE → 前端实时通知（CRITICAL 级别触发浏览器桌面通知）
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

**闭环特性:** 系统/设备/手动创建告警 → SSE 实时推送前端 → 操作员确认/解决/忽略 → 时间线追踪处理过程 → 设备恢复后自动解决 → 告警统计反馈到仪表盘。

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

**闭环特性:** 设备发送心跳 → 后端更新状态并解决告警 → 前端显示当前状态 → 操作员看到离线设备 → 手动或自动恢复设备。

---

### 6. 仪表盘概览（系统健康汇总）

```
触发: 用户访问仪表盘页面或周期性刷新

8路并发查询（WaitGroup）:
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

**闭环特性:** 仪表盘汇总所有模块数据 → 操作员看到系统全局状态 → 操作员导航到各模块处理问题 → 操作变动改变系统状态 → 仪表盘刷新反映变化。

---

## 二、部分实现的闭环

### 7. 气候控制（配置 → 调度 → 执行）

```
配置阶段:
  操作员定义气候配置 → POST /climate-profiles
  设置目标参数（温度/湿度/CO2/光照）及范围
  定义阶段（白天/夜间、幼苗/营养生长/开花）
  定义动作（温度过高→开风扇、湿度过低→开加湿器）

调度（部分实现）:
  profile_scheduler.go 已创建
  定期或事件驱动的评估
  将当前遥测与气候设定值比较
  超出范围 → 执行已配置动作

动作执行:
  关联 actuator_channel 和 command_type
  → MQTT 下发命令到设备

⚠️ 缺失:
  - 调度器自动调度逻辑未完全集成
  - 阶段切换与气候配置的自动联动未实现
```

**涉及文件:**
- `packages/backend/internal/climate/profile_handler.go`
- `packages/backend/internal/climate/stage_handler.go`
- `packages/backend/internal/climate/action_handler.go`
- `packages/backend/internal/climate/profile_scheduler.go` (部分实现)
- `packages/backend/internal/climate/model.go`

---

### 8. MQTT 配置推送（系统 → 设备配置同步）

```
触发: 任何配置变更（气候/策略/营养/作物）

处理:
  PushToDevice(deviceCode, cfgType, action, entityID, payload)
  → 组装 ConfigPushPayload {config_type, action, entity_id, payload}
  → 发布到 hydroponic/<deviceCode>/cmd/config

通过执行器通道推送:
  PushToActuatorChannel(actuatorChannelID, ...)
  → JOIN actuator_channels → actuator_devices 获取 device_code
  → 调用 PushToDevice()

⚠️ 缺失:
  - 设备端配置确认回执（handleConfigAck）暂未实现
  - 配置变更触发推送的集成点未完全打通
```

**涉及文件:**
- `packages/backend/internal/platform/mqtt/config_pusher.go`
- `packages/backend/internal/platform/mqtt/topics.go`

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
  - 阶段转换的自动触发调度逻辑未完全集成
  - 阶段配置到设备推送的完整链路未打通
```

**涉及文件:**
- `packages/backend/internal/crop/stage_handler.go`
- `packages/backend/internal/platform/mqtt/config_pusher.go`
- `packages/frontend/src/components/batch/StagePlanEditor.vue`

---

## 三、需人工干预的半自动闭环

### 10. 作物批次生命周期

```
操作员创建品种 → 创建批次(选品种/温室/区域)
  → 定义阶段规划(幼苗→营养生长→开花→成熟)
  → 设定每个阶段参数(温度/湿度/光照/营养)
  → 阶段转换(手动完成当前阶段→自动激活下一阶段)
  → 收获记录(实际产量/质量等级)
  → 台账分析(计划 vs 实际对比)

流程特点:
  - 全流程手动操作
  - 阶段转换未自动联动气候/营养策略
  - 无自动环境控制反馈
```

**涉及文件:**
- `packages/backend/internal/crop/variety_handler.go`
- `packages/backend/internal/crop/batch_handler.go`
- `packages/backend/internal/crop/stage_handler.go`
- `packages/backend/internal/crop/harvest_handler.go`
- `packages/frontend/src/views/batches/harvest.vue`
- `packages/frontend/src/views/batches/ledger.vue`

**闭环特性:** 操作员创建批次 → 品种和阶段设定预期 → 收获记录与计划比较 → 台账提供历史数据用于未来规划。自动化控制环节缺失。

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
    ├── SSE 端点 → 前端通知（CRITICAL 级别触发桌面通知）
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
| 4 | 告警生命周期 | 自动化 | ✅ 完整 | - |
| 5 | 设备健康监测 | 自动化 | ✅ 完整 | - |
| 6 | 仪表盘概览 | 查询聚合 | ✅ 完整 | - |
| 7 | 气候控制 | 半自动 | 🔶 部分 | 调度器自动逻辑 |
| 8 | MQTT配置推送 | 自动化 | 🔶 部分 | 设备确认回执 |
| 9 | 阶段→设备配置联动 | 半自动 | 🔶 部分 | 触发调度逻辑 |
| 10 | 作物批次生命周期 | 手动 | 🔸 人工 | 阶段自动联动策略 |
| 11 | 营养管理 | 手动 | 🔸 人工 | 自动投加闭环 |
| 12 | 通知分发 | 自动化 | 🔶 部分 | Email/SMS实现 |

---

## 关键架构要点

1. **Event Hub 是系统的神经中枢** — 发布/订阅模式实现模块间解耦，使遥测→策略→命令→执行的端到端自动化成为可能，无需模块间直接耦合。

2. **最核心的自动化闭环是 #2（策略自动化控制）** — 它实现了感知→决策→执行→反馈的完整闭环，是系统的"大脑"。

3. **#8、#9 是实现气候/阶段自动化控制的关键缺口** — 补齐这两个闭环后，作物批次、气候控制、营养管理可从手动转为自动化。

4. **系统整体以监控为主，自动化执行为辅** — 作物全生命周期、营养管理、通知分发等环节仍以操作员手动操作为主，自动化程度有较大提升空间。
