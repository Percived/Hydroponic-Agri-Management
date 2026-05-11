# P2 Reliable Config Delivery (config_deliveries + ack_type=config + retry) Implementation Plan
> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `cmd/config` 从“尽力发送一条 MQTT”升级为可审计、可重试、可确认（ACK）、可观测的可靠投递闭环，并保持对 legacy command ack 的兼容。

**Architecture:** Producer 侧写入 `config_deliveries` 作为投递权威（含 `entity_rev/msg_id/trace_id/request_payload`），MQTT publish 后推进状态；Ingress 解析 unified `/ack` 的 `ack_type=config` 回写 deliveries；worker 负责超时与重试；可选提供 HTTP list API 便于验收与排障。

**Tech Stack:** Go (Gin/GORM), MySQL migrations (merged/all + v2.x patch), MQTT (EMQX), internal EventHub.

---

## File Structure (locked)

- Migrations
  - Modify: `packages/backend/migrations/merged/all.up.sql`
  - Modify: `packages/backend/migrations/merged/all.down.sql`
  - Create: `packages/backend/migrations/merged/v2.3_config_deliveries.up.sql`
  - Create: `packages/backend/migrations/merged/v2.3_config_deliveries.down.sql`
- Backend (mqtt/config delivery)
  - Create: `packages/backend/internal/platform/mqtt/config_delivery_model.go`
  - Create: `packages/backend/internal/platform/mqtt/config_delivery_repo.go`
  - Create: `packages/backend/internal/platform/mqtt/config_retry_worker.go`
  - Create: `packages/backend/internal/platform/mqtt/ack_parser.go`
  - Modify: `packages/backend/internal/platform/mqtt/config_pusher.go`
  - Modify: `packages/backend/internal/platform/mqtt/ingress.go`
  - Modify: `packages/backend/internal/climate/profile_handler.go` (first adopter: climate_profile)
  - Modify: `packages/backend/internal/platform/http/router.go` (worker wiring + optional HTTP API)
- Backend (optional observability API)
  - Create: `packages/backend/internal/platform/mqtt/config_delivery_handler.go`
- Tests
  - Create: `packages/backend/internal/platform/mqtt/ack_parser_test.go`
  - Create: `packages/backend/internal/platform/mqtt/config_delivery_repo_test.go`
  - Create: `packages/backend/internal/platform/mqtt/config_retry_worker_test.go`
  - Update: existing mqtt/ingress tests if needed
- Docs
  - Modify: `shared/docs/API_SPEC.md`（新增投递列表 API；描述 mqtt 不连时语义为 FAILED）
  - Modify: `packages/backend/docs/HANDOFF.md`

---

## Chunk 1: Schema + Model

### Task 1: Add `config_deliveries` table

**Files:**
- Modify: `packages/backend/migrations/merged/all.up.sql`
- Modify: `packages/backend/migrations/merged/all.down.sql`
- Create: `packages/backend/migrations/merged/v2.3_config_deliveries.up.sql`
- Create: `packages/backend/migrations/merged/v2.3_config_deliveries.down.sql`

- [ ] **Step 1: Add DDL (all.up.sql + v2.3 up)**

```sql
CREATE TABLE IF NOT EXISTS `config_deliveries` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `msg_id` VARCHAR(64) NOT NULL,
  `trace_id` VARCHAR(64) NOT NULL DEFAULT '',
  `device_code` VARCHAR(64) NOT NULL,
  `config_type` VARCHAR(64) NOT NULL,
  `action` VARCHAR(16) NOT NULL,
  `entity_id` BIGINT UNSIGNED NOT NULL,
  `entity_rev` BIGINT UNSIGNED NOT NULL,
  `schema_version` INT NOT NULL DEFAULT 1,
  `issued_at_ms` BIGINT UNSIGNED NOT NULL,
  `ttl_sec` INT NOT NULL DEFAULT 600,
  `require_ack` TINYINT(1) NOT NULL DEFAULT 1,
  `request_payload` JSON NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'PENDING', -- PENDING/SENT/ACKED/REJECTED/FAILED
  `retry_count` INT NOT NULL DEFAULT 0,
  `next_retry_at` DATETIME NULL,
  `sent_at` DATETIME NULL,
  `acked_at` DATETIME NULL,
  `last_error_code` VARCHAR(64) NOT NULL DEFAULT '',
  `last_error_message` VARCHAR(255) NOT NULL DEFAULT '',
  `ack_payload` JSON NULL,
  `applied_hash` VARCHAR(128) NOT NULL DEFAULT '',
  `device_fw_version` VARCHAR(64) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_config_deliveries_msg_id` (`msg_id`),
  KEY `idx_config_deliveries_entity_rev` (`device_code`, `config_type`, `entity_id`, `entity_rev`),
  KEY `idx_config_deliveries_status_retry` (`status`, `next_retry_at`),
  KEY `idx_config_deliveries_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Add down migration (all.down.sql + v2.3 down)**

```sql
DROP TABLE IF EXISTS `config_deliveries`;
```

- [ ] **Step 3: Verify local migration scripts parse**

### Task 2: Add GORM model

**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_delivery_model.go`

- [ ] **Step 1: Define `ConfigDelivery` model + TableName**
- [ ] **Step 2: Define status constants**

---

## Chunk 2: Repo + Rev allocation

### Task 3: Repository methods

