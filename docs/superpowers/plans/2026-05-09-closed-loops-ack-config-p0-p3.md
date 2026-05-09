# Closed Loops (2/3/4/5/7/8/9/10/12) — Ack+Config Protocol & P0–P3 Implementation Plan
> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Goal:** Make control/alert/config/batch automation loops reliable and explainable by standardizing `/ack` with `ack_type`, adding reliable config delivery (deliveries+ack+retry), and enabling batch stage automation with audited config application.
>
> **Architecture:** Keep MQTT topic structure stable, evolve payloads via `schema_version`, and introduce `config_deliveries` as the single source of truth for config revision (`entity_rev`). Backward compatible parsing preserves existing command ACK behavior while enabling config ACK and UI observability.
>
> **Tech Stack:** Go (Gin/GORM), MySQL, MQTT (EMQX), Vue3+TS (Element Plus), SSE.
>
> **Agreed decisions:** `msg_id`/`trace_id` use UUID; `entity_rev` lives only in `config_deliveries`; unified topic `/ack` with `ack_type`; first config types to land: `climate_profile` → `crop_batch_stage`.
+
---
+
## Protocol v1 (MQTT)
+
### Unified ACK: `AckEnvelopeV1`
+
Topic: `hydroponic/{deviceCode}/ack`
+
```json
{
  "schema_version": 1,
  "ack_type": "command | config",
  "msg_id": "uuid",
  "trace_id": "trace_xxx",
  "result": "ACKED | REJECTED | FAILED",
  "error_code": "OK | VALIDATION_FAILED | UNSUPPORTED_SCHEMA | APPLY_FAILED | BUSY | EXPIRED | NOT_FOUND | STALE",
  "error_message": "string",
  "device_ts_ms": 1710000000000,
  "payload": {}
}
```
+
Backward compatibility: if `schema_version` or `ack_type` missing → treat as legacy command ACK payload.
+
### Config push: `ConfigPushPayloadV1`
+
Topic: `hydroponic/{deviceCode}/cmd/config`
+
```json
{
  "schema_version": 1,
  "msg_id": "uuid",
  "trace_id": "trace_xxx",
  "config_type": "climate_profile | crop_batch_stage | control_policy | nutrient_target",
  "action": "create | update | delete",
  "entity_id": 123,
  "entity_rev": 7,
  "issued_at_ms": 1710000000000,
  "ttl_sec": 600,
  "require_ack": true,
  "payload": {}
}
```
+
Config ACK: use `AckEnvelopeV1` with `ack_type=config`, and include:
+
```json
{
  "config_type": "climate_profile",
  "action": "update",
  "entity_id": 123,
  "entity_rev": 7,
  "applied_hash": "sha256:....",
  "applied_at_ms": 1710000001000,
  "fw_version": "v1.2.3"
}
```
+
Idempotency/out-of-order rules (device-side):
- Track `(config_type, entity_id) -> last_applied_rev`.
- Lower rev: `REJECTED + STALE` (or ignore and ACKED, but must be consistent).
- Same rev: idempotent `ACKED` (do not re-apply).
- TTL expired: `REJECTED + EXPIRED`.
+
---
+
## Repository mapping (authoritative sources)
+
### Batch device binding (for stage/config targeting)
- Authoritative binding: `batch_devices` (`crop.BatchDevice`) in [crop/model.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/crop/model.go) and handlers in [batch_handler.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/crop/batch_handler.go).
- No direct batch↔channel binding table; channels derive via device:
  - `sensor_channels.sensor_device_id`
  - `actuator_channels.actuator_device_id` in [device/model.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/device/model.go).
- Stage config push default target: active batch-bound actuator devices (`device_type=actuator AND is_active=1`).
+
---
+
## Chunk 1: P0 — Stop-the-bleeding (consistency + security + stability)
+
### Task 1: Alerts SSE filtering + payload unification + auto-resolve timeline
+
**Files:**
- Modify: `packages/backend/internal/platform/event/sse_handler.go`
- Modify: `packages/backend/internal/alert/handler.go`
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`
- Modify: `packages/frontend/src/composables/useAlertSSE.ts`
- Modify: `packages/frontend/src/types/alert.ts`
- Docs: `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md`, `packages/frontend/docs/HANDOFF.md`
+
- [ ] **Step 1: Define AlertSSE DTO v1 (fields + schema_version)**
- [ ] **Step 2: Update SSE handler to filter `new_alert` by query params (level/device_code/greenhouse_id if available)**
- [ ] **Step 3: Make all `alert:created` publishers emit a consistent DTO (not a minimal map)**
- [ ] **Step 4: On heartbeat auto-resolve offline alerts, also append timeline event (RESOLVED + reason=heartbeat)**
- [ ] **Step 5: Update frontend types + `useAlertSSE` to consume DTO v1**
- [ ] **Step 6: Verify**
  - Run: `cd packages/backend && go test ./...`
  - Run: `cd packages/frontend && npm run type-check`
- [ ] **Step 7: Update docs**

### Task 2: Command ACK robustness (strong types) + publish failure semantics
+
**Files:**
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`
- Modify: `packages/backend/internal/command/waiter.go`
- Modify: `packages/backend/internal/command/model.go` (if status enum/fields need alignment)
- Docs: `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md`
+
- [ ] **Step 1: Introduce typed parsing for command ACK payload**
  - Refactor `handleAck()` to decode into a concrete struct, avoid `map[string]interface{}` in the internal event.
