# 控制策略编辑后需重新发布 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `THRESHOLD` 与 `SCHEDULE` 策略在当前前端编辑页保存后统一回到未发布状态，只有重新发布后才参与自动执行，并让前端明确区分已发布与未发布状态。

**Architecture:** 后端继续复用 `published_at / published_by` 作为唯一发布状态来源；在策略主表更新链路中清空发布状态，并让 `THRESHOLD` 自动触发补齐 `published_at IS NOT NULL` 门槛。策略调度器作为单实例依赖注入到 handler，在重新发布 `THRESHOLD` 后清空对应的冷却与条件累计内存态；前端列表新增发布状态展示，并按状态调整发布按钮与提示文案。

**Tech Stack:** Go 1.24, Gin, GORM, Vue 3, TypeScript, Element Plus

---

## 文件范围

**Create:**
- `packages/backend/internal/policy/policy_publish_test.go`

**Modify:**
- `packages/backend/internal/policy/execution_handler.go`
- `packages/backend/internal/policy/policy_handler_test.go`
- `packages/backend/internal/policy/policy_handler.go`
- `packages/backend/internal/policy/routes.go`
- `packages/backend/internal/policy/scheduler.go`
- `packages/backend/internal/policy/scheduler_test.go`
- `packages/frontend/src/views/controls/rules.vue`
- `packages/frontend/src/api/policy.ts`
- `packages/frontend/docs/HANDOFF.md`
- `packages/backend/docs/HANDOFF.md`
- `shared/docs/API_SPEC.md`

**Verify:**
- `go test ./internal/policy/...`
- `go test ./...`
- `npm run type-check`
- `npm run build`

---

## Chunk 1: 后端发布语义统一

### Task 1: 先写“编辑后回到未发布”的失败测试

**Files:**
- Create: `packages/backend/internal/policy/policy_publish_test.go`

- [ ] **Step 1: 写已发布 `THRESHOLD` 编辑后变未发布的失败测试**

```go
func TestUpdatePublishedThresholdPolicyClearsPublishState(t *testing.T) {
    // 准备 published_at / published_by 均有值的 THRESHOLD 策略
    // 调用 UpdatePolicy
    // 断言 published_at == nil && published_by == nil
}
```

- [ ] **Step 2: 写已发布 `SCHEDULE` 编辑后变未发布的失败测试**

```go
func TestUpdatePublishedSchedulePolicyClearsPublishState(t *testing.T) {
    // 准备已发布 SCHEDULE 策略
    // 更新 name 或 priority
    // 断言发布状态被清空
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/policy/... -run PublishState -v`

Expected: FAIL，原因是当前更新接口不会清空 `published_at / published_by`

- [ ] **Step 4: 提交**

```bash
git add packages/backend/internal/policy/policy_publish_test.go
git commit -m "test: cover policy publish state reset on update"
```

### Task 2: 在策略更新链路中清空发布状态

**Files:**
- Modify: `packages/backend/internal/policy/policy_handler.go`

- [ ] **Step 1: 定位 `UpdatePolicy()` 的更新字段集合**

确认当前 `updates` map 中是否已包含:

```go
"published_at"
"published_by"
```

- [ ] **Step 2: 在 `UpdatePolicy()` 成功更新时统一清空发布状态**

最小实现:

```go
updates["published_at"] = nil
updates["published_by"] = nil
```

要求:
- 不区分 `THRESHOLD` / `SCHEDULE`
- 只要当前编辑页触发 `PUT /api/policies/:id`，就回到未发布

- [ ] **Step 3: 保持 `scheduleConfigChanged()` 相关逻辑不回退**

确保已有 `last_scheduled_for` 处理逻辑仍可继续工作，不因本次变更被覆盖。

- [ ] **Step 4: 运行刚才的测试**