**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_delivery_repo.go`
- Test: `packages/backend/internal/platform/mqtt/config_delivery_repo_test.go`

- [ ] **Step 1: Write failing tests**
  - allocate next rev increments per `(device_code, config_type, entity_id)`
  - create delivery persists request_payload + status=PENDING
  - mark sent/failed/acked transitions update timestamps
- [ ] **Step 2: Implement repo**
  - `AllocateNextRev(tx, device, type, entityID) (rev uint64, err)`
  - `CreatePending(tx, delivery) error`
  - `MarkSent(id, sentAt)`
  - `MarkFailed(id, code, msg, nextRetryAt)`
  - `MarkAckedByMsgID(msgID, ackedAt, ackPayloadJSON, fwVersion, appliedHash)`
  - `ListRetryCandidates(now, limit)`
- [ ] **Step 3: Run tests**

Run: `cd packages/backend && go test ./...`
Expected: PASS

---

## Chunk 3: MQTT publish path (ConfigPushPayloadV1) + producer integration

### Task 4: Extend config payload to v1 + deliveries recording

**Files:**
- Modify: `packages/backend/internal/platform/mqtt/config_pusher.go`
- Modify: `packages/backend/internal/climate/profile_handler.go`

- [ ] **Step 1: Define ConfigPushPayloadV1 struct**

```go
type ConfigPushPayloadV1 struct {
  SchemaVersion int `json:"schema_version"`
  MsgID string `json:"msg_id"`
  TraceID string `json:"trace_id"`
  ConfigType string `json:"config_type"`
  Action string `json:"action"`
  EntityID uint64 `json:"entity_id"`
  EntityRev uint64 `json:"entity_rev"`
  IssuedAtMS uint64 `json:"issued_at_ms"`
  TTLsec int `json:"ttl_sec"`
  RequireAck bool `json:"require_ack"`
  Payload interface{} `json:"payload"`
}
```

- [ ] **Step 2: New push method returns delivery + persists state**
  - transactional：allocate rev → create delivery(PENDING) → publish MQTT → mark SENT（或失败则 FAILED）
  - MQTT not connected / device lookup failed：不得静默成功，必须落 `FAILED`（last_error_code=MQTT_NOT_CONNECTED/DEVICE_NOT_FOUND）
- [ ] **Step 3: climate_profile 作为首个 adopter**
  - `payload` 至少包含一个非空对象（例如 profile+stages+actions 的快照或最小配置）
  - 删除 action 的时候也触发 profile update push（修复 P1 发现的遗漏点）
- [ ] **Step 4: Add targeted tests**
  - mqtt not connected -> delivery FAILED

---

## Chunk 4: Unified /ack parsing + ack_type=config

### Task 5: Ack parser

**Files:**
- Create: `packages/backend/internal/platform/mqtt/ack_parser.go`
- Test: `packages/backend/internal/platform/mqtt/ack_parser_test.go`
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`

- [ ] **Step 1: Add parser structs**

```go
type AckEnvelopeV1 struct {
  SchemaVersion int `json:"schema_version"`
  AckType string `json:"ack_type"` // command|config
  MsgID string `json:"msg_id"`
  TraceID string `json:"trace_id"`
  Result string `json:"result"` // ACKED|REJECTED|FAILED
  ErrorCode string `json:"error_code"`
  ErrorMessage string `json:"error_message"`
  DeviceTSms uint64 `json:"device_ts_ms"`
  Payload map[string]interface{} `json:"payload"`
}
```

- [ ] **Step 2: Parser behavior**
  - if schema_version missing -> legacy command ack (existing behavior)
  - if ack_type=config -> return config ack
  - if ack_type=command -> optional: keep as-is (future)
- [ ] **Step 3: Wire ingress.handleAck**
  - config ack: `MarkAckedByMsgID` / `MarkFailed` / `MarkRejected` (map to deliveries)
  - legacy command ack: keep existing flow (update control_commands + publish command:acked)

---

## Chunk 5: Retry worker + (optional) HTTP observability

### Task 6: Retry worker

**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_retry_worker.go`
- Test: `packages/backend/internal/platform/mqtt/config_retry_worker_test.go`
- Modify: `packages/backend/internal/platform/http/router.go`

- [ ] **Step 1: Worker loop**
  - tick every N seconds
  - mark SENT overtime as FAILED(ACK_TIMEOUT) with next_retry_at
  - resend FAILED when next_retry_at due
- [ ] **Step 2: Start worker in router**
  - keep single instance assumption for now
- [ ] **Step 3: Tests**（sqlite in-memory）

### Task 7 (optional but recommended): List API for deliveries

**Files:**
- Create: `packages/backend/internal/platform/mqtt/config_delivery_handler.go`
- Modify: `packages/backend/internal/platform/http/router.go`
- Modify: `shared/docs/API_SPEC.md`

- [ ] **Step 1: `GET /api/config-deliveries` with filters + paging**
- [ ] **Step 2: `GET /api/config-deliveries/:id`**
- [ ] **Step 3: Update API spec with request/response examples**

---

## Chunk 6: Docs + Verification

### Task 8: Docs

**Files:**
- Modify: `shared/docs/API_SPEC.md`
- Modify: `packages/backend/docs/HANDOFF.md`

- [ ] **Step 1: Document config delivery behavior (mqtt not connected => FAILED, retries, statuses)**
- [ ] **Step 2: Document optional deliveries list API**

### Task 9: Verify

- [ ] Run: `cd packages/backend && go test ./...`
- [ ] Run: `cd packages/frontend && npm run type-check`

