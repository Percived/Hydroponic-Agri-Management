# 系统审计与优化 PRD（v2.1.0）

> 基于 v2.0.0 重构后系统的客观审查，按"假设不合理 → 寻找证据 → 优化建议"框架编写。
> 更新日期：2026-05-06

---

## 一、架构层面

### 1.1 数据完整性：全部使用逻辑外键，零物理外键约束

**假设**：迁移文件中所有表只建索引，不建 `FOREIGN KEY`，靠应用层保证引用完整性是合理的。

**不合理证据**：
- `crop_batches.greenhouse_id`、`sensor_channels.sensor_device_id` 等字段在数据库层没有任何约束，脏数据无法追溯
- 删除温室时，关联的设备、批次、策略不会级联处理，产生孤儿记录
- 所有查询依赖 JOIN，但没有外键保证 JOIN 一定能匹配到有效行
- 丢弃了 MySQL 的 `ON DELETE CASCADE / SET NULL` 能力，改为在 Go 代码里手写检查逻辑，既低效又容易遗漏

**优化建议**：
| 层级 | 策略 |
|------|------|
| 核心关联 | **添加物理外键**：`device ↔ channel`、`greenhouse ↔ zone`、`batch ↔ greenhouse`、`batch ↔ crop_variety` |
| 跨域引用 | 保持逻辑外键 + 定期数据清理任务 |
| 删除策略 | 统一使用软删除（`deleted_at`），保留审计追溯能力，避免级联删除丢失数据 |

---

### 1.2 MQTT 控制链路虚设

**假设**：`command` 模块的 `SendCommand` 和 `AckCommand` 更新数据库状态即可。

**不合理证据**：
- `command/handler.go` 中注入了 `mqtt.Client` 依赖但**从未使用**
- 用户在前端点击"发送指令"，数据库标记为 `SENT`，但设备永远不会收到
- 如果这是预留接口，前端按钮不应该暴露给用户

**优化建议**：
1. `SendCommand` → 构建 MQTT 消息 → publish 到 `actuator/{device_code}/{channel_code}/command`
2. 发送后启动 goroutine + `context.WithTimeout`，超时自动标记 `TIMEOUT`
3. `AckCommand` → 订阅 `actuator/+/+/ack` 主题 → 匹配 `command_id` → 写入回执
4. 若暂不实现，前端命令发送按钮置灰并标注 "（未启用）"

---

### 1.3 策略引擎只有手动触发，缺乏自动化

**假设**：策略 CRUD + 手动执行就能满足自动化需求。

**不合理证据**：
- 策略的 `policy_type` 包含 `SCHEDULE`（定时）和 `THRESHOLD`（阈值触发），但两者都依赖人工调用 `POST /api/policy-executions`
- 温度超出阈值时，需要有人盯着仪表盘手动点"执行策略"
- 策略模块是系统价值最高的部分（自动化控制），当前形态只是一个"配置记录器"

**优化建议**：
| 触发方式 | 实现 |
|----------|------|
| 定时扫描 | 每 N 秒拉取 `enabled=true` 的策略，评估条件，自动执行 |
| 事件驱动 | telemetry ingest 流程中检查新数据是否触发阈值策略 |
| 冲突处理 | 评估结果写入 `policy_executions`，包括跳过原因（`COOLDOWN`、`CONFLICT` 等） |

---

### 1.4 InfluxDB 集成在 HTTP 层不可见

**假设**：遥测数据存入 MySQL 即可，InfluxDB 留给 MQTT consumer 处理。

**不合理证据**：
- InfluxDB 客户端已注入 DI 容器，但 `telemetry/handler.go` 只用 MySQL
- 查询接口走 `telemetry_records` 表，InfluxDB 形同虚设
- 数据流不清晰：是 MySQL + InfluxDB 双写？还是 MySQL 为主 InfluxDB 为辅？

**优化建议**：
- 明确数据流：`MQTT → Backend Consumer → InfluxDB（主存储）+ MySQL（元数据）`
- `GET /telemetry/query` 和 `GET /telemetry/channels/:id/history` 从 InfluxDB 读取
- 在文档中绘制数据流向图

---

## 二、代码质量

### 2.1 辅助函数重复定义

**假设**：每个 handler 文件自己定义 `parseID()`、`parsePage()`、`currentUserID()` 是合理的。

**不合理证据**：
- 同一段逻辑在 `alert/`、`climate/`、`command/`、`device/`、`policy/` 等 **至少 10 个文件**中各自定义
- 如果需要统一调整异常信息格式或默认分页大小，需要修改 10+ 处

**优化建议**：
- 在 `platform/http/helpers.go` 中提取公共函数
- 默认分页参数（`page=1, page_size=20`）写入常量

