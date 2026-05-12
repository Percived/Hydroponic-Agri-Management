# SCHEDULE 策略真正“到点执行”设计

日期: 2026-05-12
范围: `packages/backend` + `packages/frontend` + `shared/docs`
状态: 已评审前设计草案

## 1. 背景

当前 `SCHEDULE` 策略存在两个核心问题:

1. 只有 `effective_from` / `effective_to` 生效窗口，没有真正的“执行时刻”表达。
2. 前端允许“无条件仅定时执行”，但后端调度器要求 `Conditions` 非空，导致部分定时策略被静默跳过。

这会使用户将 `SCHEDULE` 理解为“到点执行”，但系统实际行为更接近“在生效窗口内周期评估一次”，语义不一致且可观测性不足。

## 2. 目标

本次改造为 `SCHEDULE` 增加真正的“到点执行”能力，首期支持以下三种计划模式:

- `ONCE`: 单次执行
- `DAILY`: 每日固定时刻执行
- `WEEKLY`: 每周指定星期和固定时刻执行

同时满足以下要求:

- `SCHEDULE` 允许没有条件，命中计划时可直接执行目标动作
- `effective_from` / `effective_to` 仅作为生效窗口，不再表示执行时间
- 调度器对同一计划点幂等，重复扫描不会重复执行
- `policy_executions` 对 `SCHEDULE` 始终写执行记录，不能再静默跳过
- 调度器只处理已发布策略，使“发布”真正成为生效门槛

## 3. 非目标

本次不做以下内容:

- 不支持 cron 表达式
- 不引入预生成任务实例表
- 不改造 `THRESHOLD` 策略执行模型
- 不支持复杂日历规则，如“每月第一个周一”

## 4. 数据模型设计

### 4.1 `control_policies` 新增字段

为 `SCHEDULE` 新增以下字段:

- `schedule_mode` `VARCHAR(16)`:
  - 可选值: `ONCE | DAILY | WEEKLY`
  - `THRESHOLD` 为 `NULL`
- `run_once_at` `DATETIME(3)`:
  - 仅 `ONCE` 使用
- `time_of_day` `TIME`:
  - `DAILY` / `WEEKLY` 使用
- `weekdays_mask` `TINYINT UNSIGNED`:
  - 仅 `WEEKLY` 使用
  - 用 7 bit 表示周一到周日
- `timezone` `VARCHAR(64)`:
  - 默认 `Asia/Shanghai`
- `last_scheduled_for` `DATETIME(3)`:
  - 最近一次已处理的计划时刻，用于幂等去重

### 4.2 语义约定

- `effective_from` / `effective_to`:
  - 表示计划是否生效
  - 不再表示要执行的具体时刻
- `published_at`:
  - 作为进入调度器扫描范围的门槛
- `last_scheduled_for`:
  - 表示最近一次已成功处理或已明确占坑处理的计划点
  - 用于避免调度器重复扫描时重复执行同一时刻

### 4.3 历史数据兼容

历史 `SCHEDULE` 策略不会自动推断成 `ONCE` / `DAILY` / `WEEKLY`，避免错误触发。

兼容规则:

- 历史记录新增字段默认 `NULL`
- 后端返回给前端时允许 `schedule_mode = null`
- 前端展示“计划未配置”
- 调度器扫描到这类记录时写 `SKIPPED`，原因为 `schedule_not_configured`

## 5. 后端接口设计

### 5.1 DTO 扩展

扩展以下 DTO:

- `CreatePolicyRequest`
- `UpdatePolicyRequest`
- `ControlPolicyResponse`

新增字段:

- `schedule_mode`
- `run_once_at`
- `time_of_day`
- `weekdays_mask`
- `timezone`
- `last_scheduled_for` 仅响应体返回

### 5.2 校验规则

当 `policy_type = SCHEDULE` 时:

- `ONCE`
  - 必填 `run_once_at`
  - 禁止 `time_of_day`
  - 禁止 `weekdays_mask`
- `DAILY`
  - 必填 `time_of_day`
  - 禁止 `run_once_at`
  - 禁止 `weekdays_mask`
- `WEEKLY`
  - 必填 `time_of_day`
  - 必填 `weekdays_mask`
  - 禁止 `run_once_at`
- `timezone`
  - 为空时后端写入 `Asia/Shanghai`

当 `policy_type = THRESHOLD` 时:

- 禁止以上所有调度字段
- 保持现有条件校验逻辑

### 5.3 发布语义

调度器只扫描满足以下条件的策略:

- `enabled = true`
- `published_at IS NOT NULL`
- `policy_type = SCHEDULE`

这样前端“保存”和“发布”行为语义一致:

- 保存: 编辑策略定义
- 发布: 让策略进入自动调度

## 6. 调度器设计

### 6.1 总体思路

保留现有定时扫描器，但拆分为两套明确逻辑:

- `THRESHOLD`: 继续沿用事件驱动 + 定时补查
- `SCHEDULE`: 改为严格的计划时刻驱动

### 6.2 `SCHEDULE` 扫描条件

扫描器每 30 秒拉取:

- `enabled = true`
- `policy_type = SCHEDULE`
- `published_at IS NOT NULL`
- `schedule_mode IS NOT NULL`

### 6.3 计划点计算

对每条 `SCHEDULE`:

1. 读取 `timezone`
2. 将当前 UTC 时间转换到本地时区时间
3. 计算“本次扫描窗口内是否存在一个应执行的 `scheduled_for`”
4. 如果存在，进一步判断:
   - 是否落在 `effective_from` / `effective_to` 窗口
   - 是否已经被 `last_scheduled_for` 处理过

### 6.4 各模式算法

