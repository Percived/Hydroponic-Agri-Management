# Minimal Audit MVP Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为系统补齐最小可用审计链路，至少覆盖登录/登出、用户、设备、策略、告警，以及每次执行器通道命令，并让前端审计页能直接查看结果。

**Architecture:** 采用“业务成功点显式写审计”的最小改动方案：在后端成功 handler 中调用统一的 `audit` 写入 helper，命令模块单独补充执行器通道元数据与最终状态；审计查询接口联表返回用户名并把详情格式化为可读文本，前端同步移除当前不存在的 IP 字段并对齐类型。

**Tech Stack:** Go 1.24, Gin, GORM, MySQL, Vue 3, TypeScript, Element Plus

---

## 文件范围

**Create:**
- `docs/superpowers/plans/2026-05-12-minimal-audit-mvp.md`
- `packages/backend/internal/audit/logger_test.go`

**Modify:**
- `packages/backend/internal/audit/logger.go`
- `packages/backend/internal/audit/handler.go`
- `packages/backend/internal/auth/handler.go`
- `packages/backend/internal/device/handler.go`
- `packages/backend/internal/policy/policy_handler.go`
- `packages/backend/internal/alert/handler.go`
- `packages/backend/internal/command/handler.go`
- `packages/backend/internal/command/handler_send_test.go`
- `packages/frontend/src/types/audit.ts`
- `packages/frontend/src/views/audit-logs/index.vue`
- `packages/frontend/docs/HANDOFF.md`
- `packages/backend/docs/HANDOFF.md`
- `shared/docs/API_SPEC.md`

**Verify:**
- `go test ./internal/audit ./internal/command`
- `go test ./...`
- `npm run type-check`

---

## Chunk 1: 后端审计写入

### Task 1: 扩展审计 helper 与测试

**Files:**
- Create: `packages/backend/internal/audit/logger_test.go`
- Modify: `packages/backend/internal/audit/logger.go`

- [ ] 先写失败测试，覆盖基础写入字段与扩展 detail/request_id/before/after。
- [ ] 运行 `go test ./internal/audit`，确认先失败。
- [ ] 实现统一 `audit.Write`/`audit.WriteEntry` 最小能力。
- [ ] 再跑 `go test ./internal/audit`，确认转绿。

### Task 2: 接入主链路 handler

**Files:**
- Modify: `packages/backend/internal/auth/handler.go`
- Modify: `packages/backend/internal/device/handler.go`
- Modify: `packages/backend/internal/policy/policy_handler.go`
- Modify: `packages/backend/internal/alert/handler.go`
- Modify: `packages/backend/internal/command/handler.go`

- [ ] 登录/登出成功后写 `LOGIN` / `LOGOUT`。
- [ ] 用户成功创建/更新后写 `CREATE_USER` / `UPDATE_USER`。
- [ ] 设备与执行器/传感器通道成功增删改后写 `CREATE_DEVICE` / `UPDATE_DEVICE` / `DELETE_DEVICE`。
- [ ] 策略成功增删改后写 `CREATE_RULE` / `UPDATE_RULE` / `DELETE_RULE`。
- [ ] 告警状态成功变更后写 `UPDATE_ALERT`。
- [ ] 命令成功发送路径写 `CONTROL_CMD`，记录 `actuator_channel_id`、`device_code`、`channel_code`、`command_type`、`payload`、最终状态。

## Chunk 2: 审计查询与前端显示

### Task 3: 审计列表接口补齐字段

**Files:**
- Modify: `packages/backend/internal/audit/handler.go`

- [ ] 联表 `users` 返回 `username`。
- [ ] 将 `detail` 统一序列化为前端可读字符串。
- [ ] 保持现有分页与 `action/start_time/end_time` 过滤可用。

### Task 4: 前端页面对齐返回结构

**Files:**
- Modify: `packages/frontend/src/types/audit.ts`
- Modify: `packages/frontend/src/views/audit-logs/index.vue`

- [ ] 删除当前并不存在的 `ip_address` 依赖。
- [ ] 对齐 `username/detail` 的真实返回结构。
- [ ] 保持现有筛选和列表布局基本不变。

## Chunk 3: 验证与文档

### Task 5: 回归验证与文档更新

**Files:**
- Modify: `packages/frontend/docs/HANDOFF.md`
- Modify: `packages/backend/docs/HANDOFF.md`
- Modify: `shared/docs/API_SPEC.md`

- [ ] 跑后端定向测试和全量测试。
- [ ] 跑前端类型检查。
- [ ] 记录审计最小可用方案、已覆盖动作和命令审计细节。
