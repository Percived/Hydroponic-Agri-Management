# P3 Crop Batch Stage Config Push + Auto Stage Scheduler Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 基于 `batch_stage_plans` 的时间窗自动判定批次当前阶段，在阶段切换时按 `stage_plan` 权威配置切换 `active_*` 并对批次绑定的 actuator 设备下发 `crop_batch_stage` 配置（走 P2 reliable deliveries + ack + retry）。

**Architecture:** `batch_stage_plans` 扩展为阶段级配置权威（recipe_id/policy_id/climate_profile_id）。新增 `batch_stage_runtime` 记录上次已应用的 stage_plan，scheduler 周期扫描 RUNNING 批次，检测阶段变化并触发：更新 runtime + crop_batches.active_* + config push（`crop_batch_stage`，并在配置引用到 climate_profile 时额外推送 `climate_profile` 快照）。

**Tech Stack:** Go (Gin/GORM), MySQL migrations (merged/all + v2.x patch), MQTT config push (P2 `config_deliveries`), Vue3 + Element Plus.

---

## File Structure (locked)

- Migrations
  - Modify: `packages/backend/migrations/merged/all.up.sql`
  - Modify: `packages/backend/migrations/merged/all.down.sql`
  - Create: `packages/backend/migrations/merged/v2.4_batch_stage_config_refs.up.sql`
  - Create: `packages/backend/migrations/merged/v2.4_batch_stage_config_refs.down.sql`
- Backend (crop)
  - Modify: `packages/backend/internal/crop/model.go`
  - Modify: `packages/backend/internal/crop/dto.go`
  - Modify: `packages/backend/internal/crop/harvest_handler.go`
  - Create: `packages/backend/internal/crop/batch_stage_runtime_model.go`
  - Create: `packages/backend/internal/crop/batch_stage_scheduler.go`
  - Modify: `packages/backend/internal/crop/routes.go` (wire scheduler)
  - Modify: `packages/backend/internal/crop/batch_handler.go` (ensure active_* fields included in response if needed)
- Backend (climate helper)
  - Modify: `packages/backend/internal/climate/dto.go` (export build function for config payload)
  - Modify: `packages/backend/internal/climate/profile_handler.go` (reuse exported builder)
- Tests
  - Create: `packages/backend/internal/crop/batch_stage_scheduler_test.go`
- Frontend
  - Modify: `packages/frontend/src/types/crop.ts`
  - Modify: `packages/frontend/src/api/crop.ts`
  - Modify: `packages/frontend/src/components/batch/StagePlanEditor.vue`
  - Modify: `packages/frontend/src/views/batches/stage-plans.vue`
- Docs
  - Modify: `shared/docs/API_SPEC.md`（stage_plan 新字段）
  - Modify: `packages/backend/docs/HANDOFF.md`

---

## Chunk 1: Schema + Models + DTOs

### Task 1: Extend `batch_stage_plans` with config refs; add runtime table; add active_climate_profile_id

**Files:**
- Modify: `packages/backend/migrations/merged/all.up.sql`
- Modify: `packages/backend/migrations/merged/all.down.sql`
- Create: `packages/backend/migrations/merged/v2.4_batch_stage_config_refs.up.sql`
- Create: `packages/backend/migrations/merged/v2.4_batch_stage_config_refs.down.sql`

- [ ] **Step 1: Add columns to `batch_stage_plans`**

```sql
ALTER TABLE `batch_stage_plans`
  ADD COLUMN `recipe_id` BIGINT UNSIGNED NULL AFTER `growth_stage_id`,
  ADD COLUMN `policy_id` BIGINT UNSIGNED NULL AFTER `recipe_id`,
  ADD COLUMN `climate_profile_id` BIGINT UNSIGNED NULL AFTER `policy_id`;
```

- [ ] **Step 2: Add `active_climate_profile_id` to `crop_batches`**

```sql
ALTER TABLE `crop_batches`
  ADD COLUMN `active_climate_profile_id` BIGINT UNSIGNED NULL AFTER `active_policy_id`;
```

- [ ] **Step 3: Create `batch_stage_runtime`**

```sql
CREATE TABLE IF NOT EXISTS `batch_stage_runtime` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `batch_id` BIGINT UNSIGNED NOT NULL,
  `current_stage_plan_id` BIGINT UNSIGNED NULL,
  `current_growth_stage_id` BIGINT UNSIGNED NULL,
  `last_switched_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_batch_stage_runtime_batch_id` (`batch_id`),
  KEY `idx_batch_stage_runtime_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 4: Down migration**

