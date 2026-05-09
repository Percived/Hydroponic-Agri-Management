# 外键 ID 列名称化 Implementation Plan
 
> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
 
**Goal:** 前端各业务列表中保留主键 ID 列，但把外键类 `*_id` 列统一改为显示关联实体名称（必要时附带编码），提升可读性与可用性。
 
**Architecture:** 在前端新增“轻量字典/目录”层（Pinia stores + 纯函数 label 生成），统一提供 `id → label` 映射；各页面用模板渲染 label，并保留找不到映射时的回退展示。
 
**Tech Stack:** Vue 3 + TS + Element Plus + Pinia
 
---
 
## Chunk 1: 基础能力（统一 label 映射）
 
### Task 1: 新增通用 label 工具
 
**Files:**
- Create: `packages/frontend/src/utils/labels.ts`
- Modify: `packages/frontend/src/utils/index.ts`（如已有统一导出）
 
- [ ] Step 1: 定义通用工具函数
  - `fallbackIdLabel(prefix: string, id?: number | null): string`
  - `buildIdLabelMap<T>(items: T[], getId: (t: T) => number, getLabel: (t: T) => string): Record<number, string>`
  - 规则：label 为空时回退到 `prefix#id`
 
- [ ] Step 2: 为设备相关 label 提供纯函数（不含 API 调用）
  - `sensorDeviceLabel(d): string` → `name/device_code` 优先，否则单字段，否则 `设备#id`
  - `greenhouseLabel(g): string` → `name/code`（如 code 存在）否则 `温室#id`
  - `growingZoneLabel(z): string` → `name/code` 否则 `种植区#id`
  - `tankLabel(t): string` → `code` 或 `液槽#id`
  - `batchLabel(b): string` → `code` 或 `批次#id`
  - `cropVarietyLabel(v): string` → `name/code` 否则 `作物#id`
  - `growthStageLabel(s): string` → `name/code` 否则 `阶段#id`
 
- [ ] Step 3: 运行 `npm run type-check`
 
### Task 2: 新增目录型 stores（缓存常用实体列表）
 
**Files:**
- Create: `packages/frontend/src/stores/catalog.greenhouse.ts`
- Create: `packages/frontend/src/stores/catalog.growingZone.ts`
- Create: `packages/frontend/src/stores/catalog.nutrientTank.ts`
- Create: `packages/frontend/src/stores/catalog.crop.ts`（cropVarieties + growthStages）
- Create: `packages/frontend/src/stores/catalog.batch.ts`
- Create: `packages/frontend/src/stores/catalog.device.ts`（sensorDevices/sensorChannels/actuatorChannels）
- Modify: `packages/frontend/src/stores/index.ts`
 
- [ ] Step 1: 每个 store 提供统一接口
  - `items`（ref 数组）
  - `loading`
  - `loaded`（是否已成功加载过）
  - `fetchAll()`：用 `page_size: EXTRA_LARGE_PAGE_SIZE` 拉全量（失败吞掉，让页面自己决定是否提示）
  - `labelById`（computed Record<number,string>）
 
- [ ] Step 2: device catalog 额外提供
  - `sensorChannelLabelById`：`channel_code (设备label)`（依赖 sensorDeviceLabelById）
  - `actuatorChannelLabelById`：`channel_code (设备label)`（如能拉 actuator devices；否则先用 `执行器通道#id` 回退）
 
- [ ] Step 3: 运行 `npm run type-check`
 
---
 
## Chunk 2: 页面逐一替换外键列展示
 
> 统一原则：主键 `id` 列保留；外键列标题从“XXID”改为“XX”，单元格显示名称 label，映射不到时显示 `#<id>`。
 
### Task 3: 种植区管理（温室ID → 温室名称）
 
**Files:**
- Modify: `packages/frontend/src/views/greenhouses/zones.vue`
 
- [ ] Step 1: 引入 greenhouse catalog store，确保 onMounted 前/或并行调用 `fetchAll()`
- [ ] Step 2: 表格列 `greenhouse_id` 改为 `label="温室"` + 模板渲染 `greenhouseLabelById[row.greenhouse_id]`
- [ ] Step 3: `npm run type-check`
 
### Task 4: 营养液槽列表（种植区ID → 种植区名称）
 
**Files:**
- Modify: `packages/frontend/src/views/nutrient/tanks.vue`
 
