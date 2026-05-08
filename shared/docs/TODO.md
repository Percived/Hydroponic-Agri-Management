# 迭代待办清单 (TODO List)

> 本文档记录水培农业管理系统当前迭代中发现的需求缺口和优化项，按优先级排列。

---

## 一、策略控制模块

### 1.1 SCHEDULE 定时策略增强

**现状**：SCHEDULE 策略每 30 秒轮询一次条件，无法指定具体执行时间点。

**缺口**：
- 不支持 cron 表达式（如"每天 6:00 执行"）
- 不支持自定义执行间隔（目前硬编码 30s 扫描周期）
- 不支持"每周一三五执行"等周期规则

**建议方案**：
- [ ] `ControlPolicy` 模型新增 `CronExpr` 字段
- [ ] Scheduler 增加 cron 解析和定时触发 goroutine
- [ ] 前端 SCHEDULE 表单新增 cron 输入框

**涉及文件**：
- `packages/backend/internal/policy/model.go`
- `packages/backend/internal/policy/scheduler.go`
- `packages/backend/internal/policy/dto.go`
- `packages/frontend/src/views/controls/rules.vue`
- `packages/frontend/src/types/policy.ts`

---

### 1.2 DURATION 持续策略实现

**现状**：模型和 DTO 已定义，前端下拉已禁用，Scheduler 完全没有实现。

**含义**：策略触发后连续执行 N 分钟，到期自动停止。例如"温度 > 30°C 时开启风扇，持续运行 10 分钟后关闭"。

**建议方案**：
- [ ] `ControlPolicy` 模型新增 `DurationSec` 字段
- [ ] Scheduler 实现持续策略：触发 → 记录开始时间 → 到期停止 → 反向命令
- [ ] 前端 DURATION 表单设计（需要开始条件 + 持续时长）

**涉及文件**：
- `packages/backend/internal/policy/model.go`
- `packages/backend/internal/policy/scheduler.go`
- `packages/backend/internal/policy/dto.go`
- `packages/frontend/src/views/controls/rules.vue`

---

### 1.3 策略冲突检测

**现状**：`PolicyExecution.Decision` 枚举了 `CONFLICT`，但从未产生。多策略同时触发同一执行器时没有互斥检查。

**场景**：策略 A（温度 > 30 → 开风扇）和策略 B（CO2 > 1200 → 关风扇）同时触发同一风扇。

**建议方案**：
- [ ] Scheduler 执行前查询目标执行器的"当前活跃策略"
- [ ] 按优先级裁决或记录冲突
- [ ] PolicyExecution 中记录 CONFLICT 决策

**涉及文件**：
- `packages/backend/internal/policy/scheduler.go`

---

### 1.4 RequiredDurationSec 完善

**现状**（v1 已实现）：内存中跟踪条件持续时间，进程重启后状态丢失。

**缺口**：
- 内存跟踪在服务重启后归零
- 执行记录中无法区分 `CONDITION_NOT_MET` 和 `DURATION_NOT_MET`

**建议方案**：
- [ ] 将条件状态持久化到 MySQL（或接受内存方案的局限性）
- [ ] 执行日志 reason 字段区分"条件不满足"和"持续时间不足"

**涉及文件**：
- `packages/backend/internal/policy/scheduler.go`

---

## 二、告警模块

### 2.1 遥测数据质量异常自动告警

**现状**：遥测 `quality_flag = 'out_of_range'` 仅作为数据标记，不会触发任何告警。

**缺口**：模拟器注入的异常数据（如温度 40°C、湿度 20%）只在数据库里有个标记，前端看不到告警。

**建议方案**：
- [ ] `IngressService.handleTelemetry()` 中检测 `quality_flag != "normal"`
- [ ] 自动创建 `DEVICE_ANOMALY` 类型告警
- [ ] 或复用 `THRESHOLD` 类型，关联到对应 sensor_channel

**涉及文件**：
- `packages/backend/internal/platform/mqtt/ingress.go`
- `packages/backend/internal/alert/model.go`
- `packages/backend/internal/telemetry/handler.go`

---

### 2.2 告警类型扩展

**现状**：只有 3 种告警类型 `THRESHOLD` / `DEVICE_OFFLINE` / `SYSTEM`。

**建议新增**：
- [ ] `DEVICE_ANOMALY` — 设备数据异常
- [ ] `DEVICE_DISCOVERED` — 已在 ingress 中使用但未列入 alert 常量
- [ ] `POLICY_CONFLICT` — 策略冲突

**涉及文件**：
- `packages/backend/internal/alert/model.go`

---

## 三、设备模拟器

### 3.1 多设备并行模拟

**现状**：模拟器一次只模拟一个设备。种子数据有 5 个设备。

**建议方案**：
- [ ] 支持 `--device` 接受逗号分隔的多个设备编码
- [ ] 或启动多个 goroutine，每个模拟一个设备
- [ ] 支持 `--count=N` 自动注册并模拟 N 个设备

**涉及文件**：
- `packages/backend/cmd/simulator/main.go`

---

### 3.2 执行器模拟支持

**现状**：模拟器只模拟传感器（上报遥测），不支持执行器（接收命令并回复 ACK）。

**建议方案**：
- [ ] 模拟器支持 `--type=actuator` 参数
- [ ] 注册执行器设备 + 通道
- [ ] 订阅 `cmd/#` 主题，收到命令后执行模拟动作并回复 ACK

**涉及文件**：
- `packages/backend/cmd/simulator/main.go`

---

## 四、前端优化

### 4.1 策略表单条件多行支持

**现状**：表单只支持添加一个条件，但后端支持多条件 AND 逻辑。

**建议方案**：
- [ ] 条件区域改为动态列表，支持"添加条件"按钮
- [ ] 每个条件独立配置指标/运算符/阈值

**涉及文件**：
- `packages/frontend/src/views/controls/rules.vue`

---

### 4.2 策略目标多行支持

**现状**：同上，只支持一个目标，但后端支持多目标顺序执行。

**建议方案**：
- [ ] 目标区域改为动态列表，支持"添加目标"按钮
- [ ] 拖拽排序 execution_order

**涉及文件**：
- `packages/frontend/src/views/controls/rules.vue`

---

### 4.3 策略执行历史展示

**现状**：列表页没有展示每条策略的执行记录。

**建议方案**：
- [ ] 策略列表新增"执行历史"按钮
- [ ] 弹窗或侧栏展示 PolicyExecution 记录（时间、触发源、决策、原因）

**涉及文件**：
- `packages/frontend/src/views/controls/rules.vue`
- `packages/frontend/src/api/policy.ts`
- `packages/frontend/src/types/policy.ts`

---

## 五、已完成项（本次迭代）

| 项目 | 状态 |
|------|:--:|
| `RequiredDurationSec` 后端实现 | ✅ |
| `Aggregation` 字段加入前端表单 | ✅ |
| 前端表单根据策略类型条件渲染 | ✅ |
| 列表页移除硬编码 THRESHOLD 过滤 | ✅ |
| 策略类型下拉 DURATION 禁用 | ✅ |
| 定时策略条件开关 | ✅ |
| `effective_from` / `effective_to` 加入前端表单 | ✅ |
| 对话框/页面标题动态化 | ✅ |
| 设备种子数据 `seed_devices.sql` | ✅ |
| 采集器 MQTT 模拟器 | ✅ |

---

> 最后更新：2026-05-08