```sql
DROP TABLE IF EXISTS `batch_stage_runtime`;
ALTER TABLE `crop_batches` DROP COLUMN `active_climate_profile_id`;
ALTER TABLE `batch_stage_plans` DROP COLUMN `climate_profile_id`, DROP COLUMN `policy_id`, DROP COLUMN `recipe_id`;
```

### Task 2: Backend model & DTO updates

**Files:**
- Modify: `packages/backend/internal/crop/model.go`
- Modify: `packages/backend/internal/crop/dto.go`
- Modify: `packages/backend/internal/crop/harvest_handler.go`
- Create: `packages/backend/internal/crop/batch_stage_runtime_model.go`

- [ ] **Step 1: Extend `BatchStagePlan` model with new fields**
- [ ] **Step 2: Extend stage plan create/update requests and response**
- [ ] **Step 3: Update stage plan handlers to persist/read new fields**

---

## Chunk 2: Scheduler + Config Push

### Task 3: Implement `BatchStageScheduler`

**Files:**
- Create: `packages/backend/internal/crop/batch_stage_scheduler.go`
- Create: `packages/backend/internal/crop/batch_stage_scheduler_test.go`
- Modify: `packages/backend/internal/crop/routes.go`

- [ ] **Step 1: Write failing tests**
  - stage plan changes (plan id changes) triggers runtime update + batch active_* updates + config push called once per device
  - repeated ticks without stage change should not re-push
- [ ] **Step 2: Implement scheduler**
  - tick every 30s/60s
  - query RUNNING batches
  - for each batch: query current stage_plan (same SQL as stage_progress：`start_at <= now < end_at`)
  - compare with `batch_stage_runtime.current_stage_plan_id`; if different -> trigger switch
  - select target actuator device_codes via `batch_devices` join `actuator_devices`
- [ ] **Step 3: Config payload**

```go
type CropBatchStageConfigPayloadV1 struct {
  SchemaVersion int `json:"schema_version"`
  BatchID uint64 `json:"batch_id"`
  StagePlanID uint64 `json:"stage_plan_id"`
  GrowthStageID uint64 `json:"growth_stage_id"`
  StageStartAt string `json:"stage_start_at"`
  StageEndAt string `json:"stage_end_at"`
  Targets struct{
    ECMin *float64 `json:"ec_min"`
    ECMax *float64 `json:"ec_max"`
    PHMin *float64 `json:"ph_min"`
    PHMax *float64 `json:"ph_max"`
  } `json:"targets"`
  RecipeID *uint64 `json:"recipe_id"`
  PolicyID *uint64 `json:"policy_id"`
  ClimateProfileID *uint64 `json:"climate_profile_id"`
  SwitchedAt string `json:"switched_at"`
  Reason string `json:"reason"`
}
```

Push:
  - `config_type="crop_batch_stage"`, `action="update"`, `entity_id=batchID`, `payload=payloadV1`
  - if `climate_profile_id!=nil`: additionally push `config_type="climate_profile"` with profile snapshot payload

- [ ] **Step 4: Wire scheduler start in crop routes**

---

## Chunk 3: Climate payload builder reuse

### Task 4: Export climate profile payload builder

**Files:**
- Modify: `packages/backend/internal/climate/dto.go`
- Modify: `packages/backend/internal/climate/profile_handler.go`

- [ ] **Step 1: Add `BuildProfileConfigPayload(profile ClimateProfile) ClimateProfileResponse`**
- [ ] **Step 2: Update existing climate config push to reuse it**

---

## Chunk 4: Frontend stage plan editor + list

### Task 5: Frontend updates

**Files:**
- Modify: `packages/frontend/src/types/crop.ts`
- Modify: `packages/frontend/src/api/crop.ts`
- Modify: `packages/frontend/src/components/batch/StagePlanEditor.vue`
- Modify: `packages/frontend/src/views/batches/stage-plans.vue`

- [ ] **Step 1: Extend types and request payload to include recipe_id/policy_id/climate_profile_id**
- [ ] **Step 2: StagePlanEditor add 3 selects (recipes/policies/climate profiles)**
- [ ] **Step 3: Stage plans table display readable names (fallback to ID)**

---

## Chunk 5: Docs + Verification

### Task 6: Docs

**Files:**
- Modify: `shared/docs/API_SPEC.md`
- Modify: `packages/backend/docs/HANDOFF.md`

- [ ] **Step 1: Document new stage plan fields and stage switch behavior**

### Task 7: Verify

- [ ] Run: `cd packages/backend && go test ./...`
- [ ] Run: `cd packages/frontend && npm run type-check`

