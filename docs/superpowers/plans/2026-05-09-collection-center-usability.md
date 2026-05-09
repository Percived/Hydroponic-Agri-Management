# 采集中心可用性提升 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让采集中心“实时总览/趋势分析”在真实场景下更可用：实时确实连接、过滤确实生效、连接状态可见、断线体验可控、数据质量展示不误导。

**Architecture:** 前端补齐 SSE 生命周期与可视化连接状态，并把筛选条件下沉到 SSE 连接参数；后端在 SSE handler 层做 per-connection 过滤与协议增强（event/id/retry），并修复 SSE 被请求日志无限缓存导致的不稳定问题。

**Tech Stack:** Vue 3 + Element Plus + Axios + EventSource；Go + Gin + SSE（text/event-stream）+ 内部 EventHub。

---

## 范围与约束

- 以“页面实用性/使用体验”为主，不以查询延迟/性能为目标；但会包含影响可用性的后端问题（如 SSE 日志内存增长）。
- 不新增大规模重构；优先小改动闭环。

---

## P0（止血：实时确实能用）

### Task P0-1: 前端 SSE 生命周期闭环（自动连接/断开）

**Files:**
- Modify: `packages/frontend/src/views/telemetry/overview.vue`
- Modify: `packages/frontend/src/composables/useTelemetrySSE.ts`

- [ ] **Step 1: 在 overview 页面挂载时 connect，卸载时 disconnect**
  - 在 `onMounted` 调用 `connect()`
  - 在 `onBeforeUnmount` 调用 `disconnect()`
  - 验收：打开实时总览页面，网络面板看到 `/api/telemetry/subscribe` 长连接；离开页面连接关闭。

- [ ] **Step 2: 暴露并展示连接状态**
  - `useTelemetrySSE` 新增 `status`（`'disconnected'|'connecting'|'connected'|'error'`）与 `lastError`（string）
  - `overview.vue` 顶部加轻量状态提示（例如：绿色“已连接”、黄色“重连中”、红色“连接失败”）
  - 验收：断网/恢复网络时状态能变化；不需要刷新页面才能恢复连接。

- [ ] **Step 3: 手动“重连”入口（用户自救）**
  - 在状态提示旁提供按钮（仅调用 `disconnect(); connect()`）
  - 验收：当服务端重启或网络波动时，用户可点击重连恢复实时数据。

**How to test:**
- 启动前后端，打开 `/collection/overview`
- 断开网络/停止后端 5 秒再启动，观察状态变化与重连是否恢复

---

### Task P0-2: SSE 过滤真正生效（选了什么就只推什么）

**Files:**
- Modify: `packages/frontend/src/composables/useTelemetrySSE.ts`
- Modify: `packages/backend/internal/platform/event/sse_handler.go`
- Modify: `packages/backend/internal/platform/http/middleware.go`（仅做 SSE 相关兼容）

- [ ] **Step 1: 前端把过滤条件写入 SSE URL**
  - `buildURL()` 追加 query：
    - `device_codes=CODE1,CODE2`（可选）
    - `metric_codes=TEMP,PH,...`（可选）
  - 验收：选择设备/指标后重新连接时 URL 带上 query。

- [ ] **Step 2: 后端 SSE handler 读取 query 并过滤事件**
  - 在 `SSEHandler` 内解析 `device_codes` / `metric_codes` 为 set
  - 对每条 `telemetry:received` 事件，读取 `Data.device_code`、`Data.metric_code`，不匹配则跳过写出
  - 验收：同一环境存在多设备时，前端选 A 设备只显示 A 的实时更新。

- [ ] **Step 3: 修复“请求日志对 SSE 的不可用影响”**
  - `RequestLogger` 对 `Content-Type: text/event-stream` 或路径 `/api/*/subscribe`：不要包裹 `ResponseWriter` 去缓存响应体
  - 同时避免记录包含 `token` 的 RawQuery（至少对 `token` 参数脱敏）
  - 验收：长时间打开实时总览，后端内存不因 SSE 输出而持续增长；日志中不出现明文 token。

**How to test:**
- 同时打开 2 个浏览器窗口（不同设备过滤），确认各自只收到自己的设备更新
- 观察后端日志与内存（任务管理器/pprof 可选）

---

### Task P0-3: SSE 协议增强（避免“断线后体验不可控”）

**Files:**
- Modify: `packages/backend/internal/platform/event/sse_handler.go`
- Modify: `packages/backend/internal/platform/event/hub.go`（如需要生成 id）