#### ONCE

- `scheduled_for = run_once_at`
- 当 `last_scan < run_once_at <= now` 时命中
- 若 `last_scheduled_for >= run_once_at`，则判定已执行过

#### DAILY

- 根据 `timezone` 取当天日期 + `time_of_day`
- 若该时刻落在扫描窗口内，则命中

#### WEEKLY

- 在 `DAILY` 基础上增加星期过滤
- 仅当当天星期命中 `weekdays_mask` 时才生成 `scheduled_for`

### 6.5 去重机制

`SCHEDULE` 的正确性不再依赖原 `cooldown`。

改为:

- 主去重键: `policy_id + scheduled_for`
- 状态持久化字段: `last_scheduled_for`

`cooldown` 可保留作为防抖，但只作为辅助保护，不作为业务正确性的核心机制。

### 6.6 执行记录

对 `SCHEDULE`，无论是否真正下发命令，都必须写 `policy_executions`:

- `EXECUTED`
  - `schedule_due`
- `SKIPPED`
  - `schedule_not_configured`
  - `outside_effective_window`
  - `already_executed_for_slot`
  - `no_targets`
- `FAILED`
  - `target_execution_failed:*`

可选增强:

- 将命中的 `scheduled_for` 文本写入 `decision_reason` 或扩展专用字段

## 7. 条件与目标行为

### 7.1 `SCHEDULE`

- 允许没有条件
- 命中计划点后直接执行 `Targets`
- 若未来需要“到点 + 条件”组合，可保留条件列表作为可选增强，但首期不把条件作为必须项

### 7.2 `THRESHOLD`

- 保持现状
- 仍要求条件和目标同时存在

## 8. 前端设计

### 8.1 策略表单

在策略管理页中:

- 选择 `SCHEDULE` 后隐藏“无条件（仅定时执行）”开关
- 改为显示结构化计划表单

字段设计:

- `计划类型`
  - 单次执行
  - 每日执行
  - 每周执行
- `执行时间`
  - 单次: 日期时间选择器
  - 每日: 时间选择器
  - 每周: 星期多选 + 时间选择器
- `时区`
  - 首期默认 `Asia/Shanghai`
  - 可先隐藏为固定值，也可展示为只读
- `生效窗口`
  - 对应原 `effective_from` / `effective_to`
  - 明确改文案，避免误解为执行时刻

### 8.2 列表展示

策略列表新增“计划描述”列:

- `ONCE`: `2026-05-13 09:00:00 单次`
- `DAILY`: `每日 08:00:00`
- `WEEKLY`: `每周一/三/五 18:30:00`

历史未配置计划显示:

- `计划未配置`

### 8.3 类型定义

更新:

- `src/types/policy.ts`
- `src/api/policy.ts`
- `src/views/controls/rules.vue`

前端 `SCHEDULE` payload 统一走结构化字段，不再通过“空条件”来表达“仅定时执行”。

## 9. 数据迁移策略

### 9.1 Schema

更新:

- `packages/backend/migrations/merged/all.up.sql`
- 如仓库已有增量 merged 迁移惯例，则补一份对应 `v2.4_*` merged 迁移

### 9.2 兼容处理

不对旧 `SCHEDULE` 自动补调度规则，理由:

- 老数据缺少可靠语义来源
- 自动推断可能造成误触发

历史策略由用户在前端重新编辑并发布后生效。

## 10. 测试设计

### 10.1 后端

新增调度器测试，覆盖:

- `ONCE` 命中一次且仅一次
- `DAILY` 在指定时刻命中
- `WEEKLY` 仅在指定星期命中
- 未发布不执行
- 超出生效窗口不执行
- 历史未配置计划写 `SKIPPED`
- 同一 `scheduled_for` 在多次扫描中不重复执行

### 10.2 前端

覆盖:

- 三种 `SCHEDULE` 模式的表单显隐
- payload 序列化正确
- 历史未配置计划回显提示正确

### 10.3 联调验证

验证路径:

1. 新建 `ONCE` 策略
2. 保存
3. 发布
4. 等待命中
5. 检查 `policy_executions`
6. 检查 `control_commands`
7. 检查执行器通道状态变化

## 11. 变更文件范围

预计涉及:

- 后端
  - `packages/backend/internal/policy/model.go`
  - `packages/backend/internal/policy/dto.go`
  - `packages/backend/internal/policy/policy_handler.go`
  - `packages/backend/internal/policy/scheduler.go`
  - `packages/backend/migrations/merged/all.up.sql`
  - 可能新增调度器测试文件
- 前端
  - `packages/frontend/src/types/policy.ts`
  - `packages/frontend/src/api/policy.ts`
  - `packages/frontend/src/views/controls/rules.vue`
- 文档
  - `packages/backend/docs/HANDOFF.md`
  - `packages/backend/docs/PROJECT_STATUS.md`
  - `packages/frontend/docs/HANDOFF.md`
  - `shared/docs/API_SPEC.md`

## 12. 风险与控制

- 风险: 时区与本地时间计算出错
  - 控制: 统一用 `timezone` + 单测覆盖边界
- 风险: 扫描窗口导致重复执行
  - 控制: 以 `last_scheduled_for` 做幂等
- 风险: 历史策略行为变化
  - 控制: 未配置计划统一视为 `SKIPPED`，不自动触发
- 风险: 前端交互仍误导用户
  - 控制: 明确区分“执行时间”和“生效窗口”

## 13. 推荐实施顺序

1. 迁移与模型字段补齐
2. DTO 与校验补齐
3. 调度器改造与后端测试
4. 前端类型与表单改造
5. 文档同步
6. 联调验证
