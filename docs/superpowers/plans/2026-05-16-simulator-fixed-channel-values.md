# Simulator Fixed Channel Values Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为模拟器增加“按传感器采集通道固定上报值”的能力，并让前端页面可设置、查看、清除固定值。

**Architecture:** 在 `simulation` 运行时维护 `sensor_channel_id -> fixed_value` 映射；传感器上报时优先读取固定值，再读取本次手动覆写，最后回退到环境模型值。前端复用现有“通道覆写”表格，补固定值输入与设置/清除按钮，并通过轻量 HTTP 接口与后端同步。

**Tech Stack:** Go 1.24, Gin, 原生 HTML/JS, MQTT, Go test

---

## Chunk 1: Backend Runtime

### Task 1: 固定值运行时与测试

**Files:**
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\server.go`
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\sensor.go`
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\types.go`
- Test: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\server_test.go`

- [ ] **Step 1: Write the failing tests**
  - 在 `server_test.go` 增加：
  - 固定值存在时，`TriggerTelemetry()` 对目标通道始终使用固定值
  - 同一通道同时存在固定值与本次手动覆写时，固定值优先
  - 固定值设置后，`BuildStatusResponse()` 能回显当前固定值

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/simulator -run "TestSimulation_(FixedValueOverridesEnvAndManualOverride|BuildStatusResponseIncludesFixedValues)" -count=1`
Expected: FAIL，提示固定值字段/方法不存在

- [ ] **Step 3: Write minimal implementation**
  - 在 `simulation` 新增固定值表和互斥访问方法
  - 在 `sensor.sendTelemetryWithOverrides()` 增加取值优先级：固定值 > overrides > env
  - 在 `statusResponse` 增加固定值回显字段

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/simulator -run "TestSimulation_(FixedValueOverridesEnvAndManualOverride|BuildStatusResponseIncludesFixedValues)" -count=1`
Expected: PASS

## Chunk 2: HTTP Interface

### Task 2: 固定值设置与清除接口

**Files:**
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\server.go`
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\types.go`
- Test: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\server_test.go`

- [ ] **Step 1: Write the failing tests**
  - 增加接口层测试或最小 handler 测试：
  - `POST /fixed-overrides` 可设置固定值
  - `DELETE /fixed-overrides/:channelId` 可清除固定值

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/simulator -run "TestSimServer_(SetFixedOverride|DeleteFixedOverride)" -count=1`
Expected: FAIL，提示路由或字段缺失

- [ ] **Step 3: Write minimal implementation**
  - 注册两个接口
  - 接口只操作当前运行实例的固定值表
  - 未运行时返回 400

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/simulator -run "TestSimServer_(SetFixedOverride|DeleteFixedOverride)" -count=1`
Expected: PASS

## Chunk 3: Frontend UI

### Task 3: 页面支持设置与清除固定值

**Files:**
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\cmd\simulator\simulator.html`

- [ ] **Step 1: Update the table structure**
  - 在“通道覆写”表格中新增“固定值”列和操作按钮
  - 页面加载 `/status` 后回显固定值

- [ ] **Step 2: Wire actions**
  - 点击“设为固定”时调用 `POST /fixed-overrides`
  - 点击“清除固定”时调用 `DELETE /fixed-overrides/:channelId`
  - 设置成功后刷新状态表

- [ ] **Step 3: Verify manually**

Run:
`go run cmd/simulator/main.go --server --port 3001`

Expected:
  - 页面能为单个传感器通道设置固定值
  - 自动上报与手动上报都使用固定值
  - 清除后恢复环境模型或临时覆写值

## Chunk 4: Verification And Docs

### Task 4: 回归与文档

**Files:**
- Modify: `e:\goProject\Hydroponic-Agri-Management\packages\backend\docs\HANDOFF.md`

- [ ] **Step 1: Run simulator tests**

Run: `go test ./cmd/simulator -count=1`
Expected: PASS

- [ ] **Step 2: Check diagnostics**

Run: 编辑后使用 IDE diagnostics 检查 `server.go`、`sensor.go`、`types.go`、`simulator.html`
Expected: 无新增错误

- [ ] **Step 3: Update docs**
  - 在 `HANDOFF.md` 记录“通道固定值”功能、优先级规则和前端入口