- [ ] **Step 2: Align waiter consumption to typed event payload**
- [ ] **Step 3: Ensure MQTT publish failure marks command FAILED (never SENT)**
  - Apply consistently in command dispatch paths (manual + policy + climate).
- [ ] **Step 4: Verify**
  - Run: `cd packages/backend && go test ./...`
- [ ] **Step 5: Update docs**

### Task 3: Device status validation + debounce knobs
+
**Files:**
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`
- Modify: `packages/backend/internal/device/offline_detector.go`
- Docs: `packages/backend/docs/HANDOFF.md`, `shared/docs/API_SPEC.md` (if behavior exposed)
+
- [ ] **Step 1: Validate status value set (ONLINE/OFFLINE/ERROR)**
- [ ] **Step 2: Add configurable offline thresholds / debounce to reduce alert flapping**
- [ ] **Step 3: Verify**
  - Run: `cd packages/backend && go test ./...`

### Task 4: Notification channel security + enum alignment
+
**Files:**
- Modify: `packages/backend/internal/notification/handler.go`
- Modify: `packages/frontend/src/types/notification.ts`
- Modify: `packages/frontend/src/views/settings/notification-channels.vue`
- Docs: `packages/backend/docs/HANDOFF.md`, `packages/frontend/docs/HANDOFF.md`, `shared/docs/API_SPEC.md`
+
- [ ] **Step 1: Enforce resource-level authorization for TestChannel (and other sensitive operations)**
- [ ] **Step 2: Align channel type enums (decide IN_APP visibility)**
- [ ] **Step 3: Verify**
  - Run: `cd packages/backend && go test ./...`
  - Run: `cd packages/frontend && npm run type-check`

---
+
## Chunk 2: P1 — Contract hardening (SSE/DTO schema_version)
+
### Task 5: SSE DTO v1 suite (alerts/commands/device_status/executions)
+
**Files:**
- Modify: `packages/backend/internal/platform/event/sse_handler.go`
- Modify: producers in `alert/`, `command/`, `device/`, `policy/`, `climate/` as needed
- Modify: `packages/frontend/src/composables/useTelemetrySSE.ts` (if affected)
- Modify: `packages/frontend/src/types/*` relevant type files
- Docs: `shared/docs/API_SPEC.md`, both HANDOFF.md
+
- [ ] **Step 1: Standardize SSE payload envelope per event with `schema_version`**
- [ ] **Step 2: Ensure producers emit full DTOs (or SSE handler enriches with DB reads)**
- [ ] **Step 3: Frontend aligns types + runtime guards (fail fast on unknown schema_version)**
- [ ] **Step 4: Verify**
  - Run: `cd packages/backend && go test ./...`
  - Run: `cd packages/frontend && npm run type-check`

---
+
## Chunk 3: P2 — Reliable config delivery (deliveries + unified ack_type=config)
+
### Task 6: Add `config_deliveries` table + model
+
**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_delivery_model.go` (or `internal/configsync/` module if preferred)
- Create: `packages/backend/migrations/<new>_config_deliveries.up.sql`
- Modify: `packages/backend/migrations/merged/all.up.sql` (if required by repo process)
- Docs: `packages/backend/docs/HANDOFF.md`, `docs/PROJECT_STATUS.md` (if needed)
+
- [ ] **Step 1: Design schema**
  - `msg_id` (unique), `trace_id`, `device_code`, `config_type`, `action`, `entity_id`, `entity_rev`, `status`, `retry_count`, `next_retry_at`, `last_error_code`, `last_error_message`, `acked_at`, `applied_hash`, `device_fw_version`, timestamps
- [ ] **Step 2: Write migration + GORM model**
- [ ] **Step 3: Verify build**
  - Run: `cd packages/backend && go test ./...`

### Task 7: Implement ConfigPushPayload v1 + record deliveries
+
**Files:**
- Modify: `packages/backend/internal/platform/mqtt/config_pusher.go`
- Modify: `packages/backend/internal/climate/profile_handler.go` (first adopter)
- Create: `packages/backend/internal/platform/mqtt/config_delivery_repo.go`
- Docs: `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md`
+
- [ ] **Step 1: Extend payload struct with v1 fields (`schema_version`, `msg_id`, `trace_id`, `entity_rev`, `issued_at_ms`, `ttl_sec`, `require_ack`)**
- [ ] **Step 2: Allocate `entity_rev` per (device_code, config_type, entity_id) by reading latest delivery and incrementing**
- [ ] **Step 3: Persist delivery with status transitions PENDING→SENT**
- [ ] **Step 4: Ensure MQTT not connected results in delivery FAILED with reason (do not silently skip)**
- [ ] **Step 5: Verify**
  - Run: `cd packages/backend && go test ./...`

### Task 8: Parse unified `/ack` with `ack_type=config` (backward compatible)
+
**Files:**
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`
- Create: `packages/backend/internal/platform/mqtt/ack_parser.go`
- Docs: `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md`
+
- [ ] **Step 1: Implement ack parser**
  - If `schema_version` present: decode `AckEnvelopeV1` and route by `ack_type`
  - Else: decode legacy command ack struct and route as `command`
- [ ] **Step 2: For `ack_type=config`, update `config_deliveries` with ACKED/REJECTED/FAILED + ack metadata**
- [ ] **Step 3: Publish internal event `config:acked` for UI / monitoring**
- [ ] **Step 4: Verify**
  - Run: `cd packages/backend && go test ./...`

### Task 9: Retry worker + observability surface
+
**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_retry_worker.go`
- Modify: `packages/backend/internal/platform/http/router.go` (startup hook) or DI wiring
- Create/Modify: `packages/backend/internal/platform/mqtt/config_delivery_handler.go` (optional HTTP list API)
- Frontend: `packages/frontend/src/views/settings/` or `views/climate/` config delivery list
- Docs: `shared/docs/API_SPEC.md`, both HANDOFF.md
+
- [ ] **Step 1: Add periodic retry worker**
- [ ] **Step 2: Add minimal HTTP API for deliveries (list by device/config_type/entity_id)**
- [ ] **Step 3: Add frontend page/section to show status (SENT/ACKED/FAILED/RETRYING)**
- [ ] **Step 4: Verify**
  - Run: `cd packages/backend && go test ./...`
  - Run: `cd packages/frontend && npm run type-check`

---
+
## Chunk 4: P3 — Batch stage automation + stage config push (crop_batch_stage)
+
### Task 10: Batch stage scheduler (time-window driven)
+
**Files:**
- Create: `packages/backend/internal/crop/batch_stage_scheduler.go`
- Modify: `packages/backend/internal/crop/routes.go` (startup hook)
- Modify: `packages/backend/internal/crop/model.go` (only if scheduler needs new state fields; avoid if possible)
- Docs: `packages/backend/docs/HANDOFF.md`, `docs/PROJECT_STATUS.md` (if schema changes), `shared/docs/API_SPEC.md` (if new APIs)
+
- [ ] **Step 1: Define scheduler behavior**
  - Determine current stage by `batch_stage_plans` time window
  - Emit stage-transition events (audit trail) when entering new stage
  - Allow manual override/pause (decide minimal control flag location)
- [ ] **Step 2: Implement scheduler loop + safe locking**
- [ ] **Step 3: Verify**
  - Run: `cd packages/backend && go test ./...`

### Task 11: Stage config payload + reliable push to batch-bound actuators
+
**Files:**
- Modify: `packages/backend/internal/platform/mqtt/config_pusher.go`
- Modify: `packages/backend/internal/crop/batch_stage_scheduler.go`
- Create: `packages/backend/internal/crop/stage_config_builder.go`
- Docs: `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md`
+
- [ ] **Step 1: Define `config_type=crop_batch_stage` payload schema (EC/PH targets + stage window + optional bindings)**
- [ ] **Step 2: Resolve target devices**
  - Query `batch_devices` where `device_type=actuator AND is_active=1`, map to actuator `device_code`
- [ ] **Step 3: Push v1 config + record deliveries + await ack asynchronously**
- [ ] **Step 4: Verify**
  - Run: `cd packages/backend && go test ./...`

### Task 12: Frontend batch stage observability (push/ack status)
+
**Files:**
- Modify: `packages/frontend/src/views/batches/detail.vue`
- Modify: `packages/frontend/src/views/batches/stage-plans.vue`
- Create/Modify: `packages/frontend/src/api/` module for deliveries (if added)
- Modify: `packages/frontend/src/types/` for delivery DTOs
- Docs: `packages/frontend/docs/HANDOFF.md`, `shared/docs/API_SPEC.md`
+
- [ ] **Step 1: Add UI sections**
  - Current stage
  - Latest stage config push per device (status + last_error + applied_hash)
- [ ] **Step 2: Verify**
  - Run: `cd packages/frontend && npm run type-check`

---
+
## Documentation checklist (must-do)
- Update `shared/docs/API_SPEC.md` with:
  - MQTT Protocol section (AckEnvelopeV1, ConfigPushPayloadV1)
  - Delivery status API (if introduced)
- Update `packages/backend/docs/HANDOFF.md` and `packages/frontend/docs/HANDOFF.md` after each chunk.
- If any schema changes beyond new deliveries table: update `packages/backend/docs/PROJECT_STATUS.md`.