- [ ] Step 1: 引入 growingZone catalog store，加载 zones
- [ ] Step 2: 列 `growing_zone_id` 改为显示名称（保留原字段用于筛选/接口）
- [ ] Step 3: `npm run type-check`
 
### Task 5: 批次台账（温室/种植区/作物外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/batches/ledger.vue`
 
- [ ] Step 1: 引入 greenhouse + growingZone + crop catalog stores，确保数据已加载
- [ ] Step 2: 将 `greenhouse_id / growing_zone_id / crop_variety_id` 列改为名称展示
- [ ] Step 3: `npm run type-check`
 
### Task 6: 病虫害观察（温室/种植区/批次外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/pest/observations.vue`
 
- [ ] Step 1: 引入 greenhouse + growingZone + batch catalog stores
- [ ] Step 2: 将 `greenhouse_id / growing_zone_id / batch_id` 列改为名称展示
- [ ] Step 3: `npm run type-check`
 
### Task 7: 能耗记录（温室/批次外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/energy/records.vue`
 
- [ ] Step 1: 引入 greenhouse + batch catalog stores
- [ ] Step 2: 将 `greenhouse_id / batch_id` 列改为名称展示
- [ ] Step 3: `npm run type-check`
 
### Task 8: 离子检测（液槽外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/nutrient/ion-tests.vue`
 
- [ ] Step 1: 引入 nutrientTank catalog store
- [ ] Step 2: 将 `tank_id` 列改为显示液槽编号/名称
- [ ] Step 3: `npm run type-check`
 
### Task 9: 配方/阶段计划（生长阶段外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/recipes/index.vue`
- Modify: `packages/frontend/src/views/batches/stage-plans.vue`
 
- [ ] Step 1: 引入 crop catalog store（growthStages）
- [ ] Step 2: 将 `growth_stage_id` 的表格/明细展示改为阶段名称
- [ ] Step 3: `npm run type-check`
 
### Task 10: 采收记录（批次外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/batches/harvest.vue`
 
- [ ] Step 1: 引入 batch catalog store
- [ ] Step 2: 将 `batch_id` 列改为批次编号/名称
- [ ] Step 3: `npm run type-check`
 
### Task 11: 气候联动（温室/触发通道/执行通道外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/climate/index.vue`
 
- [ ] Step 1: 引入 greenhouse + device catalog store
- [ ] Step 2: Profiles/Logs 中 `greenhouse_id` 改为温室名称展示
- [ ] Step 3: `trigger_sensor_channel_id` 改为 `sensorChannelLabelById` 展示（含所属设备）
- [ ] Step 4: Actions 中 `actuator_channel_id` 改为 `actuatorChannelLabelById` 展示（含所属设备；缺数据则回退）
- [ ] Step 5: `npm run type-check`
 
### Task 12: 告警列表（通道外键名称化）
 
**Files:**
- Modify: `packages/frontend/src/views/alerts/index.vue`
 
- [ ] Step 1: 引入 device catalog store
- [ ] Step 2: 将列表里 `sensor_channel_id / actuator_channel_id` 的展示替换为 label（含所属设备）
- [ ] Step 3: `npm run type-check`
 
---
 
## Chunk 3: 扫尾与文档
 
### Task 13: 扫描补漏（再次检索 *_id 直出）
 
**Files:**
- Modify: 视扫描结果而定
 
- [ ] Step 1: 在 `packages/frontend/src/views` 中再次检索 `label=\".*ID\"` 与 `row\\..*_id`
- [ ] Step 2: 对遗漏点按同一规则补齐映射与展示
- [ ] Step 3: `npm run type-check`
 
### Task 14: 更新交接文档
 
**Files:**
- Modify: `packages/frontend/docs/HANDOFF.md`
 
- [ ] Step 1: 记录“外键ID列名称化”的改动范围与涉及页面
 
---
 
## 验收清单
 
- [ ] 种植区管理：表格不再显示“温室ID”数字，改为温室名称
- [ ] 营养液槽：列表“种植区”显示名称；新建/编辑通道下拉显示所属设备（已实现的保持一致）
- [ ] 批次台账/病虫害/能耗/离子检测：外键列均为名称展示
- [ ] 气候联动/告警列表：通道类外键显示为 `channel_code (设备)` 形式
- [ ] 映射缺失时仍有可读回退（如 `温室#12`），不出现空白
- [ ] `npm run type-check` 通过
