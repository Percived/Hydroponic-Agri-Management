# 气候联动触发源固定单一采集通道 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将气候联动（Climate Profiles）的自动触发依据从“按指标全局触发”改为“固定单一采集通道触发”，并补齐禁用联动与可观测日志。

**Architecture:** 在 `climate_profiles` 中新增 `trigger_sensor_channel_id`，调度器仅消费 `telemetry:received` 事件里匹配该 `sensor_channel_id` 的 profile。通道被禁用时自动停用引用的 profile；通道删除在产品层面禁止（仅允许禁用）。执行日志补记触发来源字段。

**Tech Stack:** Go (Gin + GORM + MySQL), Vue 3 + TS + Element Plus, MQTT (paho), SSE EventHub。

---

## Chunk 1: 后端数据与API契约

### Task 1: 数据库迁移（climate_profiles + climate_execution_logs）

**Files:**
- Modify: [all.up.sql](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/migrations/merged/all.up.sql#L446-L509)
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/migrations/merged/all.down.sql`（对应回滚）
- Create: `e:/goProject/Hydroponic-Agri-Management/packages/backend/migrations/merged/v2.3.2_climate_trigger_channel.up.sql`
- Create: `e:/goProject/Hydroponic-Agri-Management/packages/backend/migrations/merged/v2.3.2_climate_trigger_channel.down.sql`

- [ ] **Step 1: 写迁移脚本（UP）**
  - `climate_profiles`：新增列 `trigger_sensor_channel_id BIGINT UNSIGNED`（迁移期允许 NULL），回填后改为 NOT NULL；新增索引 `idx_climate_profiles_trigger_sc (trigger_sensor_channel_id, enabled)`
  - `climate_execution_logs`：新增列（允许 NULL 以兼容手动执行/历史数据）
    - `trigger_sensor_channel_id BIGINT UNSIGNED NULL`
    - `trigger_metric_code VARCHAR(32) NULL`
    - `collected_at DATETIME(3) NULL`
    - 索引建议：`idx_climate_log_trigger_sc_time (trigger_sensor_channel_id, executed_at)`

- [ ] **Step 2: 回填策略**
  - 默认不做自动推导（因为旧数据只存 metric_code，无法唯一定位通道）
  - 迁移脚本里仅新增列，不强制回填
  - 在后端 Profile 更新/启用时做强校验，保证新建/编辑后的数据完整

- [ ] **Step 3: 更新全量初始化脚本**
  - 在 `all.up.sql` 的 `CREATE TABLE climate_profiles` 中补上新列与索引
  - 在 `all.up.sql` 的 `CREATE TABLE climate_execution_logs` 中补上新列与索引
  - 在 `all.down.sql` 对应补充 drop column（按项目惯例）

- [ ] **Step 4: 验证迁移 SQL 可执行**
  - Run: 通过现有迁移方式在本地 MySQL 执行（参考根 `CLAUDE.md` 的 migration 命令）
  - Expected: SQL 执行成功；新列可见；索引存在

- [ ] **Step 5: Commit**
  - `git add packages/backend/migrations/merged/*`
  - `git commit -m "feat(climate): add trigger sensor channel and log fields"`

### Task 2: 后端模型与DTO更新

**Files:**
- Modify: [climate/model.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/model.go#L5-L66)
- Modify: [climate/dto.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/dto.go#L7-L166)

- [ ] **Step 1: 更新 GORM Model**

```go
// ClimateProfile
TriggerSensorChannelID uint64 `gorm:"column:trigger_sensor_channel_id;not null"`

// ClimateExecutionLog
TriggerSensorChannelID *uint64   `gorm:"column:trigger_sensor_channel_id"`
TriggerMetricCode      *string   `gorm:"column:trigger_metric_code;size:32"`
CollectedAt            *time.Time `gorm:"column:collected_at"`
```

- [ ] **Step 2: 更新 DTO（请求/响应）**
  - `CreateClimateProfileRequest` / `CreateClimateProfileWithStagesRequest` 增加必填字段：
    - `TriggerSensorChannelID uint64 json:"trigger_sensor_channel_id" binding:"required"`
  - `UpdateClimateProfileRequest` 增加可选字段：
    - `TriggerSensorChannelID *uint64 json:"trigger_sensor_channel_id"`
  - `ClimateProfileResponse` 增加：
    - `TriggerSensorChannelID uint64 json:"trigger_sensor_channel_id"`
  - `ClimateExecutionLogResponse` 增加（omitempty/可空）：
    - `trigger_sensor_channel_id`
    - `trigger_metric_code`
    - `collected_at`

- [ ] **Step 3: Run 编译校验**
  - Run: `cd packages/backend && go test ./...`
  - Expected: 编译通过（即使没有测试用例也应 PASS）

- [ ] **Step 4: Commit**
  - `git add packages/backend/internal/climate/model.go packages/backend/internal/climate/dto.go`
  - `git commit -m "feat(climate): extend models and DTOs for trigger channel"`

### Task 3: Profile 创建/更新/嵌套创建的强校验

**Files:**
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/profile_handler.go`
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/handler.go`（若手动执行写日志需要读取 profile 新字段）

- [ ] **Step 1: 添加校验函数（climate 模块内，避免跨模块 import）**
  - 通过 `db.Table("sensor_channels") JOIN sensor_devices` 校验：
    - channel 存在且 enabled=1
    - channel.metric_code == trigger_metric_code
    - device.greenhouse_id == greenhouse_id

示例（伪码）：

```go
type channelCheck struct{ MetricCode string; GreenhouseID uint64; Enabled uint8 }
err := db.Table("sensor_channels sc").
  Select("sc.metric_code, sd.greenhouse_id, sc.enabled").
  Joins("JOIN sensor_devices sd ON sd.id = sc.sensor_device_id").
  Where("sc.id = ?", triggerSensorChannelID).
  Scan(&out).Error
```

- [ ] **Step 2: 在 CreateProfile / CreateProfileWithStages / UpdateProfile 中调用校验**
  - Create：必填校验失败 → 400 validation
  - Update：
    - 若更新 `trigger_sensor_channel_id` 或 `trigger_metric_code`，必须同时满足一致性
    - 若将 `enabled` 设为 true，也必须验证通道存在且 enabled

- [ ] **Step 3: 编译验证**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 4: Commit**
  - `git add packages/backend/internal/climate/profile_handler.go packages/backend/internal/climate/handler.go`
  - `git commit -m "feat(climate): validate trigger channel consistency"`

### Task 4: 更新 API 文档（共享规范 + OpenAPI）

**Files:**
- Modify: `e:/goProject/Hydroponic-Agri-Management/shared/docs/API_SPEC.md`（气候控制章节）
- Modify: `e:/goProject/Hydroponic-Agri-Management/shared/docs/openapi.yaml`（对应 schema/params）
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/docs/HANDOFF.md`

- [ ] **Step 1: 更新 API_SPEC**
  - 在 climate profile 创建/更新/响应里加入 `trigger_sensor_channel_id`
  - 明确自动调度按 `sensor_channel_id` 触发，不受其他温室/其他探头影响
  - 增加“通道禁用会自动停用 Profile；通道删除禁止（仅禁用）”的行为说明
  - 执行日志响应补字段：`trigger_sensor_channel_id/trigger_metric_code/collected_at`

- [ ] **Step 2: 更新 openapi.yaml**
  - 更新对应 request/response schema
  - 确保 required 字段一致

- [ ] **Step 3: 更新 backend HANDOFF**
  - 记录本次 schema + 行为变化（便于交接）

- [ ] **Step 4: Commit**
  - `git add shared/docs/API_SPEC.md shared/docs/openapi.yaml packages/backend/docs/HANDOFF.md`
  - `git commit -m "docs: update climate API for trigger channel"`

---

## Chunk 2: 后端行为改造（调度器、通道禁用联动、删除策略、日志增强）

### Task 5: 自动调度器按 sensor_channel_id 触发

**Files:**
- Modify: [profile_scheduler.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/profile_scheduler.go#L1-L344)

- [ ] **Step 1: 调整事件消费**
  - 从 event data 读取：`sensor_channel_id`、`metric_code`、`value`、`collected_at`
  - 调用新函数：`evaluateProfilesByChannel(sensorChannelID, metricCode, value, collectedAt)`

- [ ] **Step 2: 调整 profiles 查询条件**
  - `WHERE enabled=true AND trigger_sensor_channel_id = ?`
  - 不再按 `trigger_metric_code` 查询（由写入时校验保证一致性）

- [ ] **Step 3: 写入执行日志时补齐新字段**
  - `trigger_sensor_channel_id`：来自事件
  - `trigger_metric_code`：来自事件（优先真实事件值）
  - `collected_at`：来自事件（解析 RFC3339）；解析失败则置空或用接收时间

- [ ] **Step 4: MQTT 不可用的现状保持**
  - 本任务不改 MQTT offline 时 status 语义（属于后续质量改进项）

- [ ] **Step 5: 编译验证**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 6: Commit**
  - `git add packages/backend/internal/climate/profile_scheduler.go`
  - `git commit -m "feat(climate): trigger scheduler by sensor_channel_id"`

### Task 6: 手动执行写日志补字段

**Files:**
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/climate/handler.go`

- [ ] **Step 1: 在 ExecuteProfile 写 log 时补齐字段**
  - `trigger_sensor_channel_id`：取 profile.trigger_sensor_channel_id
  - `trigger_metric_code`：取 profile.trigger_metric_code
  - `collected_at`：无遥测事件，写 `executed_at`

- [ ] **Step 2: 编译验证**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 3: Commit**
  - `git add packages/backend/internal/climate/handler.go`
  - `git commit -m "feat(climate): enrich manual execution logs"`

### Task 7: 采集通道禁用 → 自动停用引用的气候 Profile

**Files:**
- Modify: [device/handler.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/device/handler.go#L240-L295)

- [ ] **Step 1: 在 UpdateSensorChannel 中识别 enabled=false 的变更**
  - 若请求把通道 `enabled` 更新为 false（且数据库 rowsAffected>0）：
    - 执行：`UPDATE climate_profiles SET enabled=0 WHERE trigger_sensor_channel_id = :id`
  - 建议与通道更新放到同一个事务中，减少竞态窗口

- [ ] **Step 2: 编译验证**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 3: Commit**
  - `git add packages/backend/internal/device/handler.go`
  - `git commit -m "feat(device): disable climate profiles when sensor channel disabled"`

### Task 8: 禁止删除采集通道（仅禁用）

**Files:**
- Modify: [device/handler.go](file:///e:/goProject/Hydroponic-Agri-Management/packages/backend/internal/device/handler.go#L351-L369)
- Modify: `e:/goProject/Hydroponic-Agri-Management/shared/docs/API_SPEC.md`（设备-采集通道 DELETE 行为说明）

- [ ] **Step 1: 修改 DeleteSensorChannel 行为**
  - 返回 409 conflict（或 400），message/key 类似 `delete_disabled_use_disable`
  - 不执行物理删除

- [ ] **Step 2: 更新 API 文档对应端点说明**

- [ ] **Step 3: 编译验证**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 4: Commit**
  - `git add packages/backend/internal/device/handler.go shared/docs/API_SPEC.md`
  - `git commit -m "feat(device): forbid deleting sensor channels"`

---

## Chunk 3: 前端改造（气候联动页面）

### Task 9: 前端类型与 API 对齐

**Files:**
- Modify: [types/climate.ts](file:///e:/goProject/Hydroponic-Agri-Management/packages/frontend/src/types/climate.ts#L1-L106)
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/frontend/src/api/climate.ts`
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/frontend/docs/HANDOFF.md`

- [ ] **Step 1: 更新 types**
  - `ClimateProfile` 增加 `trigger_sensor_channel_id: number`
  - `CreateClimateProfileRequest` / `CreateClimateProfileWithStagesRequest` 增加 `trigger_sensor_channel_id`
  - `ClimateExecutionLog` 增加：
    - `trigger_sensor_channel_id?: number`
    - `trigger_metric_code?: string`
    - `collected_at?: string`

- [ ] **Step 2: 更新 api/climate.ts 的请求体类型引用**
  - 确保 create/update/full-create 传新字段

- [ ] **Step 3: type-check**
  - Run: `cd packages/frontend && npm run type-check`
  - Expected: PASS

- [ ] **Step 4: Commit**
  - `git add packages/frontend/src/types/climate.ts packages/frontend/src/api/climate.ts packages/frontend/docs/HANDOFF.md`
  - `git commit -m "feat(frontend): align climate types with trigger channel"`

### Task 10: 气候配置表单增加“设备/指标/采集通道”四级选择（本地过滤）

**Files:**
- Modify: [views/climate/index.vue](file:///e:/goProject/Hydroponic-Agri-Management/packages/frontend/src/views/climate/index.vue)

- [ ] **Step 1: 增加本地状态**
  - `sensorDevices`（按温室加载）
  - `deviceChannels`（按设备加载全量）
  - 表单字段：
    - `trigger_sensor_device_id`（仅前端态，用于选择与过滤）
    - `trigger_metric_code`
    - `trigger_sensor_channel_id`

- [ ] **Step 2: 联动加载逻辑**
  - 选温室 → 拉取该温室的 sensor devices：`deviceApi.getSensorDevices({ greenhouse_id, page_size: LARGE_PAGE_SIZE })`
  - 选设备 → 拉取该设备全部 sensor channels：`deviceApi.getSensorChannels({ sensor_device_id, page_size: LARGE_PAGE_SIZE, enabled: 1 })`
  - 选指标 → 在本地按 `metric_code` 过滤 channel 下拉选项
  - 选通道 → 写入 `trigger_sensor_channel_id`

- [ ] **Step 3: 提交 create/update/full-create 时携带 trigger_sensor_channel_id**
  - create/update：使用表单字段
  - full-create：JSON 模式需要用户提供 `trigger_sensor_channel_id`（同时在 UI 上给提示文案/校验）

- [ ] **Step 4: 列表展示补充触发来源**
  - 在表格列新增“触发通道”展示：
    - 优先显示 channel_code（通过本地缓存 map：channel_id -> channel_code）
    - 若未加载到，至少显示 `#<id>`

- [ ] **Step 5: 运行前端验证**
  - Run: `cd packages/frontend && npm run dev`
  - Expected: `/strategy/climate` 可创建 profile，必填四项；保存成功；列表可刷新；日志弹窗可正常展示新字段（若后端返回）

- [ ] **Step 6: Commit**
  - `git add packages/frontend/src/views/climate/index.vue`
  - `git commit -m "feat(frontend): select trigger device/metric/channel for climate profile"`

---

## Chunk 4: 文档与交付校验

### Task 11: 更新前端/后端交接文档与项目状态

**Files:**
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/frontend/docs/HANDOFF.md`
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/backend/docs/HANDOFF.md`
- Modify: `e:/goProject/Hydroponic-Agri-Management/packages/frontend/docs/PROJECT_STATUS.md`（如需）

- [ ] **Step 1: 记录变更**
  - 气候联动触发逻辑变化、表单新增字段、通道删除策略

- [ ] **Step 2: Commit**
  - `git add packages/frontend/docs/HANDOFF.md packages/backend/docs/HANDOFF.md packages/frontend/docs/PROJECT_STATUS.md`
  - `git commit -m "docs: handoff climate trigger channel change"`

### Task 12: 最终验证

- [ ] **Step 1: 后端编译**
  - Run: `cd packages/backend && go test ./...`
  - Expected: PASS

- [ ] **Step 2: 前端构建/类型检查**
  - Run: `cd packages/frontend && npm run type-check`
  - Expected: PASS
  - Run: `cd packages/frontend && npm run build`
  - Expected: PASS

- [ ] **Step 3: 关键手工验收用例**
  - 不同温室同 TEMP：只触发绑定通道的 profile
  - 同温室多温度通道：只触发绑定通道的 profile
  - 禁用触发通道：引用 profile 自动 enabled=false；scheduler 不再执行
  - 删除通道：API 返回冲突提示“仅禁用”

