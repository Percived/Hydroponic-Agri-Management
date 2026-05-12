# SCHEDULE 真正到点执行 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `SCHEDULE` 策略补齐真正的“到点执行”能力，支持 `ONCE / DAILY / WEEKLY`，并同步完成后端调度、前端表单、表结构迁移、测试与文档更新。

**Architecture:** 在 `control_policies` 增加结构化调度字段，将 `SCHEDULE` 从“生效窗口扫描”改为“计划时刻命中”模型；后端调度器用 `last_scheduled_for` 做幂等去重，前端使用表单化配置单次/每日/每周执行计划，保留 `effective_from / effective_to` 作为生效窗口。`THRESHOLD` 行为维持不变，但调度器统一要求策略已发布后才参与自动执行。

**Tech Stack:** Go 1.24, Gin, GORM, MySQL, Vue 3, TypeScript, Element Plus

---

## 文件范围

**Create:**
- `packages/backend/internal/policy/scheduler_test.go`
- `docs/superpowers/plans/2026-05-12-schedule-policy-true-time-trigger.md`

**Modify:**
- `packages/backend/migrations/merged/all.up.sql`
- `packages/backend/internal/policy/model.go`
- `packages/backend/internal/policy/dto.go`
- `packages/backend/internal/policy/policy_handler.go`
- `packages/backend/internal/policy/scheduler.go`
- `packages/frontend/src/types/policy.ts`
- `packages/frontend/src/views/controls/rules.vue`
- `packages/backend/docs/HANDOFF.md`
- `packages/backend/docs/PROJECT_STATUS.md`
- `packages/frontend/docs/HANDOFF.md`
- `shared/docs/API_SPEC.md`

**Verify:**
- `go test ./internal/policy/...`
- `go test ./...`
- `npm run type-check`
- 需要时补 `npm run build`

---

## Chunk 1: 后端 Schema 与 DTO

### Task 1: 为 `control_policies` 增加调度字段

**Files:**
- Modify: `packages/backend/migrations/merged/all.up.sql`

- [ ] **Step 1: 先写迁移影响检查清单**

记录要新增的列与默认值:

```text
schedule_mode VARCHAR(16) NULL
run_once_at DATETIME(3) NULL
time_of_day TIME NULL
weekdays_mask TINYINT UNSIGNED NULL
timezone VARCHAR(64) NULL DEFAULT 'Asia/Shanghai'
last_scheduled_for DATETIME(3) NULL
```

- [ ] **Step 2: 更新 `control_policies` 表定义**

在 `effective_to` 后加入新列，并保留已有 `published_by / published_at` 字段顺序的可读性。

- [ ] **Step 3: 为调度扫描补索引**

新增或扩展索引，至少覆盖以下过滤组合:

```sql
KEY `idx_policies_schedule_scan` (`policy_type`, `enabled`, `published_at`, `schedule_mode`)
```

- [ ] **Step 4: 复查 SQL 初始化兼容性**

检查 `all.up.sql` 仍可用于全量初始化，不对历史 `SCHEDULE` 自动推断计划。

- [ ] **Step 5: 提交**

```bash
git add packages/backend/migrations/merged/all.up.sql
git commit -m "feat: add schedule fields to control policies"
```

### Task 2: 扩展后端模型与 DTO

**Files:**
- Modify: `packages/backend/internal/policy/model.go`
- Modify: `packages/backend/internal/policy/dto.go`

- [ ] **Step 1: 先写字段映射测试草稿**

列出 `ControlPolicy` 需要新增的字段:

```go
ScheduleMode      *string
RunOnceAt         *time.Time
TimeOfDay         *string
WeekdaysMask      *uint8
Timezone          string
LastScheduledFor  *time.Time
```

- [ ] **Step 2: 更新 `ControlPolicy` 模型**

在 `model.go` 中加入 GORM tag，保持 `THRESHOLD` 可为空、`timezone` 默认 `Asia/Shanghai`。

- [ ] **Step 3: 更新请求/响应 DTO**

在以下 DTO 中新增调度字段:

```go
CreatePolicyRequest
UpdatePolicyRequest
ControlPolicyResponse
```

- [ ] **Step 4: 增加基础校验注释与枚举说明**

为 `schedule_mode` 增加 `oneof=ONCE DAILY WEEKLY` 约束。

- [ ] **Step 5: 编译检查**

Run: `go test ./internal/policy/...`

Expected: 编译可通过，测试此时可能仍失败

- [ ] **Step 6: 提交**