```go
// platform/http/helpers.go
func ParseID(c *gin.Context, key string) (uint64, error)
func ParsePage(c *gin.Context) (page int, pageSize int)
func CurrentUserID(c *gin.Context) uint64
```

---

### 2.2 序列化风格不一致

**假设**：不同模块使用不同的响应构建方式是灵活性的体现。

**不合理证据**：

| 模块 | 方式 | 时间字段类型 |
|------|------|-------------|
| `device` | 强类型 DTO | `string` |
| `climate` | 混合 DTO + `gin.H{}` | `time.Time` |
| `overview` | 全部 `gin.H{}` | 无类型 |

后果：同一 API 文档中，`created_at` 字段在不同端点可能是 ISO 8601 字符串或 Unix 时间戳，前端需做防御性解析。

**优化建议**：
1. 统一使用**强类型响应 DTO**，禁止在 handler 中使用 `gin.H{}`
2. 时间字段统一为 `string`，在 DTO 构造时显式格式化
3. 通过 code review 强制执行

---

### 2.3 部分模块缺失 RowAffected 检查

**假设**：GORM 的 `Updates`/`Delete` 不检查 `RowsAffected` 也能接受。

**不合理证据**：
- 以下模块的更新/删除 handler **未检查** `RowsAffected`：`crop`、`energy`、`nutrient`、`pest`、`recipe`、`review`
- 操作不存在的资源时返回 200 而非 404，前端无法区分"成功修改"和"资源不存在"
- 已实现的模块：`device`、`alert`、`climate`、`command`、`policy`

**优化建议**：
1. **立即修复**所有缺失 `RowsAffected` 检查的 handler
2. 在 `platform/http/` 中封装公共函数：

```go
func EnsureOneRowAffected(result *gorm.DB, resource string) error {
    if result.RowsAffected == 0 {
        return ErrNotFound(resource)
    }
    return nil
}
```

---

### 2.4 单文件 handler 过于庞大

**假设**：一个模块的所有 handler 放在一个文件里不影响可维护性。

**不合理证据**：
- `climate/handler.go`：**1018 行**，包含 Profile / Stage / Action / ExecutionLog 四类资源
- `nutrient/handler.go`、`policy/handler.go`、`crop/handler.go` 同样存在此问题

**优化建议**：
- 拆分为按资源的文件：`profile_handler.go`、`stage_handler.go`、`action_handler.go`、`execution_handler.go`
- 目标：单文件不超过 400 行

---

### 2.5 模型与 DTO 之间缺乏转换层

**假设**：在 handler 中直接做 model → DTO 字段拷贝是合理的。

**不合理证据**：
- 敏感字段（如 `password_hash`）可能意外暴露
- 计算字段（如"设备是否在线"需要判断 `last_seen_at > now - 5min`）与业务逻辑混在一起
- 转换逻辑散落各处，无法单独测试

**优化建议**：
- 每个模块添加 `ToResponse()` 方法或 `mapper.go`，统一 model → DTO 转换
- 禁止在 handler 中直接返回 GORM model

---

## 三、前端问题

### 3.1 SSE 实时推送形同虚设

**假设**：实现了 `useAlertSSE` 和 `useTelemetrySSE` 就能工作。

**不合理证据**：
- `AppHeader.vue` 中 `alertBadgeCount` 初始为 0，**没有任何 watch 或事件处理器更新它**
- 告警数永远是 0，除非手动刷新页面
- `useAlertSSE` 提供的 `lastAlert` 和 `alertCount` 被解构出来但**从未使用**

**优化建议**：
1. 在 `AppHeader.vue` 中 watch `alertCount` 变化，同步更新 `alertBadgeCount`
2. 为 SSE 不可用场景添加轮询降级方案
3. 在 Dashboard 页面消费 `useTelemetrySSE` 实现实时曲线更新

---

### 3.2 死代码类型定义

**假设**：`types/domain.ts` 中定义的枚举在项目中是有用的。

**不合理证据**：
- `AssetOnlineStatus`、`AssetChannelType`、`AssetAlertLevel`、`AssetAlertStatus` 在代码库中**零引用**
- 实际类型在各自模块文件中独立定义（`alert.ts`、`device.ts`）
- 枚举值与实际 API 返回不一致

**优化建议**：删除 `domain.ts` 中的死代码，或将其迁移为各模块类型文件的 source of truth。

---

### 3.3 权限检查变量冗余

**假设**：`canOperate` 和 `canControl` 是不同的权限语义。

**不合理证据**：
```typescript
const canOperate = hasRole('ADMIN') || hasRole('OPERATOR')
const canControl = canControlDevice() // 内部: hasRole(ADMIN) || hasRole(OPERATOR)
```
两者值完全相同，但命名不同，造成混淆。