Run: `go test ./internal/policy/... -run PublishState -v`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add packages/backend/internal/policy/policy_handler.go packages/backend/internal/policy/policy_publish_test.go
git commit -m "feat: reset publish state after policy update"
```

### Task 3: 让 `THRESHOLD` 自动触发也要求已发布

**Files:**
- Modify: `packages/backend/internal/policy/scheduler.go`
- Modify: `packages/backend/internal/policy/scheduler_test.go`

- [ ] **Step 1: 先写未发布 `THRESHOLD` 不自动触发的失败测试**

```go
func TestThresholdPolicyRequiresPublishedState(t *testing.T) {
    // 准备 enabled=true 但 published_at=nil 的 THRESHOLD 策略
    // 触发 telemetry 评估
    // 断言没有执行记录或没有 control_commands
}
```

- [ ] **Step 2: 写已发布 `THRESHOLD` 仍可自动触发的保护测试**

```go
func TestPublishedThresholdPolicyStillAutoExecutes(t *testing.T) {
    // 准备 published_at != nil 的 THRESHOLD 策略
    // 触发 telemetry 评估
    // 断言命令和执行记录正常生成
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/policy/... -run "TestThresholdPolicyRequiresPublishedState|TestPublishedThresholdPolicyStillAutoExecutes" -v`

Expected: FAIL，未发布策略当前仍会参与阈值自动触发

- [ ] **Step 4: 在 `evaluateThresholdPolicies()` 查询中补 `published_at IS NOT NULL`**

最小代码方向:

```go
.Where("published_at IS NOT NULL")
```

- [ ] **Step 5: 重跑阈值发布状态测试**

Run: `go test ./internal/policy/... -run Threshold.*Published -v`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add packages/backend/internal/policy/scheduler.go packages/backend/internal/policy/scheduler_test.go
git commit -m "feat: require published threshold policies for auto execution"
```

---

## Chunk 2: 发布权限与运行时缓存重置

### Task 4: 先写重新发布后清空阈值运行时缓存的失败测试

**Files:**
- Modify: `packages/backend/internal/policy/scheduler_test.go`
- Create: `packages/backend/internal/policy/policy_publish_test.go`

- [ ] **Step 1: 写冷却时间被清空的失败测试**

```go
func TestPublishPolicyResetsThresholdCooldown(t *testing.T) {
    // 让某个 THRESHOLD 策略先进入 cooldown
    // 调用发布动作
    // 再次触发评估时不应被旧 cooldown 拦截
}
```

- [ ] **Step 2: 写条件持续命中累计被清空的失败测试**

```go
func TestPublishPolicyResetsThresholdConditionState(t *testing.T) {
    // 让 RequiredDurationSec 累计到中间状态
    // 重新发布后应从头累计
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/policy/... -run "PublishPolicyResets|ThresholdConditionState" -v`

Expected: FAIL，当前发布接口不会触发调度器内存态清理

### Task 5: 把 `Scheduler` 变成可注入依赖

**Files:**
- Modify: `packages/backend/internal/policy/execution_handler.go`
- Modify: `packages/backend/internal/policy/policy_handler_test.go`
- Modify: `packages/backend/internal/policy/routes.go`
- Modify: `packages/backend/internal/policy/scheduler.go`

- [ ] **Step 1: 为 `Handler` 增加调度器依赖字段**

建议最小结构:

```go
type Handler struct {
    db        *gorm.DB
    scheduler *Scheduler
}
```

- [ ] **Step 2: 调整 `NewHandler(...)` 构造函数**

改为接收调度器实例:

```go
func NewHandler(db *gorm.DB, scheduler *Scheduler) *Handler
```

- [ ] **Step 3: 同步修正现有测试调用点**

至少更新当前直接调用 `NewHandler(db)` 的测试文件，改为显式传 `nil` 或测试调度器实例，避免仅因构造签名变化导致编译失败。

- [ ] **Step 4: 在 `routes.go` 中创建单个 `Scheduler` 实例并复用**

执行顺序:
1. `scheduler := NewScheduler(...)`
2. `scheduler.Start()`
3. `h := NewHandler(deps.MySQL, scheduler)`

- [ ] **Step 5: 在 `scheduler.go` 暴露策略级运行时重置方法**

新增最小方法:

```go
func (s *Scheduler) ResetPolicyRuntime(policyID uint64)
```

方法内负责:
- 清理 `cooldowns`
- 清理 `condStates` 中属于该策略的键

- [ ] **Step 6: 运行编译检查**

Run: `go test ./internal/policy/...`

Expected: 编译通过，发布缓存测试可能仍失败

### Task 6: 发布接口补权限与缓存清理

**Files:**
- Modify: `packages/backend/internal/policy/policy_handler.go`
- Modify: `packages/backend/internal/policy/routes.go`
- Modify: `packages/backend/internal/policy/policy_publish_test.go`

- [ ] **Step 1: 把发布路由权限从 `ADMIN` 扩到 `ADMIN + OPERATOR`**

目标位置:

```go
pol.POST("/:id/publish", auth.AuthRequired(...))
```

这一步属于已确认产品规则，不是实现时临时扩权:
- `OPERATOR` 当前已可编辑策略
- 编辑页保存后策略会回到未发布
- 若仍限制 `ADMIN` 才能发布，会造成编辑与生效链路断开

- [ ] **Step 2: 在 `PublishPolicy()` 成功写入后调用调度器重置**

最小方向:

```go
if h.scheduler != nil {
    h.scheduler.ResetPolicyRuntime(id)
}
```

- [ ] **Step 3: 增加 `OPERATOR` 可发布的 handler 测试或路由级用例**

如果当前模块没有现成鉴权测试，可至少在计划执行时补一个最小路由用例，验证 `OPERATOR` 不再被拒绝。

- [ ] **Step 4: 为 `publish` 动作补审计计划**

优先使用与 `create/update/delete` 一致的轻量包装方式，至少让 `publish` 被记录为关键运维动作；若实现成本很低，可顺手让 `archive` 也保持一致。

- [ ] **Step 5: 重跑缓存重置相关测试**

Run: `go test ./internal/policy/... -run "PublishPolicyResets|ThresholdConditionState|PublishState" -v`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add packages/backend/internal/policy/execution_handler.go packages/backend/internal/policy/policy_handler_test.go packages/backend/internal/policy/routes.go packages/backend/internal/policy/scheduler.go packages/backend/internal/policy/policy_handler.go packages/backend/internal/policy/policy_publish_test.go packages/backend/internal/policy/scheduler_test.go
git commit -m "feat: republish policies after edit and reset threshold runtime"
```

---

## Chunk 3: 前端状态区分、文档与整体验证

### Task 7: 先写前端状态展示改造清单

**Files:**
- Modify: `packages/frontend/src/views/controls/rules.vue`
- Modify: `packages/frontend/src/api/policy.ts`

- [ ] **Step 1: 列出页面要新增/修改的可视化元素**

至少包括:

```text
发布状态列
发布按钮状态区分
创建成功提示文案
编辑成功提示文案
可选发布时间 tooltip/文案
```

- [ ] **Step 2: 确认前端已经拥有 `published_at` 字段**

检查 `ControlPolicy` 类型与列表接口数据映射，不新增多余请求。

### Task 8: 实现前端发布状态区分

**Files:**
- Modify: `packages/frontend/src/views/controls/rules.vue`

- [ ] **Step 1: 为列表新增“发布状态”列**

渲染规则:

```ts
published_at ? '已发布' : '未发布'
```

- [ ] **Step 2: 增加辅助函数**

建议最小函数:

```ts
function isPolicyPublished(policy: ControlPolicy) {
  return Boolean(policy.published_at)
}
```

- [ ] **Step 3: 调整发布按钮**

要求:
- 已发布: 显示 `已发布` 且禁用，或用状态标签替代
- 未发布: 显示可点击 `发布`

- [ ] **Step 4: 调整成功提示文案**

改成:

```ts
'策略创建成功，请发布后生效'
'策略更新成功，当前为未发布状态，请重新发布后生效'
```

- [ ] **Step 5: 确保编辑后刷新列表读取最新发布状态**

保留或强化:

```ts
await fetchData()
```

避免依赖 `PUT /api/policies/:id` 的空响应体推断状态。

### Task 9: 同步前端 API 注释与文档

**Files:**
- Modify: `packages/frontend/src/api/policy.ts`
- Modify: `packages/frontend/docs/HANDOFF.md`
- Modify: `packages/backend/docs/HANDOFF.md`
- Modify: `shared/docs/API_SPEC.md`

- [ ] **Step 1: 在前端 API/注释中写明 `published_at` 是状态来源**

如果不想新增注释，可至少保持类型与调用语义一致。

- [ ] **Step 2: 更新前端交接文档**

记录:
- 策略列表新增发布状态展示
- 编辑页保存后回到未发布
- 已发布按钮不可重复点击

- [ ] **Step 3: 更新后端交接文档**

记录:
- `THRESHOLD` 自动触发也要求已发布
- `UpdatePolicy()` 清空发布状态
- 发布时重置阈值运行时缓存
- `OPERATOR` 可发布

- [ ] **Step 4: 更新共享 API 文档**

明确:
- 自动执行要求已发布
- 编辑页保存后需重新发布
- 发布状态由 `published_at` 判定

- [ ] **Step 5: 提交**

```bash
git add packages/frontend/src/views/controls/rules.vue packages/frontend/src/api/policy.ts packages/frontend/docs/HANDOFF.md packages/backend/docs/HANDOFF.md shared/docs/API_SPEC.md
git commit -m "feat: show policy publish status in control rules"
```

### Task 10: 整体验证

**Files:**
- Verify only

- [ ] **Step 1: 跑 policy 模块测试**

Run: `go test ./internal/policy/...`

Expected: PASS

- [ ] **Step 2: 跑后端全量测试**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 3: 跑前端类型检查**

Run: `npm run type-check`

Expected: PASS

- [ ] **Step 4: 跑前端构建**

Run: `npm run build`

Expected: PASS

- [ ] **Step 5: 手动验证关键路径**

至少验证:
1. 新建策略后列表显示未发布
2. 发布后列表显示已发布
3. 编辑页保存后列表重新变未发布
4. 未发布 `THRESHOLD` 不自动触发
5. 未发布 `SCHEDULE` 不自动执行
6. 重新发布 `THRESHOLD` 后不会沿用旧 cooldown 或旧 `RequiredDurationSec` 累计状态

- [ ] **Step 6: 最终提交**

```bash
git add -A
git commit -m "feat: require republish after editing control policies"
```

---

## 计划备注

- 当前设计明确以“前端现有编辑页保存路径”为边界，不扩展到独立条件/目标子资源接口的统一失效逻辑。
- 调度器运行时缓存清理仅保证当前单实例进程内一致性，不覆盖未来多实例部署的跨进程同步问题。
- 若在实现中发现 `policy_handler.go` 或 `rules.vue` 再次膨胀，可接受小范围抽 helper，但不要做无关重构。
