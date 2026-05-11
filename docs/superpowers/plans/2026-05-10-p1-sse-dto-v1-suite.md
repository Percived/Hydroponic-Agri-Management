# P1 SSE DTO v1 Suite Implementation Plan
> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 固化 SSE 对外契约（DTO v1 + schema_version），并补齐 device_status / commands 的实时订阅端点，确保前后端可依赖、可验收、可排障。

**Architecture:** Producer 侧统一发布 DTO v1（强类型 struct + `schema_version`），SSE handler 仅负责订阅/过滤/事件名映射/输出；对 commands 采用一个 subscribe 端点多事件类型输出。

**Tech Stack:** Go (Gin/GORM), Vue3+TS, SSE (EventSource), MQTT ingress internal events.

---

## Scope

- Backend:
  - Telemetry SSE data 增加 `schema_version:1` 并改为强类型 DTO v1
  - Device status SSE：新增 `/api/devices/subscribe` + DTO v1 + query 过滤
  - Commands SSE：新增 `/api/commands/subscribe`（推送 `command_dispatched` + `command_acked`）+ DTO v1 + query 过滤
  - SSE handler：支持 map/struct 数据提取；支持 multi-type subscribe
- Frontend:
  - `useTelemetrySSE` / `useAlertSSE` 增加 `schema_version` 运行时校验
  - 补齐 device_status / commands SSE DTO TS 类型（为后续 UI 使用做准备）
- Docs:
  - `shared/docs/API_SPEC.md`、`packages/backend/docs/HANDOFF.md`、`packages/frontend/docs/HANDOFF.md`

---

## Chunk 1: Backend — DTO v1 + SSE endpoints

### Task 1: Define SSE DTO v1 structs

**Files:**
- Create: `packages/backend/internal/platform/event/sse_dto_v1.go`
- Modify: `packages/backend/internal/platform/event/types.go` (extend CommandAckData)

- [ ] **Step 1: Add TelemetrySSEDataV1**
- [ ] **Step 2: Add DeviceStatusSSEDataV1**
- [ ] **Step 3: Add CommandDispatchedSSEDataV1**
- [ ] **Step 4: Extend CommandAckData with schema_version + acked_at**

### Task 2: Publish telemetry/device_status/commands with DTO v1

**Files:**
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`
- Modify: `packages/backend/internal/device/offline_detector.go`
- Modify: `packages/backend/internal/command/handler.go`
- Modify: `packages/backend/internal/climate/profile_scheduler.go`
- Modify: `packages/backend/internal/policy/scheduler.go`

- [ ] **Step 1: telemetry:received publishes TelemetrySSEDataV1**
- [ ] **Step 2: device:status publishes DeviceStatusSSEDataV1 (ingress + offline_detector)**
- [ ] **Step 3: command:acked publishes CommandAckData(schema_version=1, acked_at=now)**
- [ ] **Step 4: command:dispatched publishes CommandDispatchedSSEDataV1 (manual + policy + climate)**
- [ ] **Step 5: Ensure mqtt publish error branches also publish FAILED dispatched event**

### Task 3: SSE handler hardening (struct/map + multi subscribe + filtering)

**Files:**
- Modify: `packages/backend/internal/platform/event/sse_handler.go`
- Modify: `packages/backend/internal/platform/http/router.go`
- Test: `packages/backend/internal/platform/event/sse_handler_test.go`

- [ ] **Step 1: Add eventMappings for command events**
- [ ] **Step 2: Implement SSEHandlerMulti(hub, []string{...})**
- [ ] **Step 3: Add filtering for device:status + commands by device_codes**
- [ ] **Step 4: Ensure id generation uses collected_at/command_id/device_code+ts**
- [ ] **Step 5: Add tests for new endpoints mappings + filters**

### Task 4: Alerts edge-case hardening (device_code completeness)

**Files:**
- Modify: `packages/backend/internal/alert/handler.go`

- [ ] **Step 1: When creating alert manually, derive device_code from sensor_channel_id/actuator_channel_id if present**
- [ ] **Step 2: Publish alert:created still using Alert SSE v1 builder but with device_code filled**

---

## Chunk 2: Frontend — runtime guard + types

### Task 5: Add schema_version guards

**Files:**
- Modify: `packages/frontend/src/composables/useTelemetrySSE.ts`
- Modify: `packages/frontend/src/composables/useAlertSSE.ts`

- [ ] **Step 1: Reject payload missing/unknown schema_version (set error state)**
- [ ] **Step 2: Keep backward compatibility only if explicitly needed (otherwise ignore)**

### Task 6: Add TS types for new SSE DTOs

**Files:**
- Modify: `packages/frontend/src/types/device.ts`
- Modify: `packages/frontend/src/types/control.ts`
- Modify: `packages/frontend/src/types/index.ts` (export)

- [ ] **Step 1: Add DeviceStatusSSEDataV1 type**
- [ ] **Step 2: Add CommandDispatchedSSEDataV1 / CommandAckSSEDataV1 types**

---

## Chunk 3: Docs + Verification

### Task 7: Update API spec + handoff

**Files:**
- Modify: `shared/docs/API_SPEC.md`
- Modify: `packages/backend/docs/HANDOFF.md`
- Modify: `packages/frontend/docs/HANDOFF.md`

- [ ] **Step 1: telemetry SSE 示例 data 增加 schema_version**
- [ ] **Step 2: Add /api/devices/subscribe spec (device_codes filter + payload)**
- [ ] **Step 3: Add /api/commands/subscribe spec (device_codes filter + payloads)**
- [ ] **Step 4: Handoff 补充 P1 变更点**

### Task 8: Verify

- [ ] Run: `cd packages/backend && go test ./...`
- [ ] Run: `cd packages/frontend && npm run type-check`