**优化建议**：合并为一个变量 `canOperate`，或按实际业务拆分为有差异的权限粒度。

---

### 3.4 分页响应类型不一致

**假设**：`PaginatedResponse<T>` 和 `PaginatedData<T>` 并存是可以的。

**不合理证据**：
- 两者在 `api.ts` 中定义，结构完全相同（`{page, page_size, total, items[]}`）
- `recipe.ts` 使用 `PaginatedData`，其余模块使用 `PaginatedResponse`

**优化建议**：删除 `PaginatedData`，全部统一为 `PaginatedResponse<T>`。

---

## 四、数据模型设计

### 4.1 `enabled` 字段类型语义不清

**假设**：用 `uint8` 表示布尔值是数据库惯例。

**不合理证据**：
- Go 有 `bool`，MySQL 有 `TINYINT(1)`，用 `uint8` 增加心智负担
- 如果未来需要扩展（`0=禁用, 1=启用, 2=仅告警`），应该用枚举字符串而非 magic number
- 受影响模型：`ClimateProfile.Enabled`、`PolicyCondition.Enabled` 等

**优化建议**：
- 统一改为 `bool` + `gorm:"default:true"`
- 如需多状态，使用 `varchar` 枚举字符串（与项目风格一致）

---

### 4.2 JSON 字段标签不统一

**假设**：不同模型对 JSON 列的 GORM 标签差异不重要。

**不合理证据**：

| 模型 | 标签 |
|------|------|
| `AlertTimelineEvent.EventPayload` | `gorm:"type:json"` |
| `PolicyTarget.CommandPayload` | `gorm:"type:json;not null"` |
| `NotificationChannel.Config` | 无显式标签（依赖 `json.RawMessage` 推断） |

**优化建议**：统一为 `gorm:"type:json"` + 业务层判断空值。

---

## 五、安全性

### 5.1 JWT Secret 默认值未强制修改

**假设**：文档中说明了需要修改，用户会自觉修改。

**不合理证据**：
- `config.yaml` 中 `auth.jwt_secret: "change-me"`
- `influx.token: "your-token"`
- 如果部署后忘记修改，系统形同裸奔

**优化建议**：
- 后端启动时检查默认值，拒绝启动并打印警告
- 高危配置项强制从环境变量读取，禁止使用默认值

---

## 六、迁移文件管理

### 6.1 迁移文件混乱

**假设**：原始编号文件删除后，合并文件能独立承担迁移职责。

**不合理证据**：
- 原始 `migrations/0001_*.sql` ~ `0007_*.sql` 已被删除（git status 显示 `D`）
- HANDOFF.md 声称"原始编号迁移文件保留不变"，与实际不符
- `migrations/merged/all.up.sql` 成为唯一的迁移入口，失去了增量迁移的优势

**优化建议**：
1. 恢复原始编号文件作为增量迁移的历史记录
2. `all.up.sql` 保留为"全量初始化"的便捷入口
3. 或明确声明项目只使用合并迁移，更新 HANDOFF.md 说明

---

## 七、优先级排序

| 优先级 | 问题 | 影响 | 修复成本 |
|--------|------|------|----------|
| **P0** | 缺失 RowAffected 检查 | 操作不存在资源返回 200 | 低 |
| **P0** | JWT secret 默认值风险 | 安全漏洞 | 低 |
| **P1** | MQTT 控制链路缺失 | 核心功能不可用 | 中 |
| **P1** | 策略引擎无自动化 | 自动化控制形同虚设 | 高 |
| **P1** | SSE 告警徽章不工作 | 用户看到假数据 | 低 |
| **P1** | 重复辅助函数 | 维护成本高 | 低 |
| **P2** | 无物理外键 | 数据完整性风险 | 中（需评估兼容性） |
| **P2** | 序列化风格不统一 | 维护成本 + 前端解析负担 | 中 |
| **P2** | 审计日志非自动 | 合规风险 | 中 |
| **P2** | 迁移文件管理混乱 | 开发者困惑 | 低 |
| **P3** | 死代码（domain.ts） | 代码可读性 | 低 |
| **P3** | `enabled` 字段类型 | 一致性 | 低 |
| **P3** | JSON 标签不统一 | 一致性 | 低 |
| **P3** | 分页类型冗余 | 轻微混淆 | 低 |

---

## 八、执行建议

### 第一阶段（1-2 天）
- P0 全部修复
- P1 中的重复辅助函数、SSE 徽章修复

### 第二阶段（3-5 天）
- P1 中的 MQTT 控制链路
- P2 中的序列化统一、RowAffected 封装、审计中间件

### 第三阶段（1-2 周）
- P1 中的策略引擎自动化
- P2 中的物理外键评估与实施