```bash
git add packages/backend/internal/policy/model.go packages/backend/internal/policy/dto.go
git commit -m "feat: add schedule dto and model fields"
```

---

## Chunk 2: 后端校验与调度器

### Task 3: 先写调度器失败测试

**Files:**
- Create: `packages/backend/internal/policy/scheduler_test.go`

- [ ] **Step 1: 写 `ONCE` 只执行一次的失败测试**

```go
func TestScheduleOnceExecutesOnlyOneSlot(t *testing.T) {
    // 创建已发布 ONCE 策略，命中一次后再次扫描不应重复执行
}
```

- [ ] **Step 2: 写 `DAILY` 命中测试**

```go
func TestScheduleDailyExecutesWhenSlotDue(t *testing.T) {
    // 当前时间跨过指定 time_of_day 时生成执行记录
}
```

- [ ] **Step 3: 写 `WEEKLY` 星期过滤测试**

```go
func TestScheduleWeeklySkipsWhenWeekdayNotMatched(t *testing.T) {
    // 非命中星期应写 SKIPPED 或不执行命令
}
```

- [ ] **Step 4: 写历史未配置计划测试**

```go
func TestSchedulePolicyWithoutPlanWritesSkippedExecution(t *testing.T) {
    // schedule_mode 为空时写 schedule_not_configured
}
```

- [ ] **Step 5: 运行测试确认失败**

Run: `go test ./internal/policy/... -run Schedule -v`

Expected: FAIL，原因是字段/调度逻辑尚未实现

### Task 4: 实现策略请求校验

**Files:**
- Modify: `packages/backend/internal/policy/policy_handler.go`

- [ ] **Step 1: 抽取 `validatePolicyScheduleFields` 辅助函数**

覆盖以下规则:

```go
if policyType == "SCHEDULE" { ... }
if policyType == "THRESHOLD" { reject schedule fields }
```

- [ ] **Step 2: 在 `CreatePolicy` 中接入校验与默认时区**

空 `timezone` 自动填 `Asia/Shanghai`。

- [ ] **Step 3: 在 `UpdatePolicy` 中接入校验**

更新时需要“现有策略 + patch 请求”合并后再校验，避免局部更新绕过约束。

- [ ] **Step 4: 让发布语义成为自动调度门槛**

不改现有 `publish` 接口，但后续调度查询只扫描 `published_at IS NOT NULL`。

- [ ] **Step 5: 运行后端测试**

Run: `go test ./internal/policy/...`

Expected: 仍有调度类测试失败，但校验相关编译通过

### Task 5: 改造 `SCHEDULE` 调度器

**Files:**
- Modify: `packages/backend/internal/policy/scheduler.go`

- [ ] **Step 1: 抽取计划时刻计算辅助方法**

新增小函数，保持职责清晰:

```go
func (s *Scheduler) findDueScheduledFor(...)
func buildDailySlot(...)
func weekdayMatched(mask uint8, weekday time.Weekday) bool
```

- [ ] **Step 2: 改造 `evaluateScheduledPolicies` 查询**

加入过滤:

```go
published_at IS NOT NULL
schedule_mode IS NOT NULL
```

- [ ] **Step 3: 为 `SCHEDULE` 引入独立执行入口**

避免复用 `evaluateAndExecute` 的“条件必填”假设，新增类似:

```go
func (s *Scheduler) evaluateScheduledPolicy(p ControlPolicy)
```

- [ ] **Step 4: 实现幂等去重**

按 `scheduled_for` 与 `last_scheduled_for` 比较，重复扫描不重复下发命令。

- [ ] **Step 5: 强制写执行记录**

至少覆盖:

```go
schedule_due
schedule_not_configured
outside_effective_window
already_executed_for_slot
no_targets
target_execution_failed:*
```

- [ ] **Step 6: 保持 `THRESHOLD` 原逻辑可用**

不要让 `THRESHOLD` 误受新字段影响。

- [ ] **Step 7: 运行测试直到通过**

Run: `go test ./internal/policy/... -v`

Expected: PASS

- [ ] **Step 8: 运行全量后端测试**

Run: `go test ./...`

Expected: PASS，若出现仓库内既有失败，记录并隔离说明

- [ ] **Step 9: 提交**

```bash
git add packages/backend/internal/policy/policy_handler.go packages/backend/internal/policy/scheduler.go packages/backend/internal/policy/scheduler_test.go
git commit -m "feat: add true-time schedule policy execution"
```

---

## Chunk 3: 前端类型与策略页面

### Task 6: 扩展前端类型定义

**Files:**
- Modify: `packages/frontend/src/types/policy.ts`