- [ ] **Step 1: 输出 `event:` 与 `retry:`**
  - 每条消息输出 `event: telemetry`（或使用 `Type` 派生）
  - 输出 `retry: 2000`（或配置项）
  - 验收：浏览器 EventSource 在服务端短暂断开后会按 retry 自动重连（可在浏览器网络面板观察）。

- [ ] **Step 2: 输出 `id:`（最小实现即可）**
  - 生成方式二选一：
    1) 使用 `collected_at` 作为 id（RFC3339 或 unix ms）
    2) 使用 hub 内递增序号
  - 验收：SSE 消息带 `id:` 字段；浏览器自动携带 `Last-Event-ID`（不强制后端实现补发，仅保证协议完备）。

---

## P1（增强：不误导、更稳定、更一致）

### Task P1-1: 数据质量标识一致性（不再强行 normal）

**Files:**
- Modify: `packages/backend/internal/platform/mqtt/ingress.go`（publish telemetry:received data）
- Modify: `packages/frontend/src/views/telemetry/overview.vue`
- Modify: `packages/frontend/src/types/telemetry.ts`
- Modify: `packages/frontend/src/views/telemetry/trends.vue`（筛选项/展示）

- [ ] **Step 1: SSE payload 增加 `quality_flag`**
  - 发布事件时在 `Data` 里加入 `quality_flag`
  - 验收：前端收到的 SSE event 中带 `quality_flag`

- [ ] **Step 2: 前端不再覆盖 quality_flag**
  - `overview.vue` 更新卡片时使用 SSE 的 `quality_flag`；若缺失则显示“未知”或沿用最近一次值
  - 验收：当后端发出非 normal 时，卡片上能看到对应标识

- [ ] **Step 3: 统一质量枚举与筛选项**
  - 以共享文档 `TelemetryRecord.quality_flag` 为准（如 `normal/missing/out_of_range/device_offline`）
  - 修正前端 type 与筛选 option，避免筛选传参不生效/展示不一致
  - 验收：趋势分析按质量过滤时，筛选项与返回数据一致（不出现“永远筛不到”的情况）。

---

### Task P1-2: 竞态与错误可见（快速切换不“跳回旧数据”）

**Files:**
- Modify: `packages/frontend/src/views/telemetry/overview.vue`
- Modify: `packages/frontend/src/views/telemetry/trends.vue`
- (Optional) Modify: `packages/frontend/src/api/request.ts`

- [ ] **Step 1: 为 overview 的级联加载增加竞态保护**
  - 引入 `requestSeq`（递增序号）或 `AbortController`（如 request 层支持）
  - 每次切换温室/种植区/设备时，只允许“最后一次请求”落地写入 state
  - 验收：快速切换筛选，页面不会回闪到旧温室/旧设备的数据。

- [ ] **Step 2: 减少静默吞错，给出用户可理解的空态**
  - 对关键请求失败展示轻提示（例如顶部 `ElMessage` 或空态文本“加载失败，请重试”）
  - 验收：后端停止时，页面不再只是空白/永远 loading，而是明确提示并可重试。

---

### Task P1-3: SSE 过滤参数与文档对齐

**Files:**
- Modify: `shared/docs/API_SPEC.md`
- Modify: `shared/docs/openapi.yaml`
- Modify: `packages/backend/docs/HANDOFF.md`

- [ ] **Step 1: 文档补齐 telemetry subscribe 的 query 参数**
  - 增加 `/api/telemetry/subscribe` 的说明：支持 `device_codes`、`metric_codes`（逗号分隔）
  - 验收：文档与实现一致，后续开发能直接复用。

---

## 验证清单（完成后必须能做到）

- [ ] 打开 `/collection/overview` 能看到“已连接/重连中/连接失败”状态
- [ ] 选择设备/指标后，实时更新只来自选中的设备/指标
- [ ] 后端长时间保持 SSE 连接不会出现明显内存持续增长（至少不再因为日志缓存响应体而增长）
- [ ] `quality_flag` 在实时卡片与趋势分析筛选项中一致且不误导
- [ ] 快速切换筛选不会出现旧请求覆盖新状态

---

## 运行与检查命令（参考）

**Frontend**
- `npm run -s type-check`（必要时设置 `NODE_OPTIONS=--max-old-space-size=4096`）
- `npm run -s build`

**Backend**
- `go test ./...`
- `go run cmd/api/main.go`

