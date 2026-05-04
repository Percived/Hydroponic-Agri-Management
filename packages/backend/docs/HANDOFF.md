# 交接文档

最后更新: 2026-05-04
当前分支: main
当前重点: API 响应字段补齐 + 控制规则种子数据 + 文档同步更新。

## 1. 近期变更（2026-05-04 会话）

### API 响应字段补齐（Bug Fix）
- **控制规则列表** `GET /api/controls/rules`：
  - 新增 LEFT JOIN devices 填充 `target_device_name`
  - 新增返回 `action`（JSON，含 command_type + payload）、`updated_at`
- **告警列表/详情** `GET /api/alerts`、`GET /api/alerts/:id`：
  - 新增 LEFT JOIN devices 填充 `device_name`
  - 新增返回 `created_at`
  - Alert 模型新增 `DeviceName`（gorm:"->"）、`CreatedAt`、`UpdatedAt`
- **告警统计** `GET /api/alerts/stats`：
  - 返回 key 从 `open/ack/closed` 修正为 `open_count/ack_count/closed_count`（对齐前端 AlertStats 接口）

### 前端类型与 API 层修复
- `types/control.ts` — ControlRule 新增 `action` 可选字段
- `types/alert.ts` — Alert 新增 `triggered_at`、`resolved_at` 字段
- `api/control.ts` — `getRules` 从 `action` JSON 中提取 `command_type` + `command_payload`

### 控制规则种子数据（新）
- `migrations/0006_seed_control_rules.up.sql` — 22 条自动控制规则
  - 覆盖 TEMP/HUMIDITY/PH/EC/CO2/LIGHT 六大指标
  - 1号温室（叶菜）12 条 + 2号温室（草莓）10 条
  - 每对开关规则含回差区间，避免阈值波动反复触发

### 文档更新
- `shared/docs/API_SPEC.md` — ControlRule/Alert 数据模型 & 响应示例同步更新

---

## 2. 历史变更

### Phase 2
- 设备模块：`GET /api/devices/:deviceId/telemetry-summary`、`POST /api/devices/batch-update`、`DELETE /api/devices/batch`
- 控制模块：`POST /api/controls/batch-commands`
- 通知模块：完整 CRUD + Webhook 测试（HMAC-SHA256）
- 遥测模块：`GET/PUT /api/telemetry/system-configs`
- 基础设施：`migrations/0004_notification_channels.up.sql`、`migrations/0005_seed_devices.up.sql`

### 更早变更
- 在设备模块中新增删除 API：
  - `DELETE /api/devices/greenhouses/:greenhouseId`
  - `DELETE /api/device-groups/:groupId`
- 设备模块 handler 测试、温室/分组级联解绑行为
- Swagger UI、OpenAPI 规范、API 演示页面等

---

## 3. 待办事项（前 5 项）

1. 为通知模块 dispatch 逻辑补齐——目前 `evaluateAndTrigger` 创建告警后未实际调用 `go dispatchNotifications(alert)`。
2. 继续为剩余的更新/删除 handler 添加 `RowsAffected` 检查，对不存在的资源返回 404。
3. 为 auth/device/telemetry/control 关键路径创建首批自动化测试。
4. 确定并实现控制模板应用 + 告警订阅 SSE 流式传输的策略。
5. 对齐并验证 MQTT/Influx 在开发/生产环境中的启动行为。

## 4. 阻碍 / 风险

- 自动化测试覆盖仍然不足（仅 device 模块有少量测试）。
- 部分行为级 API 缺口目前被成功响应所掩盖（模板应用、告警订阅）。
- 通知模块异步 dispatch 尚未与告警引擎集成。

## 5. 验证说明

- 如果本地 shell 中 Go 工具链不可用，则无法从终端验证编译/测试状态。
- 使用 `scripts/status-snapshot.ps1` 可快速刷新结构快照。

## 6. 下个会话如何继续

1. 阅读 `docs/PROJECT_STATUS.md`。
2. 阅读本文件（`docs/HANDOFF.md`）。
3. 从待办事项 #1 开始，除非优先级发生变化。

## 7. 快速填写模板（每次交接使用）

- 日期：2026-05-04
- 分支：main
- 已完成范围：API 响应字段补齐（控制规则/告警列表 device_name、action、stats keys）+ 前端类型同步 + 22 条控制规则种子数据 + 文档更新
- 待完成范围：通知异步 dispatch 集成、自动化测试、模板应用/告警订阅落地
- 风险：通知 dispatch 与告警引擎未集成；自动化测试覆盖低
- 下个首要命令：`go build ./...` / `go test ./...` / `npm run build`