- [ ] **Step 1: 为 `ControlPolicy` 增加调度字段**

```ts
schedule_mode?: 'ONCE' | 'DAILY' | 'WEEKLY' | null
run_once_at?: string
time_of_day?: string
weekdays_mask?: number
timezone?: string
last_scheduled_for?: string
```

- [ ] **Step 2: 为创建/更新请求增加字段**

让前端 payload 能表达结构化计划，而不是靠空条件。

- [ ] **Step 3: 运行类型检查预热**

Run: `npm run type-check`

Expected: 可能因页面未适配而失败

### Task 7: 改造策略表单与列表

**Files:**
- Modify: `packages/frontend/src/views/controls/rules.vue`

- [ ] **Step 1: 删除 `SCHEDULE` 的“无条件仅定时执行”交互**

移除或停用:

```ts
scheduleUseCondition
```

- [ ] **Step 2: 增加结构化计划表单状态**

例如:

```ts
schedule_mode
run_once_at
time_of_day
weekdays_mask
timezone
```

- [ ] **Step 3: 为三种模式新增 UI**

使用:

```vue
el-date-picker
el-time-picker
el-checkbox-group
```

- [ ] **Step 4: 更新提交 payload**

当 `policy_type === 'SCHEDULE'` 时:

- 不再用空条件表达“仅定时执行”
- 直接提交结构化调度字段
- 条件列表变为可选，不再隐式依赖 `scheduleUseCondition`

- [ ] **Step 5: 更新回显逻辑**

编辑已有 `SCHEDULE` 时正确回显:

- `ONCE`
- `DAILY`
- `WEEKLY`
- 历史未配置计划

- [ ] **Step 6: 增加列表计划描述列**

渲染示例:

```ts
每日 08:00:00
每周一/三/五 18:30:00
2026-05-13 09:00:00 单次
计划未配置
```

- [ ] **Step 7: 运行类型检查**

Run: `npm run type-check`

Expected: PASS

- [ ] **Step 8: 需要时运行构建**

Run: `npm run build`

Expected: PASS

- [ ] **Step 9: 提交**

```bash
git add packages/frontend/src/types/policy.ts packages/frontend/src/views/controls/rules.vue
git commit -m "feat: add structured schedule policy editor"
```

---

## Chunk 4: 文档与联调验证

### Task 8: 更新文档

**Files:**
- Modify: `packages/backend/docs/HANDOFF.md`
- Modify: `packages/backend/docs/PROJECT_STATUS.md`
- Modify: `packages/frontend/docs/HANDOFF.md`
- Modify: `shared/docs/API_SPEC.md`

- [ ] **Step 1: 更新后端交接文档**

说明:

- 新增调度字段
- `SCHEDULE` 真正到点执行
- 调度器只扫描已发布策略

- [ ] **Step 2: 更新项目状态**

补充:

- SCHEDULE 执行语义升级
- 历史未配置策略兼容行为

- [ ] **Step 3: 更新前端交接文档**

说明策略页从“无条件开关”改为结构化计划编辑。

- [ ] **Step 4: 更新 API 规格**

补齐 `ControlPolicy` 请求/响应字段与 `SCHEDULE` 语义说明。

- [ ] **Step 5: 提交**

```bash
git add packages/backend/docs/HANDOFF.md packages/backend/docs/PROJECT_STATUS.md packages/frontend/docs/HANDOFF.md shared/docs/API_SPEC.md
git commit -m "docs: document schedule policy true-time execution"
```

### Task 9: 做联调与收尾验证

**Files:**
- Verify only

- [ ] **Step 1: 初始化或确认数据库结构**

根据当前环境执行必要迁移，确保新增字段存在。

- [ ] **Step 2: 新建并发布 `ONCE` 策略**

选择一个可控执行器通道与接近当前时间的单次执行时间。

- [ ] **Step 3: 检查执行记录**

确认 `policy_executions` 有 `EXECUTED` 或明确 `SKIPPED` 原因。

- [ ] **Step 4: 检查命令记录**

确认 `control_commands` 生成新命令并具备合理状态。

- [ ] **Step 5: 做最终验证命令**

Run:

```bash
cd packages/backend; go test ./...
cd ..\frontend; npm run type-check
```

Expected: PASS

- [ ] **Step 6: 汇总未解决风险**

若存在仓库既有失败、环境依赖或历史数据待人工补配，明确写入交付说明。

- [ ] **Step 7: 最终提交**

```bash
git status --short
git commit -m "feat: implement true-time schedule policies"
```
