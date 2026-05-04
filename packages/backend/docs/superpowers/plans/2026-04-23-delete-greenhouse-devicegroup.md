# Delete Greenhouse And DeviceGroup Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为温室与设备组补齐删除接口，采用事务内级联解绑并返回正确的 404 语义。

**Architecture:** 在 `internal/device` 模块新增两个 handler 与路由。每个删除 handler 在单事务内完成存在性检查、关联解绑、目标删除，并用 `RowsAffected` 判定资源不存在。测试在现有 `handler_test.go` 中补充。

**Tech Stack:** Go, Gin, GORM, SQLite(in-memory tests), net/http/httptest

---

### Task 1: Add Failing Tests First

**Files:**
- Modify: `internal/device/handler_test.go`
- Test: `internal/device/handler_test.go`

- [ ] **Step 1: Write failing test for greenhouse delete success + unbind**
- [ ] **Step 2: Run `go test ./internal/device -run TestDeleteGreenhouse` and observe FAIL**
- [ ] **Step 3: Write failing test for greenhouse delete not found**
- [ ] **Step 4: Write failing tests for group delete success + unbind + not found**
- [ ] **Step 5: Run targeted tests and confirm failures reflect missing handlers/routes**

### Task 2: Implement Delete Handlers And Routes

**Files:**
- Modify: `internal/device/handler.go`
- Modify: `internal/device/routes.go`

- [ ] **Step 1: Add `DeleteGreenhouse` handler with transaction + unbind + delete**
- [ ] **Step 2: Add `DeleteGroup` handler with transaction + unbind + delete**
- [ ] **Step 3: Ensure not-found detection uses `RowsAffected` and returns 404**
- [ ] **Step 4: Register DELETE routes for greenhouse and group resources**

### Task 3: Verify And Update Docs

**Files:**
- Modify: `docs/specs/API_SPEC.md`
- Modify: `docs/HANDOFF.md`

- [ ] **Step 1: Run `go test ./internal/device` and `go test ./...`**
- [ ] **Step 2: Update API spec to include two DELETE endpoints**
- [ ] **Step 3: Update handoff notes with changes and verification commands**
